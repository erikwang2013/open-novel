package biz

// T-P-19/20 新渠道验签单测：Razorpay/KOMOJU HMAC-SHA256 向量、PortOne/Xendit token 比对、
// Mercado Pago IPN payment_id 解析 + 回查模拟、PayPal 官方 verify-webhook-signature 验签 mock。

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
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
