package biz

// KOMOJU 渠道（T-P-19，语言路由 ja）：REST v1 创建 hosted payment page，
// webhook 用 HMAC-SHA256(secret, rawBody) 十六进制验签（官方 X-KOMOJU-SIGNATURE）。
// 金额：KOMOJU amount 为最小货币单位（JPY 无分位，单位即日元）；webhook 的
// data.amounts.cents 为分位值、units 为整数元——分位缺失时退回 units（JPY 两者等价）。
//
// config 键：api_key / webhook_secret

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const komojuBase = "https://komoju.com/api/v1"

type KomojuProvider struct {
	http       *http.Client
	apiKey     string
	webhookSec string
}

func newKomojuProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	return &KomojuProvider{http: &http.Client{Timeout: 10 * time.Second},
		apiKey: cfgStr(cfg, "api_key"), webhookSec: cfgStr(cfg, "webhook_secret")}, nil
}

func (p *KomojuProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("komoju: %w", pkg.ErrProviderOn)
	}
	body := map[string]any{
		"amount":            strconv.FormatInt(amountCents, 10),
		"currency":          currency,
		"external_order_no": orderNo,
		"description":       desc,
		"payment_details":   map[string]string{"type": "hosted_page"},
	}
	var resp struct {
		ID         string `json:"id"`
		PaymentURL string `json:"payment_url"`
	}
	auth := map[string]string{"Authorization": "Bearer " + p.apiKey}
	if err := doJSON(ctx, p.http, http.MethodPost, komojuBase+"/v1/payments", auth, body, &resp); err != nil {
		return "", err
	}
	if resp.PaymentURL == "" {
		return "", errors.New("komoju: empty payment_url")
	}
	return resp.PaymentURL, nil
}

// komojuEvent 是 webhook payload。
type komojuEvent struct {
	Type string `json:"type"`
	Data struct {
		ID              string `json:"id"`
		Status          string `json:"status"`
		Currency        string `json:"currency"`
		ExternalOrderNo string `json:"external_order_no"`
		Amounts         struct {
			Cents int64 `json:"cents"`
			Units int64 `json:"units"`
		} `json:"amounts"`
	} `json:"data"`
}

func (p *KomojuProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.webhookSec == "" {
		return PaymentEvent{}, errors.New("komoju: webhook secret not configured")
	}
	sig, _ := hex.DecodeString(headers["X-KOMOJU-SIGNATURE"])
	mac := hmac.New(sha256.New, []byte(p.webhookSec))
	mac.Write(rawBody)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return PaymentEvent{}, errors.New("komoju: webhook signature invalid")
	}
	var evt komojuEvent
	if err := json.Unmarshal(rawBody, &evt); err != nil {
		return PaymentEvent{}, errors.New("komoju: parse event failed")
	}
	if evt.Type != "payment.paid" {
		return PaymentEvent{}, nil // 非支付成功事件直接 ack
	}
	cents := evt.Data.Amounts.Cents
	if cents == 0 {
		cents = evt.Data.Amounts.Units // JPY 无分位：units（整数日元）即分位等价
	}
	return PaymentEvent{OrderNo: evt.Data.ExternalOrderNo, TxID: evt.Data.ID, Paid: true,
		AmountCents: cents, Currency: evt.Data.Currency}, nil
}
