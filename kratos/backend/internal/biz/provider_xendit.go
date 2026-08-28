package biz

// Xendit 渠道（T-P-19，语言路由 id/th/vn）：legacy invoice v2 创建支付页，
// webhook 用 X-CALLBACK-TOKEN 与 config 的 callback_token 比对 + rawBody 金额解析。
//
// config 键：api_key / callback_token

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const xenditBase = "https://api.xendit.co"

type XenditProvider struct {
	http          *http.Client
	apiKey        string
	callbackToken string
}

func newXenditProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	return &XenditProvider{http: &http.Client{Timeout: 10 * time.Second},
		apiKey: cfgStr(cfg, "api_key"), callbackToken: cfgStr(cfg, "callback_token")}, nil
}

func (p *XenditProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("xendit: %w", pkg.ErrProviderOn)
	}
	body := map[string]any{
		"external_id": orderNo,
		"amount":      amountCents, // 最小货币单位（整数分）
		"currency":    currency,
		"description": desc,
	}
	var resp struct {
		ID         string `json:"id"`
		InvoiceURL string `json:"invoice_url"`
	}
	headers := map[string]string{"X-API-Key": p.apiKey, "Idempotency-Key": orderNo}
	if err := doJSON(ctx, p.http, http.MethodPost, xenditBase+"/v2/invoices", headers, body, &resp); err != nil {
		return "", err
	}
	if resp.InvoiceURL == "" {
		return "", errors.New("xendit: empty invoice_url")
	}
	return resp.InvoiceURL, nil
}

// xenditCallback 是 invoice webhook payload。
type xenditCallback struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
	Status     string `json:"status"`
	Amount     int64  `json:"amount"`
	PaidAmount int64  `json:"paid_amount"`
	Currency   string `json:"currency"`
}

func (p *XenditProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.callbackToken == "" {
		return PaymentEvent{}, errors.New("xendit: callback token not configured")
	}
	if !hmac.Equal([]byte(headers["X-CALLBACK-TOKEN"]), []byte(p.callbackToken)) {
		return PaymentEvent{}, errors.New("xendit: callback token mismatch")
	}
	var cb xenditCallback
	if err := json.Unmarshal(rawBody, &cb); err != nil {
		return PaymentEvent{}, errors.New("xendit: parse callback failed")
	}
	if cb.Status != "PAID" {
		return PaymentEvent{}, nil
	}
	cents := cb.Amount
	if cents <= 0 {
		cents = cb.PaidAmount
	}
	return PaymentEvent{OrderNo: cb.ExternalID, TxID: cb.ID, Paid: true,
		AmountCents: cents, Currency: cb.Currency}, nil
}
