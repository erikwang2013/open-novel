package biz

// Mercado Pago 渠道（T-P-19，语言路由 pt-BR）：POST /checkout/preferences 创建，
// init_point 返回跳转 URL；IPN webhook 无签名，按官方标准做法收到通知后
// 用 access_token GET /v1/payments/{id} 核对 status=approved 与金额（token 查询验证）。
//
// config 键：access_token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const mpBase = "https://api.mercadopago.com"

type MercadoPagoProvider struct {
	http        *http.Client
	accessToken string
}

func newMercadoPagoProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	return &MercadoPagoProvider{http: &http.Client{Timeout: 10 * time.Second},
		accessToken: cfgStr(cfg, "access_token")}, nil
}

func (p *MercadoPagoProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.accessToken == "" {
		return "", fmt.Errorf("mercadopago: %w", pkg.ErrProviderOn)
	}
	body := map[string]any{
		"external_reference": orderNo,
		"auto_return":        "approved",
		"items": []map[string]any{{
			"id": orderNo, "title": desc, "quantity": 1,
			"unit_price": float64(amountCents) / 100, "currency_id": currency,
		}},
	}
	var resp struct {
		ID        string `json:"id"`
		InitPoint string `json:"init_point"`
	}
	auth := map[string]string{"Authorization": "Bearer " + p.accessToken}
	if err := doJSON(ctx, p.http, http.MethodPost, mpBase+"/checkout/preferences", auth, body, &resp); err != nil {
		return "", err
	}
	if resp.InitPoint == "" {
		return "", errors.New("mercadopago: empty init_point")
	}
	return resp.InitPoint, nil
}

func (p *MercadoPagoProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.accessToken == "" {
		return PaymentEvent{}, errors.New("mercadopago: access token not configured")
	}
	var notif struct {
		ID   int64  `json:"id"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(rawBody, &notif); err != nil {
		return PaymentEvent{}, errors.New("mercadopago: parse notification failed")
	}
	if notif.ID == 0 {
		return PaymentEvent{}, errors.New("mercadopago: missing payment id")
	}
	// 官方 IPN 无 webhook 签名：回查支付详情核对状态与金额。
	var pay struct {
		ID                int64   `json:"id"`
		Status            string  `json:"status"`
		ExternalReference string  `json:"external_reference"`
		CurrencyID        string  `json:"currency_id"`
		TransactionAmount float64 `json:"transaction_amount"`
	}
	auth := map[string]string{"Authorization": "Bearer " + p.accessToken}
	url := mpBase + "/v1/payments/" + strconv.FormatInt(notif.ID, 10)
	if err := doJSON(ctx, p.http, http.MethodGet, url, auth, nil, &pay); err != nil {
		return PaymentEvent{}, err
	}
	if pay.Status != "approved" {
		return PaymentEvent{}, nil
	}
	return PaymentEvent{OrderNo: pay.ExternalReference, TxID: strconv.FormatInt(pay.ID, 10), Paid: true,
		AmountCents: int64(math.Round(pay.TransactionAmount * 100)), Currency: pay.CurrencyID}, nil
}
