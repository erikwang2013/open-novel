package biz

// UnionPay 银联在线支付渠道（网关支付 5.1，语言路由 zh-CN）：
// 下单不调后端接口，构造网关跳转 URL（form 参数 + SHA256withRSA 签名，与 alipay 同构，
// 客户端按 URL 直接打开，零改动）；后台 notify 为表单 POST，银联公钥验签后按
// respCode=00 判定成功，应答 success 纯文本（HTTP 200），重试/幂等由上层 settle 保证。
// base_url 缺省生产网关，可配沙箱地址覆盖（与 alipay 的 base_url 模式一致）。
//
// config 键：mer_id / sign_cert_id / merchant_private_key / unionpay_public_key /
// notify_url（均必填），front_url / base_url（可选）

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
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
	unionpayBase       = "https://gateway.95516.com/gateway/api/frontTransReq.do"
	unionpaySignMethod = "01" // RSA
)

type UnionPayProvider struct {
	merID     string
	certID    string // 签名证书 certId
	privKey   string // 商户私钥 PEM（PKCS1/PKCS8）
	pubKey    string // 银联验签公钥 PEM（PKIX/PKCS1）
	notifyURL string
	frontURL  string
	gateway   string
}

func newUnionPayProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	base := cfgStr(cfg, "base_url")
	if base == "" {
		base = unionpayBase
	}
	return &UnionPayProvider{merID: cfgStr(cfg, "mer_id"), certID: cfgStr(cfg, "sign_cert_id"),
		privKey: cfgStr(cfg, "merchant_private_key"), pubKey: cfgStr(cfg, "unionpay_public_key"),
		notifyURL: cfgStr(cfg, "notify_url"), frontURL: cfgStr(cfg, "front_url"),
		gateway: base}, nil
}

// CreateCheckout 构造已签名网关跳转 URL（银联前台表单参数以 GET query 传递，与 alipay 同构）。
func (p *UnionPayProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.merID == "" || p.certID == "" || p.privKey == "" || p.notifyURL == "" {
		return "", fmt.Errorf("unionpay: %w", pkg.ErrProviderOn)
	}
	priv, err := parseRSAPrivateKey(p.privKey)
	if err != nil {
		return "", fmt.Errorf("unionpay: %w", pkg.ErrProviderOn)
	}
	if len(orderNo) > 32 { // 银联 orderId ≤ 32 字符
		return "", errors.New("unionpay: orderId too long")
	}
	params := map[string]string{
		"version":      "5.1.0",
		"encoding":     "utf-8",
		"signMethod":   unionpaySignMethod,
		"txnType":      "01", // 消费
		"txnSubType":   "01",
		"bizType":      "000201",
		"channelType":  "07", // 互联网
		"accessType":   "0",
		"merId":        p.merID,
		"orderId":      orderNo,
		"txnTime":      txnTimeNow(), // MMddHHmmss，UTC+8
		"txnAmt":       strconv.FormatInt(amountCents, 10),
		"currencyCode": "156", // CNY，ponytail: 银联多币种接入时按币种码表映射
		"backUrl":      p.notifyURL,
		"certId":       p.certID,
	}
	if p.frontURL != "" {
		params["frontUrl"] = p.frontURL
	}
	sign, err := unionpaySign(params, priv)
	if err != nil {
		return "", err
	}
	params["signature"] = sign
	return p.gateway + "?" + urlValues(params).Encode(), nil
}

// VerifyWebhook 银联后台 notify：解析表单 → 银联公钥验签 → 提取交易字段。
func (p *UnionPayProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.pubKey == "" {
		return PaymentEvent{}, fmt.Errorf("unionpay: %w", pkg.ErrProviderOn)
	}
	pub, err := parseRSAPublicKey(p.pubKey)
	if err != nil {
		return PaymentEvent{}, fmt.Errorf("unionpay: %w", pkg.ErrProviderOn)
	}
	vals, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return PaymentEvent{}, errors.New("unionpay: parse notify form failed")
	}
	params := map[string]string{}
	for k := range vals {
		params[k] = vals.Get(k)
	}
	if err := unionpayVerify(params, pub); err != nil {
		return PaymentEvent{}, err
	}
	amt, _ := strconv.ParseInt(params["txnAmt"], 10, 64)
	return PaymentEvent{OrderNo: params["orderId"], TxID: params["queryId"],
		Paid: params["respCode"] == "00",
		AmountCents: amt, Currency: unionpayCurrencyName(params["currencyCode"])}, nil
}

// unionpaySign 按银联规范：除 signature/signMethod 外所有数据元字典序（ASCII 升序）
// 拼 key=value&（原生值，不 URL 编码），SHA-256 摘要 → RSA 签名 → Base64。
func unionpaySign(params map[string]string, priv *rsa.PrivateKey) (string, error) {
	h := sha256.Sum256([]byte(unionpaySignString(params)))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// unionpayVerify 校验 signature：按 unionpaySign 规则重拼签名串，银联公钥验证。
func unionpayVerify(params map[string]string, pub *rsa.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(params["signature"])
	if err != nil || len(sig) == 0 {
		return errors.New("unionpay: missing or bad signature")
	}
	h := sha256.Sum256([]byte(unionpaySignString(params)))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sig); err != nil {
		return errors.New("unionpay: notify signature invalid")
	}
	return nil
}

func unionpaySignString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, k := range keys {
		if k == "signature" || k == "signMethod" {
			continue // 银联约定：signature/signMethod 不参与签名
		}
		v := params[k]
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

// txnTimeNow 银联订单时间：MMddHHmmss，取当前 UTC+8。
func txnTimeNow() string {
	return time.Now().In(time.FixedZone("CST", 8*3600)).Format("0102150405")
}

// unionpayCurrencyName 银联币种码 → ISO 码；156=CNY，未知码原样透传
// （ponytail: 金额强校验在上层，币种不符会 ErrAmountMism 拒绝）。
func unionpayCurrencyName(code string) string {
	if code == "" || code == "156" {
		return "CNY"
	}
	return code
}
