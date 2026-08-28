package biz

// Razorpay 渠道（T-P-19，语言路由 hi）：REST v1 创建 payment link，webhook 用
// HMAC-SHA256(webhook_secret, rawBody) 十六进制验签（官方 X-Razorpay-Signature）。
// amount 单位为分（INR 最小货币单位），currency 由订单传入。
//
// config 键（admin「支付方式」行配置，DB 加密）：key_id / key_secret / webhook_secret

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const razorpayBase = "https://api.razorpay.com"

type RazorpayProvider struct {
	http       *http.Client
	keyID      string
	keySec     string
	webhookSec string
}

func newRazorpayProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	return &RazorpayProvider{http: &http.Client{Timeout: 10 * time.Second},
		keyID: cfgStr(cfg, "key_id"), keySec: cfgStr(cfg, "key_secret"),
		webhookSec: cfgStr(cfg, "webhook_secret")}, nil
}

func (p *RazorpayProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.keyID == "" || p.keySec == "" {
		return "", fmt.Errorf("razorpay: %w", pkg.ErrProviderOn)
	}
	body := map[string]any{
		"amount":         amountCents, // 印度卢比：单位为分
		"currency":       currency,
		"accept_partial": false,
		"description":    desc,
		"notes":          map[string]string{"order_no": orderNo},
	}
	var resp struct {
		ID       string `json:"id"`
		ShortURL string `json:"short_url"`
	}
	auth := map[string]string{"Authorization": basicAuthHeader(p.keyID, p.keySec)}
	if err := doJSON(ctx, p.http, http.MethodPost, razorpayBase+"/v1/payment_links", auth, body, &resp); err != nil {
		return "", err
	}
	if resp.ShortURL == "" {
		return "", errors.New("razorpay: empty short_url")
	}
	return resp.ShortURL, nil
}

// razorpayEvent 是 webhook payload（payment_link 事件）。
type razorpayEvent struct {
	Event   string `json:"event"`
	Payload struct {
		PaymentLink struct {
			Entity struct {
				ID       string `json:"id"`
				Status   string `json:"status"`
				Amount   int64  `json:"amount"`
				Currency string `json:"currency"`
				Notes    struct {
					OrderNo string `json:"order_no"`
				} `json:"notes"`
			} `json:"entity"`
		} `json:"payment_link"`
	} `json:"payload"`
}

func (p *RazorpayProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.webhookSec == "" {
		return PaymentEvent{}, errors.New("razorpay: webhook secret not configured")
	}
	sig, _ := hex.DecodeString(headers["X-Razorpay-Signature"])
	mac := hmac.New(sha256.New, []byte(p.webhookSec))
	mac.Write(rawBody)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return PaymentEvent{}, errors.New("razorpay: webhook signature invalid")
	}
	var evt razorpayEvent
	if err := json.Unmarshal(rawBody, &evt); err != nil {
		return PaymentEvent{}, errors.New("razorpay: parse event failed")
	}
	if evt.Event != "payment_link.paid" {
		return PaymentEvent{}, nil // 非支付成功事件直接 ack
	}
	e := evt.Payload.PaymentLink.Entity
	return PaymentEvent{OrderNo: e.Notes.OrderNo, TxID: e.ID, Paid: true,
		AmountCents: e.Amount, Currency: e.Currency}, nil
}
