package biz

// PortOne（旧 IamPort）渠道（T-P-19，语言路由 ko）：REST v1 POST /payments/{paymentId}
// 发起支付返回 hosted 页面；webhook 用 X-IAMPORT-TOKEN 与 config 的 webhook_token 比对
// + rawBody 解析。
//
// config 键：api_secret / webhook_token

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

const portoneBase = "https://api.portone.io"

type PortOneProvider struct {
	http         *http.Client
	apiSecret    string
	webhookToken string
}

func newPortOneProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	return &PortOneProvider{http: &http.Client{Timeout: 10 * time.Second},
		apiSecret: cfgStr(cfg, "api_secret"), webhookToken: cfgStr(cfg, "webhook_token")}, nil
}

func (p *PortOneProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.apiSecret == "" {
		return "", fmt.Errorf("portone: %w", pkg.ErrProviderOn)
	}
	body := map[string]any{
		"amount":      map[string]any{"total": amountCents, "currency": currency},
		"orderName":   desc,
		"merchantUid": orderNo,
	}
	var resp struct {
		PaymentURL string `json:"paymentUrl"`
		Payment    struct {
			PaymentURL string `json:"paymentUrl"`
		} `json:"payment"`
	}
	auth := map[string]string{"Authorization": "PortOne " + p.apiSecret}
	if err := doJSON(ctx, p.http, http.MethodPost, portoneBase+"/v1/payments/"+orderNo, auth, body, &resp); err != nil {
		return "", err
	}
	url := resp.PaymentURL
	if url == "" {
		url = resp.Payment.PaymentURL
	}
	if url == "" {
		return "", errors.New("portone: empty payment url")
	}
	return url, nil
}

// portoneEvent 是 webhook payload（v1 Webhook 事件）。
type portoneEvent struct {
	Type string `json:"type"`
	Data struct {
		PaymentID   string `json:"paymentId"`
		MerchantUID string `json:"merchantUid"`
		Status      string `json:"status"`
		Amount      struct {
			Total    int64  `json:"total"`
			Currency string `json:"currency"`
		} `json:"amount"`
	} `json:"data"`
}

func (p *PortOneProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.webhookToken == "" {
		return PaymentEvent{}, errors.New("portone: webhook token not configured")
	}
	// 官方 v1 webhook 无 HMAC：X-IAMPORT-TOKEN 与配置 token 比对。
	if !hmac.Equal([]byte(headers["X-IAMPORT-TOKEN"]), []byte(p.webhookToken)) {
		return PaymentEvent{}, errors.New("portone: webhook token mismatch")
	}
	var evt portoneEvent
	if err := json.Unmarshal(rawBody, &evt); err != nil {
		return PaymentEvent{}, errors.New("portone: parse event failed")
	}
	if evt.Type != "Payment.Paid" && evt.Data.Status != "PAID" && evt.Data.Status != "paid" {
		return PaymentEvent{}, nil // 非支付成功事件直接 ack
	}
	return PaymentEvent{OrderNo: evt.Data.MerchantUID, TxID: evt.Data.PaymentID, Paid: true,
		AmountCents: evt.Data.Amount.Total, Currency: evt.Data.Amount.Currency}, nil
}
