package biz

// UnionPay 银联在线支付单测：真实 RSA 签名往返（下单/notify 共用 unionpaySign 规则），
// 篡改拒绝、respCode 映射、必填配置缺失 → ErrProviderOn、base_url 覆盖。

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// unionpayNotifyBody 构造已签名 notify 表单体（模拟银联回调：signMethod 在报文中但不参与签名）。
func unionpayNotifyBody(t *testing.T, privPEM string, fields map[string]string) []byte {
	t.Helper()
	priv, err := parseRSAPrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	sign, err := unionpaySign(fields, priv)
	if err != nil {
		t.Fatal(err)
	}
	fields["signature"] = sign
	fields["signMethod"] = "01"
	return []byte(urlValues(fields).Encode())
}

func TestUnionPaySignVerify(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	priv, _ := parseRSAPrivateKey(privPEM)
	pub, _ := parseRSAPublicKey(pubPEM)
	params := map[string]string{
		"version": "5.1.0", "encoding": "utf-8", "signMethod": "01", "txnType": "01",
		"merId": "700000000000001", "orderId": "2026082999", "txnTime": "0829120000",
		"txnAmt": "300", "currencyCode": "156", "backUrl": "https://example.com/webhook/unionpay",
		"certId": "TEST-CERT-1",
	}
	sign, err := unionpaySign(params, priv)
	if err != nil {
		t.Fatal(err)
	}
	params["signature"] = sign
	if err := unionpayVerify(copyParams(params), pub); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	// 篡改任一数据元 → 拒绝
	tampered := copyParams(params)
	tampered["txnAmt"] = "3000"
	if err := unionpayVerify(tampered, pub); err == nil {
		t.Fatal("tampered params must fail")
	}
	// 坏 signature → 拒绝
	wrong := copyParams(params)
	wrong["signature"] = "AAAA"
	if err := unionpayVerify(wrong, pub); err == nil {
		t.Fatal("wrong signature must fail")
	}
}

func TestUnionPayVerifyWebhook(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	p := &UnionPayProvider{pubKey: pubPEM}
	ctx := context.Background()
	base := map[string]string{
		"version": "5.1.0", "encoding": "utf-8", "signMethod": "01", "txnType": "01",
		"merId": "700000000000001", "orderId": "2026082999", "txnTime": "0829120000",
		"txnAmt": "990", "currencyCode": "156", "respCode": "00", "respMsg": "Success",
		"queryId": "2026082912000000001", "certId": "TEST-CERT-1",
	}
	ev, err := p.VerifyWebhook(ctx, "unionpay", unionpayNotifyBody(t, privPEM, copyParams(base)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082999" || ev.TxID != "2026082912000000001" ||
		ev.AmountCents != 990 || ev.Currency != "CNY" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	// respCode != 00 → Paid=false 不报错（上层直接 ack，不结算）
	fail := copyParams(base)
	fail["respCode"] = "34"
	fail["respMsg"] = "Rejected"
	ev, err = p.VerifyWebhook(ctx, "unionpay", unionpayNotifyBody(t, privPEM, fail), nil)
	if err != nil || ev.Paid {
		t.Fatalf("failed resp: err=%v ev=%+v", err, ev)
	}

	// 篡改金额（不重签）→ 验签拒绝
	tampered := unionpayNotifyBody(t, privPEM, copyParams(base))
	tampered = []byte(strings.Replace(string(tampered), "txnAmt=990", "txnAmt=9990", 1))
	if _, err := p.VerifyWebhook(ctx, "unionpay", tampered, nil); err == nil {
		t.Fatal("tampered notify must be rejected")
	}

	// 错公钥（另一把测试钥）→ 拒绝
	_, otherPub := testAlipayKeys(t)
	if _, err := (&UnionPayProvider{pubKey: otherPub}).VerifyWebhook(ctx, "unionpay", unionpayNotifyBody(t, privPEM, copyParams(base)), nil); err == nil {
		t.Fatal("wrong pubkey must fail")
	}

	// 缺 unionpay_public_key → ErrProviderOn
	if _, err := (&UnionPayProvider{}).VerifyWebhook(ctx, "unionpay", []byte("x=1"), nil); err == nil {
		t.Fatal("missing pubkey must fail")
	}
}

func TestUnionPayCreateCheckout(t *testing.T) {
	privPEM, pubPEM := testAlipayKeys(t)
	p := &UnionPayProvider{merID: "700000000000001", certID: "TEST-CERT-1", privKey: privPEM, pubKey: pubPEM,
		notifyURL: "https://example.com/webhook/unionpay", frontURL: "https://example.com/vip",
		gateway: unionpayBase}
	u, err := p.CreateCheckout(context.Background(), "2026082999", 300, "CNY", "Open Novel VIP monthly")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host != "gateway.95516.com" {
		t.Fatalf("bad checkout url: %v err=%v", u, err)
	}
	q := parsed.Query()
	for key, want := range map[string]string{
		"version": "5.1.0", "encoding": "utf-8", "signMethod": "01", "txnType": "01",
		"txnSubType": "01", "bizType": "000201", "channelType": "07", "accessType": "0",
		"merId": "700000000000001", "orderId": "2026082999", "txnAmt": "300",
		"currencyCode": "156", "backUrl": "https://example.com/webhook/unionpay",
		"frontUrl": "https://example.com/vip", "certId": "TEST-CERT-1",
	} {
		if q.Get(key) != want {
			t.Fatalf("%s = %q, want %q", key, q.Get(key), want)
		}
	}
	if !regexp.MustCompile(`^\d{10}$`).MatchString(q.Get("txnTime")) {
		t.Fatalf("bad txnTime: %q", q.Get("txnTime"))
	}
	// 从 URL 还原参数并复验签名（签名字段在 query 中，签名串不含 signature/signMethod）
	params := map[string]string{}
	for k := range q {
		params[k] = q.Get(k)
	}
	pub, _ := parseRSAPublicKey(pubPEM)
	if err := unionpayVerify(params, pub); err != nil {
		t.Fatalf("checkout signature invalid: %v", err)
	}

	// orderId 超 32 字符 → 拒绝
	if _, err := p.CreateCheckout(context.Background(), strings.Repeat("1", 33), 300, "CNY", "x"); err == nil {
		t.Fatal("long orderId must fail")
	}
	// 缺必填配置 → ErrProviderOn
	bad := &UnionPayProvider{merID: "", certID: "", privKey: "", notifyURL: ""}
	if _, err := bad.CreateCheckout(context.Background(), "1", 100, "CNY", "x"); err == nil {
		t.Fatal("missing config must fail")
	}
}

func TestUnionPayBaseURLOverride(t *testing.T) {
	privPEM, _ := testAlipayKeys(t)
	p := &UnionPayProvider{merID: "700000000000001", certID: "TEST-CERT-1", privKey: privPEM,
		notifyURL: "https://example.com/webhook/unionpay", gateway: "http://58.246.226.99/UpopWeb/api/Pay.action"}
	u, err := p.CreateCheckout(context.Background(), "1", 100, "CNY", "x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "http://58.246.226.99/UpopWeb/api/Pay.action?") {
		t.Fatalf("base_url override failed: %v", u)
	}
}
