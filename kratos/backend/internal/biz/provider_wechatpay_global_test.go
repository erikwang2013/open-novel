package biz

// WeChat Pay Global（HK API v3）单测：请求/webhook 两组 canonical 签名往返 + 篡改拒绝、
// VerifyWebhook 状态映射与防重放、CreateCheckout 签名可复验 + base_url 覆盖。

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// wechatTestKeys 生成商户/平台两对 RSA-2048 密钥（复用 testAlipayKeys 的 PEM 编码）。
func wechatTestKeys(t *testing.T) (merchantPriv, merchantPub, platformPriv, platformPub string) {
	t.Helper()
	merchantPriv, merchantPub = testAlipayKeys(t)
	platformPriv, platformPub = testAlipayKeys(t)
	return merchantPriv, merchantPub, platformPriv, platformPub
}

// wechatEncrypt 用 apiv3Key 对 resource 明文做 AES-256-GCM 加密（固定 12 字节 nonce）。
func wechatEncrypt(t *testing.T, key, aad, plaintext string) (cipherB64, nonce string) {
	t.Helper()
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce = "0123456789ab"
	ct := gcm.Seal(nil, []byte(nonce), []byte(plaintext), []byte(aad))
	return base64.StdEncoding.EncodeToString(ct), nonce
}

// wechatNotify 构造真实形态回调：AES-GCM 加密 resource + 平台私钥对原始 body 验签。
func wechatNotify(t *testing.T, platformPriv, apiv3Key, state, orderNo string, total int64, currency string) ([]byte, map[string]string) {
	t.Helper()
	plain := fmt.Sprintf(`{"out_trade_no":%q,"transaction_id":"4200000000001","trade_state":%q,"amount":{"total":%d,"currency":%q}}`,
		orderNo, state, total, currency)
	cipher, nonce := wechatEncrypt(t, apiv3Key, "transaction", plain)
	body := []byte(fmt.Sprintf(`{"id":"EV-1","event_type":"TRANSACTION.SUCCESS","resource_type":"encrypt-resource","resource":{"original_type":"transaction","algorithm":"AEAD_AES_256_GCM","ciphertext":%q,"associated_data":"transaction","nonce":%q}}`, cipher, nonce))
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sigNonce := "sig-nonce-1"
	hdrs := map[string]string{"Wechatpay-Timestamp": ts, "Wechatpay-Nonce": sigNonce,
		"Wechatpay-Signature": wechatSign(t, platformPriv, ts, sigNonce, string(body))}
	return body, hdrs
}

// wechatSign 对 canonical 片段（自动按 \n 连接并补末行换行）做 RSA-SHA256 签名，Base64 输出。
func wechatSign(t *testing.T, privPEM string, parts ...string) string {
	t.Helper()
	priv, err := parseRSAPrivateKey(privPEM)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(strings.Join(parts, "\n") + "\n"))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}

func TestWeChatPaySignVerify(t *testing.T) {
	merchantPriv, merchantPub, platformPriv, platformPub := wechatTestKeys(t)
	ts, nonce := "1724900000", "abcdef0123456789"

	// 请求 canonical（下单）：POST\npath\nts\nnonce\nbody\n → 商户公钥可验
	body := `{"appid":"wx-app-1","mchid":"1900000001","amount":{"total":300,"currency":"HKD"}}`
	sig := wechatSign(t, merchantPriv, "POST", "/v3/pay/transactions/h5", ts, nonce, body)
	mPub, _ := parseRSAPublicKey(merchantPub)
	h := sha256.Sum256([]byte(strings.Join([]string{"POST", "/v3/pay/transactions/h5", ts, nonce, body}, "\n") + "\n"))
	sigRaw, _ := base64.StdEncoding.DecodeString(sig)
	if err := rsa.VerifyPKCS1v15(mPub, crypto.SHA256, h[:], sigRaw); err != nil {
		t.Fatalf("request signature invalid: %v", err)
	}
	// 篡改 body → 拒绝
	h = sha256.Sum256([]byte(strings.Join([]string{"POST", "/v3/pay/transactions/h5", ts, nonce, body + " "}, "\n") + "\n"))
	if err := rsa.VerifyPKCS1v15(mPub, crypto.SHA256, h[:], sigRaw); err == nil {
		t.Fatal("tampered request canonical must fail")
	}

	// webhook canonical（回调）：ts\nnonce\nbody\n → 平台公钥可验
	whBody := `{"resource":{"out_trade_no":"2026082999"}}`
	whSig := wechatSign(t, platformPriv, ts, nonce, whBody)
	pPub, _ := parseRSAPublicKey(platformPub)
	h = sha256.Sum256([]byte(strings.Join([]string{ts, nonce, whBody}, "\n") + "\n"))
	whRaw, _ := base64.StdEncoding.DecodeString(whSig)
	if err := rsa.VerifyPKCS1v15(pPub, crypto.SHA256, h[:], whRaw); err != nil {
		t.Fatalf("webhook signature invalid: %v", err)
	}
	// 平台公钥验不了商户签名（密钥不匹配）→ 拒绝
	h = sha256.Sum256([]byte(strings.Join([]string{"POST", "/v3/pay/transactions/h5", ts, nonce, body}, "\n") + "\n"))
	if err := rsa.VerifyPKCS1v15(pPub, crypto.SHA256, h[:], sigRaw); err == nil {
		t.Fatal("wrong key must fail")
	}
}

func TestWeChatPayVerifyWebhook(t *testing.T) {
	_, _, platformPriv, platformPub := wechatTestKeys(t)
	const apiv3Key = "0123456789abcdef0123456789abcdef"
	p := &WeChatPayGlobalProvider{platformKey: platformPub, apiv3Key: apiv3Key}
	ctx := context.Background()

	// SUCCESS → Paid=true 且订单号/金额/币种正确
	body, hdrs := wechatNotify(t, platformPriv, apiv3Key, "SUCCESS", "2026082999", 990, "HKD")
	ev, err := p.VerifyWebhook(ctx, "wechatpay_global", body, hdrs)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082999" || ev.TxID != "4200000000001" ||
		ev.AmountCents != 990 || ev.Currency != "HKD" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	// 篡改 body → 拒绝
	if _, err := p.VerifyWebhook(ctx, "wechatpay_global", append(body, ' '), hdrs); err == nil {
		t.Fatal("tampered body must fail")
	}
	// 时间戳超窗（600s 前）→ 拒绝
	old := strconv.FormatInt(time.Now().Unix()-600, 10)
	sigNonce := "sig-nonce-1"
	oldHdrs := map[string]string{"Wechatpay-Timestamp": old, "Wechatpay-Nonce": sigNonce,
		"Wechatpay-Signature": wechatSign(t, platformPriv, old, sigNonce, string(body))}
	if _, err := p.VerifyWebhook(ctx, "wechatpay_global", body, oldHdrs); err == nil {
		t.Fatal("stale timestamp must fail")
	}
	// 验签头缺失 → 拒绝
	if _, err := p.VerifyWebhook(ctx, "wechatpay_global", body, map[string]string{"Wechatpay-Timestamp": old}); err == nil {
		t.Fatal("missing headers must fail")
	}
	// 错误 apiv3_key → 解密失败拒绝
	body2, hdrs2 := wechatNotify(t, platformPriv, "ffffffffffffffffffffffffffffffff", "SUCCESS", "2026082997", 990, "HKD")
	if _, err := p.VerifyWebhook(ctx, "wechatpay_global", body2, hdrs2); err == nil {
		t.Fatal("wrong apiv3_key must fail")
	}
	// 缺 apiv3_key（config 未配）→ ErrProviderOn
	if _, err := (&WeChatPayGlobalProvider{platformKey: platformPub}).VerifyWebhook(ctx, "wechatpay_global", body, hdrs); err == nil {
		t.Fatal("missing apiv3_key must fail")
	}

	// NOTPAY / REFUND → Paid=false 不报错
	for _, state := range []string{"NOTPAY", "REFUND"} {
		b, hdr := wechatNotify(t, platformPriv, apiv3Key, state, "2026082998", 990, "HKD")
		ev, err := p.VerifyWebhook(ctx, "wechatpay_global", b, hdr)
		if err != nil || ev.Paid {
			t.Fatalf("%s: err=%v ev=%+v", state, err, ev)
		}
	}
}

func TestWeChatPayCreateCheckout(t *testing.T) {
	merchantPriv, merchantPub, _, _ := wechatTestKeys(t)
	var gotAuth, gotBody, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Write([]byte(`{"h5_url":"https://wx.tenpay.com/cgi-bin/mmpayweb/..."}`))
	}))
	defer srv.Close()
	p := &WeChatPayGlobalProvider{
		http: srv.Client(), appID: "wx-app-1", mchID: "1900000001", serialNo: "SERIAL-1",
		merchantKey: merchantPriv, notifyURL: "https://example.com/webhook/wechatpay_global",
		base: srv.URL, // base_url 覆盖：默认 apihk 应被替换为测试服务器
	}
	u, err := p.CreateCheckout(context.Background(), "2026082999", 300, "HKD", "Open Novel VIP monthly")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://wx.tenpay.com/cgi-bin/mmpayweb/..." {
		t.Fatalf("unexpected h5_url: %v", u)
	}
	if gotPath != "/v3/pay/transactions/h5" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048 ") {
		t.Fatalf("bad auth scheme: %s", gotAuth)
	}
	// 解析 Authorization 参数，用商户公钥复验 canonical 签名
	params := map[string]string{}
	for _, part := range strings.Split(gotAuth[len("WECHATPAY2-SHA256-RSA2048 "):], ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.Trim(kv[1], `"`)
		}
	}
	if params["mchid"] != "1900000001" || params["serial_no"] != "SERIAL-1" ||
		params["timestamp"] == "" || params["nonce_str"] == "" || params["signature"] == "" {
		t.Fatalf("bad authorization params: %v", params)
	}
	canonical := strings.Join([]string{"POST", gotPath, params["timestamp"], params["nonce_str"], gotBody}, "\n") + "\n"
	mPub, _ := parseRSAPublicKey(merchantPub)
	h := sha256.Sum256([]byte(canonical))
	sigRaw, _ := base64.StdEncoding.DecodeString(params["signature"])
	if err := rsa.VerifyPKCS1v15(mPub, crypto.SHA256, h[:], sigRaw); err != nil {
		t.Fatalf("request signature invalid: %v", err)
	}
	// 请求体字段完整
	var sent map[string]any
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["out_trade_no"] != "2026082999" || sent["appid"] != "wx-app-1" || sent["mchid"] != "1900000001" ||
		sent["notify_url"] != "https://example.com/webhook/wechatpay_global" {
		t.Fatalf("body missing fields: %v", sent)
	}
	amt := sent["amount"].(map[string]any)
	if amt["total"] != float64(300) || amt["currency"] != "HKD" {
		t.Fatalf("bad amount: %v", amt)
	}
	scene := sent["scene_info"].(map[string]any)["h5_info"].(map[string]any)
	if scene["type"] != "Wap" {
		t.Fatalf("bad scene_info: %v", sent["scene_info"])
	}

	// 缺密钥 → 报错
	bad := &WeChatPayGlobalProvider{http: srv.Client(), appID: "", mchID: "", serialNo: "", merchantKey: ""}
	if _, err := bad.CreateCheckout(context.Background(), "1", 100, "HKD", "x"); err == nil {
		t.Fatal("missing keys must fail")
	}
}
