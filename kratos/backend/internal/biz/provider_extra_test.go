package biz

// T-P-19/20 新渠道验签单测：Razorpay/KOMOJU HMAC-SHA256 向量、PortOne/Xendit token 比对、
// Mercado Pago IPN payment_id 解析 + 回查模拟、PayPal 官方 verify-webhook-signature 验签 mock。

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// hmac256Headers 本地构造 HMAC-SHA256 十六进制签名头（Razorpay/KOMOJU 测试向量）。
func hmac256Headers(name, secret string, body []byte) map[string]string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return map[string]string{name: hex.EncodeToString(mac.Sum(nil))}
}

func TestRazorpayVerifyWebhook(t *testing.T) {
	p := &RazorpayProvider{webhookSec: "rzp-test-secret"}
	ctx := context.Background()
	body := []byte(`{"event":"payment_link.paid","payload":{"payment_link":{"entity":{"id":"plink_1","status":"paid","amount":300,"currency":"INR","notes":{"order_no":"2026082999"}}}}}`)
	ev, err := p.VerifyWebhook(ctx, "razorpay", body, hmac256Headers("X-Razorpay-Signature", "rzp-test-secret", body))
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082999" || ev.AmountCents != 300 || ev.Currency != "INR" || ev.TxID != "plink_1" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// 篡改 body → 验签失败
	if _, err := p.VerifyWebhook(ctx, "razorpay", append(body, ' '),
		hmac256Headers("X-Razorpay-Signature", "rzp-test-secret", body)); err == nil {
		t.Fatal("tampered payload must fail")
	}
	// 非支付事件 → Paid=false 不报错
	pend := []byte(`{"event":"payment_link.created","payload":{"payment_link":{"entity":{"id":"plink_2"}}}}`)
	ev, err = p.VerifyWebhook(ctx, "razorpay", pend, hmac256Headers("X-Razorpay-Signature", "rzp-test-secret", pend))
	if err != nil || ev.Paid {
		t.Fatalf("non-paid event: err=%v ev=%+v", err, ev)
	}
}

func TestKomojuVerifyWebhook(t *testing.T) {
	p := &KomojuProvider{webhookSec: "kmj-test-secret"}
	ctx := context.Background()
	body := []byte(`{"type":"payment.paid","data":{"id":"pm_1","status":"paid","currency":"JPY","external_order_no":"2026082998","amounts":{"cents":3000,"units":3000}}}`)
	ev, err := p.VerifyWebhook(ctx, "komoju", body, hmac256Headers("X-KOMOJU-SIGNATURE", "kmj-test-secret", body))
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082998" || ev.AmountCents != 3000 || ev.Currency != "JPY" || ev.TxID != "pm_1" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// 篡改 body → 验签失败
	if _, err := p.VerifyWebhook(ctx, "komoju", append(body, ' '),
		hmac256Headers("X-KOMOJU-SIGNATURE", "kmj-test-secret", body)); err == nil {
		t.Fatal("tampered payload must fail")
	}
}

func TestPortOneVerifyWebhook(t *testing.T) {
	p := &PortOneProvider{webhookToken: "iamport-test-token"}
	ctx := context.Background()
	body := []byte(`{"type":"Payment.Paid","data":{"paymentId":"imp_1","merchantUid":"2026082997","status":"PAID","amount":{"total":800,"currency":"KRW"}}}`)
	ev, err := p.VerifyWebhook(ctx, "portone", body, map[string]string{"X-IAMPORT-TOKEN": "iamport-test-token"})
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082997" || ev.AmountCents != 800 || ev.Currency != "KRW" || ev.TxID != "imp_1" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// token 不匹配 → 拒绝
	if _, err := p.VerifyWebhook(ctx, "portone", body, map[string]string{"X-IAMPORT-TOKEN": "wrong"}); err == nil {
		t.Fatal("wrong token must fail")
	}
}

func TestXenditVerifyWebhook(t *testing.T) {
	p := &XenditProvider{callbackToken: "xnd-callback-token"}
	ctx := context.Background()
	body := []byte(`{"id":"inv_1","external_id":"2026082996","status":"PAID","amount":300,"paid_amount":300,"currency":"IDR"}`)
	ev, err := p.VerifyWebhook(ctx, "xendit", body, map[string]string{"X-CALLBACK-TOKEN": "xnd-callback-token"})
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082996" || ev.AmountCents != 300 || ev.Currency != "IDR" || ev.TxID != "inv_1" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// token 不匹配 → 拒绝
	if _, err := p.VerifyWebhook(ctx, "xendit", body, map[string]string{"X-CALLBACK-TOKEN": "wrong"}); err == nil {
		t.Fatal("wrong token must fail")
	}
	// 非 PAID → Paid=false
	pend := []byte(`{"id":"inv_2","external_id":"2026082995","status":"PENDING","amount":300,"currency":"IDR"}`)
	ev, err = p.VerifyWebhook(ctx, "xendit", pend, map[string]string{"X-CALLBACK-TOKEN": "xnd-callback-token"})
	if err != nil || ev.Paid {
		t.Fatalf("pending: err=%v ev=%+v", err, ev)
	}
}

// fakeRoundTripper 固定响应（Mercado Pago 回查模拟）。
type fakeRoundTripper struct {
	body   string
	status int
}

func (f fakeRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: f.status,
		Body: io.NopCloser(strings.NewReader(f.body)), Header: http.Header{}}, nil
}

// routeRoundTripper 按 URL path 分发的假 transport（PayPal oauth + verify 两个端点用）。
type routeRoundTripper struct {
	routes map[string]fakeRoundTripper
}

func (r routeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt, ok := r.routes[req.URL.Path]; ok {
		return rt.RoundTrip(req)
	}
	return &http.Response{StatusCode: 500,
		Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
}

func TestMercadoPagoVerifyWebhook(t *testing.T) {
	p := &MercadoPagoProvider{accessToken: "TEST-access-token",
		http: &http.Client{Transport: fakeRoundTripper{status: 200,
			body: `{"id":123,"status":"approved","external_reference":"2026082994","currency_id":"BRL","transaction_amount":29.99}`}}}
	ev, err := p.VerifyWebhook(context.Background(), "mercadopago", []byte(`{"id":123,"type":"payment"}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082994" || ev.AmountCents != 2999 || ev.Currency != "BRL" || ev.TxID != "123" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// 未支付状态 → Paid=false
	p.http.Transport = fakeRoundTripper{status: 200,
		body: `{"id":124,"status":"pending","external_reference":"2026082993","currency_id":"BRL","transaction_amount":29.99}`}
	ev, err = p.VerifyWebhook(context.Background(), "mercadopago", []byte(`{"id":124,"type":"payment"}`), nil)
	if err != nil || ev.Paid {
		t.Fatalf("pending: err=%v ev=%+v", err, ev)
	}
}

func TestPayPalVerifyWebhook(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"event_type":"CHECKOUT.ORDER.APPROVED","resource":{"id":"ORDER-1","status":"APPROVED","purchase_units":[{"reference_id":"2026082992","amount":{"currency_code":"USD","value":"3.00"}}]}}`)
	hdrs := map[string]string{
		"PayPal-Auth-Algo": "SHA256withRSA", "PayPal-Cert-Url": "https://api.paypal.com/cert",
		"PayPal-Transmission-Id": "txn-1", "PayPal-Transmission-Time": "2026-08-29T00:00:00Z",
		"PayPal-Transmission-Sig": "sig-1", "PayPal-Webhook-Id": "WH-1",
	}
	rt := routeRoundTripper{routes: map[string]fakeRoundTripper{
		"/v1/oauth2/token": {status: 200, body: `{"access_token":"tok-1"}`},
		"/v1/notifications/verify-webhook-signature": {status: 200, body: `{"verification_status":"SUCCESS"}`},
	}}
	p := &PayPalProvider{webhookID: "WH-1", clientID: "cid", clientSecret: "cs",
		http: &http.Client{Transport: rt}}

	// 官方验签 SUCCESS → 通过，金额/订单解析保留
	ev, err := p.VerifyWebhook(ctx, "paypal", body, hdrs)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082992" || ev.AmountCents != 300 || ev.Currency != "USD" || ev.TxID != "ORDER-1" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	// 验签返回 FAILURE → 拒绝
	rt.routes["/v1/notifications/verify-webhook-signature"] =
		fakeRoundTripper{status: 200, body: `{"verification_status":"FAILURE"}`}
	if _, err := p.VerifyWebhook(ctx, "paypal", body, hdrs); err == nil {
		t.Fatal("verify FAILURE must be rejected")
	}

	// 验签端点 HTTP 错误 → 拒绝
	rt.routes["/v1/notifications/verify-webhook-signature"] =
		fakeRoundTripper{status: 500, body: `internal error`}
	if _, err := p.VerifyWebhook(ctx, "paypal", body, hdrs); err == nil {
		t.Fatal("verify http error must be rejected")
	}

	// 恢复 SUCCESS；webhook id 不匹配 → 拒绝（本地快速失败，不打 PayPal）
	rt.routes["/v1/notifications/verify-webhook-signature"] =
		fakeRoundTripper{status: 200, body: `{"verification_status":"SUCCESS"}`}
	bad := map[string]string{"PayPal-Auth-Algo": "SHA256withRSA", "PayPal-Cert-Url": "https://api.paypal.com/cert",
		"PayPal-Transmission-Id": "txn-1", "PayPal-Transmission-Time": "t",
		"PayPal-Transmission-Sig": "s", "PayPal-Webhook-Id": "WH-OTHER"}
	if _, err := p.VerifyWebhook(ctx, "paypal", body, bad); err == nil {
		t.Fatal("wrong webhook id must fail")
	}
	// transmission 头缺失 → 拒绝
	if _, err := p.VerifyWebhook(ctx, "paypal", body, map[string]string{"PayPal-Webhook-Id": "WH-1"}); err == nil {
		t.Fatal("missing transmission headers must fail")
	}
}

// ---- Alipay（RSA2 签名/验签 + 表单 notify 解析，自造密钥对）----

// testAlipayKeys 生成 RSA-2048 密钥对并编码为 PKCS8 私钥 / PKIX 公钥 PEM。
func testAlipayKeys(t *testing.T) (privPEM, pubPEM string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	return privPEM, pubPEM
}

// alipayNotifyBody 构造已签名的 notify 表单体（模拟支付宝回调）。
func alipayNotifyBody(t *testing.T, privPEM string, fields map[string]string) []byte {
	t.Helper()
	priv, err := parseRSAPrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	sign, err := alipaySign(fields, priv)
	if err != nil {
		t.Fatal(err)
	}
	fields["sign"] = sign
	fields["sign_type"] = "RSA2"
	return []byte(urlValues(fields).Encode())
}

func TestAlipaySignVerify(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	priv, _ := parseRSAPrivateKey(privPEM)
	pub, _ := parseRSAPublicKey(pubPEM)
	params := map[string]string{
		"app_id": "2021000000000001", "method": "alipay.trade.wap.pay", "charset": "utf-8",
		"sign_type": "RSA2", "timestamp": "2026-08-29 12:00:00", "version": "1.0",
		"notify_url": "https://example.com/webhook/alipay",
		"biz_content": `{"out_trade_no":"2026082999","total_amount":"3.00","product_code":"QUICK_WAP_WAY","subject":"Open Novel VIP monthly"}`,
	}
	// 签名 → 验签通过
	params["sign"] = mustSign(t, params, priv)
	if err := alipayVerify(copyParams(params), pub); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	// 篡改任一参数 → 拒绝
	tampered := copyParams(params)
	tampered["total_amount"] = "300.00"
	// 重新构造 biz_content 使篡改进入签名串
	tampered["biz_content"] = `{"out_trade_no":"2026082999","total_amount":"300.00","product_code":"QUICK_WAP_WAY","subject":"Open Novel VIP monthly"}`
	if err := alipayVerify(tampered, pub); err == nil {
		t.Fatal("tampered params must fail")
	}
	// sign 错误 → 拒绝
	wrong := copyParams(params)
	wrong["sign"] = "AAAA"
	if err := alipayVerify(wrong, pub); err == nil {
		t.Fatal("wrong sign must fail")
	}
}

func TestAlipayVerifyWebhook(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	p := &AlipayProvider{pubKey: pubPEM}
	ctx := context.Background()
	base := map[string]string{
		"app_id": "2021000000000001", "charset": "utf-8", "timestamp": "2026-08-29 12:00:00",
		"version": "1.0", "trade_status": "TRADE_SUCCESS", "out_trade_no": "2026082999",
		"trade_no": "2026082922001400000001", "total_amount": "9.90",
	}
	ev, err := p.VerifyWebhook(ctx, "alipay", alipayNotifyBody(t, privPEM, copyParams(base)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082999" || ev.TxID != "2026082922001400000001" ||
		ev.AmountCents != 990 || ev.Currency != "CNY" {
		t.Fatalf("unexpected event: %+v", ev)
	}
	// TRADE_FINISHED 同样算支付成功
	fin := copyParams(base)
	fin["trade_status"] = "TRADE_FINISHED"
	ev, err = p.VerifyWebhook(ctx, "alipay", alipayNotifyBody(t, privPEM, fin), nil)
	if err != nil || !ev.Paid {
		t.Fatalf("TRADE_FINISHED: err=%v ev=%+v", err, ev)
	}
	// WAIT_BUYER_PAY → Paid=false
	wait := copyParams(base)
	wait["trade_status"] = "TRADE_WAIT_BUYER_PAY"
	ev, err = p.VerifyWebhook(ctx, "alipay", alipayNotifyBody(t, privPEM, wait), nil)
	if err != nil || ev.Paid {
		t.Fatalf("pending: err=%v ev=%+v", err, ev)
	}
	// 篡改表单（金额被换）→ 验签拒绝
	tampered := alipayNotifyBody(t, privPEM, copyParams(base))
	tampered = []byte(strings.Replace(string(tampered), "total_amount=9.90", "total_amount=99.00", 1))
	if _, err := p.VerifyWebhook(ctx, "alipay", tampered, nil); err == nil {
		t.Fatal("tampered notify must be rejected")
	}
}

func TestAlipayCreateCheckout(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	p := &AlipayProvider{appID: "2021000000000001", privKey: privPEM, pubKey: pubPEM,
		notifyURL: "https://example.com/webhook/alipay", gateway: alipayBase}
	u, err := p.CreateCheckout(context.Background(), "2026082999", 300, "CNY", "Open Novel VIP monthly")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host != "openapi.alipay.com" {
		t.Fatalf("bad checkout url: %v err=%v", u, err)
	}
	q := parsed.Query()
	if q.Get("method") != "alipay.trade.wap.pay" || q.Get("sign_type") != "RSA2" ||
		q.Get("app_id") != "2021000000000001" || q.Get("sign") == "" ||
		!strings.Contains(q.Get("biz_content"), `"out_trade_no":"2026082999"`) {
		t.Fatalf("checkout params missing: %v", q)
	}
	// 从 URL 还原参数并复验签名
	params := map[string]string{}
	for k := range q {
		params[k] = q.Get(k)
	}
	pub, _ := parseRSAPublicKey(pubPEM)
	if err := alipayVerify(params, pub); err != nil {
		t.Fatalf("checkout url signature invalid: %v", err)
	}
	// 未配置密钥 → ErrProviderOn
	bad := &AlipayProvider{appID: "", privKey: "", pubKey: pubPEM}
	if _, err := bad.CreateCheckout(context.Background(), "1", 100, "CNY", "x"); err == nil {
		t.Fatal("missing keys must fail")
	}
}

func mustSign(t *testing.T, params map[string]string, priv *rsa.PrivateKey) string {
	t.Helper()
	s, err := alipaySign(params, priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func copyParams(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
