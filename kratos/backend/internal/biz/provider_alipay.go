package biz

// Alipay 渠道（T-P-19 收尾，语言路由 zh-CN）：手机网站支付 alipay.trade.wap.pay。
// 下单不调后端接口，构造网关 GET 跳转 URL（RSA2 签名）；异步 notify 为
// application/x-www-form-urlencoded 表单 POST，RSA2 验签后解析交易字段。
// 沙箱/正式通用：base_url 缺省正式网关，可配 openapi.alipaydev.com 走沙箱。
//
// config 键：app_id / merchant_private_key / alipay_public_key / notify_url
// （可选 base_url）

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const (
	alipayBase = "https://openapi.alipay.com/gateway.do"
	// ponytail: 占位地址收不到回调，上线前必须在渠道 config 配真实 notify_url
	alipayNotifyPlaceholder = "https://example.com/webhook/alipay"
)

type AlipayProvider struct {
	appID     string
	privKey   string // 商户私钥 PEM（PKCS1/PKCS8）
	pubKey    string // 支付宝公钥 PEM（PKIX/PKCS1）
	notifyURL string
	gateway   string
}

func newAlipayProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	base := cfgStr(cfg, "base_url")
	if base == "" {
		base = alipayBase
	}
	return &AlipayProvider{appID: cfgStr(cfg, "app_id"),
		privKey: cfgStr(cfg, "merchant_private_key"), pubKey: cfgStr(cfg, "alipay_public_key"),
		notifyURL: cfgStr(cfg, "notify_url"), gateway: base}, nil
}

// CreateCheckout 构造已签名网关跳转 URL（RSA2 签名后拼到网关地址）。
func (p *AlipayProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.appID == "" || p.privKey == "" {
		return "", fmt.Errorf("alipay: %w", pkg.ErrProviderOn)
	}
	priv, err := parseRSAPrivateKey(p.privKey)
	if err != nil {
		return "", fmt.Errorf("alipay: %w", pkg.ErrProviderOn)
	}
	notify := p.notifyURL
	if notify == "" {
		notify = alipayNotifyPlaceholder
	}
	biz, _ := json.Marshal(map[string]string{
		"out_trade_no": orderNo,
		"total_amount": strconv.FormatFloat(float64(amountCents)/100, 'f', 2, 64), // 元，两位小数
		"subject":      desc,
		"product_code": "QUICK_WAP_WAY",
	})
	params := map[string]string{
		"app_id":      p.appID,
		"method":      "alipay.trade.wap.pay",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  notify,
		"biz_content": string(biz),
	}
	sign, err := alipaySign(params, priv)
	if err != nil {
		return "", err
	}
	params["sign"] = sign
	return p.gateway + "?" + urlValues(params).Encode(), nil
}

// VerifyWebhook 支付宝异步通知：解析表单 → RSA2 验签 → 提取交易字段。
func (p *AlipayProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.pubKey == "" {
		return PaymentEvent{}, fmt.Errorf("alipay: %w", pkg.ErrProviderOn)
	}
	pub, err := parseRSAPublicKey(p.pubKey)
	if err != nil {
		return PaymentEvent{}, fmt.Errorf("alipay: %w", pkg.ErrProviderOn)
	}
	vals, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return PaymentEvent{}, errors.New("alipay: parse notify form failed")
	}
	params := map[string]string{}
	for k := range vals {
		params[k] = vals.Get(k)
	}
	if err := alipayVerify(params, pub); err != nil {
		return PaymentEvent{}, err
	}
	status := params["trade_status"]
	return PaymentEvent{OrderNo: params["out_trade_no"], TxID: params["trade_no"],
		Paid: status == "TRADE_SUCCESS" || status == "TRADE_FINISHED",
		AmountCents: moneyToCents(params["total_amount"]), Currency: "CNY"}, nil
}

// alipaySign 参数按键升序拼接 key=value&（空值跳过），SHA256withRSA 签名，Base64 输出。
// 下单与验签共用同一拼接规则，保证签名串一致。
func alipaySign(params map[string]string, priv *rsa.PrivateKey) (string, error) {
	msg := alipaySignString(params)
	h := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// alipayVerify 校验 sign：按 alipaySign 规则重拼签名串（sign/sign_type 均不参与）。
func alipayVerify(params map[string]string, pub *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(params["sign"])
	if err != nil || len(sig) == 0 {
		return errors.New("alipay: missing or bad sign")
	}
	h := sha256.Sum256([]byte(alipaySignString(params)))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return errors.New("alipay: notify signature invalid")
	}
	return nil
}

func alipaySignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, k := range keys {
		if k == "sign" || k == "sign_type" {
			continue // 支付宝约定：sign/sign_type 不参与签名
		}
		v := params[k]
		if v == "" {
			continue // 空值参数不参与签名
		}
		if !first {
			b.WriteByte('&')
		}
		first = false
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
	}
	return b.String()
}

func urlValues(params map[string]string) url.Values {
	v := url.Values{}
	for k, val := range params {
		v.Set(k, val)
	}
	return v
}

// parseRSAPrivateKey 解析 PKCS1/PKCS8 PEM 商户私钥。
func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid private key PEM")
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rk, ok := k.(*rsa.PrivateKey); ok {
			return rk, nil
		}
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// parseRSAPublicKey 解析 PKIX（BEGIN PUBLIC KEY）或 PKCS1（BEGIN RSA PUBLIC KEY）PEM 支付宝公钥。
func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid public key PEM")
	}
	if k, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rk, ok := k.(*rsa.PublicKey); ok {
			return rk, nil
		}
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
