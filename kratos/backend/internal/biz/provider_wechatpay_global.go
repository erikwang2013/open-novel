package biz

// WeChat Pay Global 渠道（国际微信支付，HK 区域 API v3）：H5 支付。
// 下单 POST {base}/v3/pay/transactions/h5 换取 h5_url 跳转；回调 Webhook 本地
// RSA-SHA256 验签（平台公钥）+ 时间戳防重放，解析 resource 交易字段。
//
// config 键：app_id / mch_id / merchant_serial_no / merchant_private_key（PEM，
// PKCS1/PKCS8）/ platform_public_key（PEM，PKIX/PKCS1）/ apiv3_key（32 字节 APIv3
// 密钥，回调 resource 解密用）/ notify_url
// （可选 base_url，默认 HK 生产 https://apihk.mch.weixin.qq.com，沙箱/其他区域可覆盖）

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const (
	wechatpayGlobalBase = "https://apihk.mch.weixin.qq.com"
	// ponytail: 占位地址收不到回调，上线前必须在渠道 config 配真实 notify_url
	wechatpayNotifyPlaceholder = "https://example.com/webhook/wechatpay_global"
	// wechatpayReplayWindow 回调时间戳允许的最大偏差（秒），防重放
	wechatpayReplayWindow = 300
)

type WeChatPayGlobalProvider struct {
	http        *http.Client
	appID       string
	mchID       string
	serialNo    string
	merchantKey string // 商户私钥 PEM（PKCS1/PKCS8）
	platformKey string // 微信支付平台公钥 PEM（PKIX/PKCS1）
	apiv3Key    string // APIv3 密钥（32 字节），回调 resource 解密
	notifyURL   string
	base        string
}

func newWeChatPayGlobalProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	base := cfgStr(cfg, "base_url")
	if base == "" {
		base = wechatpayGlobalBase
	}
	return &WeChatPayGlobalProvider{
		http:        &http.Client{Timeout: 10 * time.Second},
		appID:       cfgStr(cfg, "app_id"),
		mchID:       cfgStr(cfg, "mch_id"),
		serialNo:    cfgStr(cfg, "merchant_serial_no"),
		merchantKey: cfgStr(cfg, "merchant_private_key"),
		platformKey: cfgStr(cfg, "platform_public_key"),
		apiv3Key:    cfgStr(cfg, "apiv3_key"),
		notifyURL:   cfgStr(cfg, "notify_url"),
		base:        base,
	}, nil
}

// CreateCheckout H5 支付下单：APIv3 签名 POST，返回 h5_url 跳转地址。
func (p *WeChatPayGlobalProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.appID == "" || p.mchID == "" || p.serialNo == "" || p.merchantKey == "" {
		return "", fmt.Errorf("wechatpay_global: %w", pkg.ErrProviderOn)
	}
	notify := p.notifyURL
	if notify == "" {
		notify = wechatpayNotifyPlaceholder
	}
	body := map[string]any{
		"appid":        p.appID,
		"mchid":        p.mchID,
		"description":  desc,
		"out_trade_no": orderNo,
		"notify_url":   notify,
		"amount": map[string]any{
			"total":    amountCents, // 整数分
			"currency": currency,
		},
		"scene_info": map[string]any{
			"h5_info": map[string]any{"type": "Wap"},
		},
	}
	endpoint := p.base + "/v3/pay/transactions/h5"
	auth, rawBody, err := p.signRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return "", err
	}
	var resp struct {
		H5URL string `json:"h5_url"`
	}
	// rawBody 为 json.RawMessage，doJSON 重新 Marshal 时逐字节原样保留，签名与发送一致
	if err := doJSON(ctx, p.http, http.MethodPost, endpoint,
		map[string]string{"Authorization": auth}, json.RawMessage(rawBody), &resp); err != nil {
		return "", err
	}
	if resp.H5URL == "" {
		return "", errors.New("wechatpay_global: empty h5_url in response")
	}
	return resp.H5URL, nil
}

// signRequest 构造 APIv3 请求签名（WECHATPAY2-SHA256-RSA2048）。
// canonical = {HTTP方法}\n{URL路径含查询}\n{timestamp}\n{nonce_str}\n{请求体}\n
// （末行换行必须保留；H5 接口无查询串，路径为 /v3/pay/transactions/h5）。
func (p *WeChatPayGlobalProvider) signRequest(method, endpoint string, body map[string]any) (auth string, raw []byte, err error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := wechatpayNonce()
	raw, err = json.Marshal(body)
	if err != nil {
		return "", nil, err
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", nil, err
	}
	path := u.Path
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	priv, err := parseRSAPrivateKey(p.merchantKey)
	if err != nil {
		return "", nil, err
	}
	canonical := strings.Join([]string{method, path, ts, nonce, string(raw)}, "\n") + "\n"
	h := sha256.Sum256([]byte(canonical))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		return "", nil, err
	}
	auth = fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		p.mchID, nonce, ts, p.serialNo, base64.StdEncoding.EncodeToString(sig))
	return auth, raw, nil
}

// VerifyWebhook 支付回调：Wechatpay 三头验签 → 时间戳防重放 → AES-GCM 解密
// resource（真实回调为 AEAD_AES_256_GCM 加密）→ 解析交易字段。
func (p *WeChatPayGlobalProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.platformKey == "" || p.apiv3Key == "" {
		return PaymentEvent{}, fmt.Errorf("wechatpay_global: %w", pkg.ErrProviderOn)
	}
	ts := headers["Wechatpay-Timestamp"]
	nonce := headers["Wechatpay-Nonce"]
	sigB64 := headers["Wechatpay-Signature"]
	if ts == "" || nonce == "" || sigB64 == "" {
		return PaymentEvent{}, errors.New("wechatpay_global: signature headers missing")
	}
	// 防重放：时间戳与本地偏差超 300s 直接拒绝（快速失败，不继续验签）
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || abs64(time.Now().Unix()-tsInt) > wechatpayReplayWindow {
		return PaymentEvent{}, errors.New("wechatpay_global: timestamp out of replay window")
	}
	pub, err := parseRSAPublicKey(p.platformKey)
	if err != nil {
		return PaymentEvent{}, fmt.Errorf("wechatpay_global: %w", pkg.ErrProviderOn)
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil || len(sig) == 0 {
		return PaymentEvent{}, errors.New("wechatpay_global: bad signature encoding")
	}
	// canonical = {timestamp}\n{nonce}\n{原始body}\n（与官方 SDK 一致）
	canonical := strings.Join([]string{ts, nonce, string(rawBody)}, "\n") + "\n"
	h := sha256.Sum256([]byte(canonical))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return PaymentEvent{}, errors.New("wechatpay_global: notify signature invalid")
	}
	var env struct {
		Resource struct {
			Algorithm      string `json:"algorithm"`
			Ciphertext     string `json:"ciphertext"`
			AssociatedData string `json:"associated_data"`
			Nonce          string `json:"nonce"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(rawBody, &env); err != nil {
		return PaymentEvent{}, errors.New("wechatpay_global: parse event failed")
	}
	plain, err := wechatAesGcmDecrypt(p.apiv3Key, env.Resource.Algorithm,
		env.Resource.Ciphertext, env.Resource.AssociatedData, env.Resource.Nonce)
	if err != nil {
		return PaymentEvent{}, fmt.Errorf("wechatpay_global: %w", err)
	}
	var tx struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		Amount        struct {
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(plain, &tx); err != nil {
		return PaymentEvent{}, errors.New("wechatpay_global: parse resource failed")
	}
	// 仅 SUCCESS 结算；REFUND/NOTPAY 等一律 Paid=false（事件仍返回，由上层 ack）
	return PaymentEvent{OrderNo: tx.OutTradeNo, TxID: tx.TransactionID,
		Paid:        tx.TradeState == "SUCCESS",
		AmountCents: tx.Amount.Total, Currency: tx.Amount.Currency}, nil
}

// wechatAesGcmDecrypt 用 APIv3 密钥解密通知 resource（AEAD_AES_256_GCM）。
func wechatAesGcmDecrypt(key, algorithm, cipherB64, aad, nonce string) ([]byte, error) {
	if algorithm != "AEAD_AES_256_GCM" {
		return nil, errors.New("unsupported resource algorithm " + algorithm)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: apiv3_key must be 32 bytes", pkg.ErrProviderOn)
	}
	ct, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, errors.New("bad ciphertext encoding")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, []byte(nonce), ct, []byte(aad))
}

// wechatpayNonce 随机 32 位十六进制 nonce（APIv3 请求签名）。
func wechatpayNonce() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
