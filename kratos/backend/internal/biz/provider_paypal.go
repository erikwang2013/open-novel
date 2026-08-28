package biz

// PayPal 渠道（T-P-20，兜底）：Orders API v2——OAuth2 换 token → POST /v2/checkout/orders
// → 取 approve link 跳转；webhook 为简化验签（transmission 头存在 + webhook-id token 比对 +
// 金额核对），见 VerifyWebhook 内 ponytail 注释。
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
	// ponytail: 官方 Webhooks 验签需按 PayPal-Transmission-Id|Time|Sig + PayPal-Webhook-Id
	// 下载 cert（PayPal-Cert-Url）→ 提取公钥 → RSA-SHA256 摘要比对，并建议 GET
	// /v2/checkout/events/{id} 复核。此处简化：transmission 头存在 + webhook-id 与配置
	// 比对 + 金额核对。升级路径：实现 cert 下载与 crypto/rsa VerifyPKCS1v15(sha256) 完整验签。
	if headers["PayPal-Transmission-Id"] == "" || headers["PayPal-Transmission-Time"] == "" ||
		headers["PayPal-Transmission-Sig"] == "" {
		return PaymentEvent{}, errors.New("paypal: transmission headers missing")
	}
	if !hmac.Equal([]byte(headers["PayPal-Webhook-Id"]), []byte(p.webhookID)) {
		return PaymentEvent{}, errors.New("paypal: webhook id mismatch")
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
