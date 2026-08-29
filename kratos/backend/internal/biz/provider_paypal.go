package biz

// PayPal 渠道（T-P-20，兜底）：Orders API v2——OAuth2 换 token → POST /v2/checkout/orders
// → 取 approve link 跳转；webhook 调官方 /v1/notifications/verify-webhook-signature 验签
// （复用 OAuth access_token），见 VerifyWebhook。
//
// config 键：client_id / client_secret / webhook_id；可选 base_url（沙箱
// https://api-m.sandbox.paypal.com）

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

const (
	paypalBase    = "https://api-m.paypal.com"
	paypalSandbox = "https://api-m.sandbox.paypal.com"
)

type PayPalProvider struct {
	http         *http.Client
	clientID     string
	clientSecret string
	webhookID    string
	base         string
}

func newPayPalProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	base := cfgStr(cfg, "base_url")
	if base == "" {
		base = paypalBase
	}
	return &PayPalProvider{http: &http.Client{Timeout: 10 * time.Second},
		clientID: cfgStr(cfg, "client_id"), clientSecret: cfgStr(cfg, "client_secret"),
		webhookID: cfgStr(cfg, "webhook_id"), base: base}, nil
}

func (p *PayPalProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.clientID == "" || p.clientSecret == "" {
		return "", fmt.Errorf("paypal: %w", pkg.ErrProviderOn)
	}
	token, err := p.accessToken(ctx)
	if err != nil {
		return "", err
	}
	body := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{{
			"reference_id": orderNo,
			"description":  desc,
			"amount": map[string]string{
				"currency_code": currency,
				"value":         strconv.FormatFloat(float64(amountCents)/100, 'f', -1, 64), // 最短表示：3 / 2.5 / 2.99
			},
		}},
	}
	var resp struct {
		ID    string `json:"id"`
		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
	}
	headers := map[string]string{"Authorization": "Bearer " + token}
	if err := doJSON(ctx, p.http, http.MethodPost, p.base+"/v2/checkout/orders", headers, body, &resp); err != nil {
		return "", err
	}
	for _, l := range resp.Links {
		if l.Rel == "approve" && l.Href != "" {
			return l.Href, nil
		}
	}
	return "", errors.New("paypal: no approve link in response")
}

// accessToken OAuth2 client_credentials 换 token（表单体，不走 doJSON）。
func (p *PayPalProvider) accessToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.base+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", basicAuthHeader(p.clientID, p.clientSecret))
	resp, err := p.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		s := string(b)
		if len(s) > 200 {
			s = s[:200]
		}
		return "", fmt.Errorf("paypal oauth: %s", s)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.AccessToken == "" {
		return "", errors.New("paypal: empty access token")
	}
	return out.AccessToken, nil
}

// paypalEvent 是 webhook payload（checkout.order.approved 等）。
type paypalEvent struct {
	EventType string `json:"event_type"`
	Resource  struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PurchaseUnits []struct {
			ReferenceID string `json:"reference_id"`
			Amount      struct {
				CurrencyCode string `json:"currency_code"`
				Value        string `json:"value"`
			} `json:"amount"`
		} `json:"purchase_units"`
	} `json:"resource"`
}

func (p *PayPalProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.webhookID == "" {
		return PaymentEvent{}, errors.New("paypal: webhook id not configured")
	}
	// 快速失败：transmission 头缺失 / webhook-id 与配置不匹配（本地即拒，不必打 PayPal）。
	if headers["PayPal-Transmission-Id"] == "" || headers["PayPal-Transmission-Time"] == "" ||
		headers["PayPal-Transmission-Sig"] == "" || headers["PayPal-Auth-Algo"] == "" || headers["PayPal-Cert-Url"] == "" {
		return PaymentEvent{}, errors.New("paypal: transmission headers missing")
	}
	if !hmac.Equal([]byte(headers["PayPal-Webhook-Id"]), []byte(p.webhookID)) {
		return PaymentEvent{}, errors.New("paypal: webhook id mismatch")
	}
	// 官方验签：verify-webhook-signature 接口复验（cert_url/auth_algo 与签名由 PayPal 侧校验，
	// 比本地下载 cert + RSA 摘要比对更可靠）。正式上线前需用真实沙箱 webhook 验证一次。
	if err := p.verifyWebhookSignature(ctx, rawBody, headers); err != nil {
		return PaymentEvent{}, err
	}
	var evt paypalEvent
	if err := json.Unmarshal(rawBody, &evt); err != nil {
		return PaymentEvent{}, errors.New("paypal: parse event failed")
	}
	if evt.EventType != "CHECKOUT.ORDER.APPROVED" {
		return PaymentEvent{}, nil // 其他事件（含 capture）不结算
	}
	if len(evt.Resource.PurchaseUnits) == 0 {
		return PaymentEvent{}, errors.New("paypal: no purchase_units in event")
	}
	pu := evt.Resource.PurchaseUnits[0]
	return PaymentEvent{OrderNo: pu.ReferenceID, TxID: evt.Resource.ID, Paid: true,
		AmountCents: moneyToCents(pu.Amount.Value), Currency: pu.Amount.CurrencyCode}, nil
}

// verifyWebhookSignature 调 PayPal 官方验签接口（OAuth access_token 复用，与 CreateCheckout 一致）。
func (p *PayPalProvider) verifyWebhookSignature(ctx context.Context, rawBody []byte, headers map[string]string) error {
	if p.clientID == "" || p.clientSecret == "" {
		return errors.New("paypal: credentials not configured")
	}
	token, err := p.accessToken(ctx)
	if err != nil {
		return err
	}
	body := map[string]any{
		"auth_algo":         headers["PayPal-Auth-Algo"],
		"cert_url":          headers["PayPal-Cert-Url"],
		"transmission_id":   headers["PayPal-Transmission-Id"],
		"transmission_sig":  headers["PayPal-Transmission-Sig"],
		"transmission_time": headers["PayPal-Transmission-Time"],
		"webhook_id":        headers["PayPal-Webhook-Id"],
		"webhook_event":     json.RawMessage(rawBody),
	}
	var resp struct {
		VerificationStatus string `json:"verification_status"`
	}
	if err := doJSON(ctx, p.http, http.MethodPost, p.base+"/v1/notifications/verify-webhook-signature",
		map[string]string{"Authorization": "Bearer " + token}, body, &resp); err != nil {
		return err
	}
	if resp.VerificationStatus != "SUCCESS" {
		return fmt.Errorf("paypal: verify webhook signature: %s", resp.VerificationStatus)
	}
	return nil
}
