package biz

// NOWPayments 渠道（T-P-05）：REST 直调创建 USDT 支付，IPN webhook 用
// HMAC-SHA512(ipn_secret, rawBody) 恒定时间验签（无官方 SDK）。

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

type NPProvider struct {
	http   *http.Client
	apiKey string
	ipnSec string
	coin   string // usdttrc20 / usdterc20 / usdtbsc20
}

func newNPProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	// apiKey 可空：CreateCheckout 使用时校验；IPN 验签只需 ipnSec。
	coin, _ := cfg["coin"].(string)
	if coin == "" {
		coin = "usdttrc20"
	}
	return &NPProvider{http: &http.Client{Timeout: 10 * time.Second},
		apiKey: pay.NpApiKey, ipnSec: pay.NpIpnSecret, coin: coin}, nil
}

type npCreateReq struct {
	PriceAmount      string `json:"price_amount"`
	PriceCurrency    string `json:"price_currency"`
	PayCurrency      string `json:"pay_currency"`
	OrderID          string `json:"order_id"`
	OrderDescription string `json:"order_description"`
	IPNCallbackURL   string `json:"ipn_callback_url"`
}

type npCreateResp struct {
	PaymentID  int64  `json:"payment_id"`
	PaymentURL string `json:"payment_url"`
	PaymentStatus string `json:"payment_status"`
}

func (p *NPProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("np: %w", pkg.ErrProviderOn)
	}
	body := npCreateReq{
		PriceAmount:      strconv.FormatFloat(float64(amountCents)/100, 'f', 2, 64),
		PriceCurrency:    currency,
		PayCurrency:      p.coin,
		OrderID:          orderNo,
		OrderDescription: desc,
	}
	var resp npCreateResp
	if err := p.do(ctx, http.MethodPost, "/v1/payment", body, &resp, true); err != nil {
		return "", err
	}
	if resp.PaymentURL == "" {
		return "", errors.New("np: empty payment url")
	}
	return resp.PaymentURL, nil
}

// npIpn 是 IPN webhook 的 payload 字段（透传 json.RawMessage 防重排）。
type npIpn struct {
	PaymentID    int64           `json:"payment_id"`
	PaymentStatus string         `json:"payment_status"`
	OrderID      string          `json:"order_id"`
	PriceAmount  json.RawMessage `json:"price_amount"`   // 计价币种金额（下单时按订单币种设定）
	PriceCurrency string         `json:"price_currency"` // 如 USD
	PayAmount    json.RawMessage `json:"pay_amount"`     // 支付币种金额（USDT 8 位小数）
	ActuallyPaid json.RawMessage `json:"actually_paid"`
	PayCurrency  string          `json:"pay_currency"`
}

func (p *NPProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.ipnSec == "" {
		return PaymentEvent{}, errors.New("np: ipn secret not configured")
	}
	sig, _ := hex.DecodeString(headers["X-Nowpayments-Sig"])
	mac := hmac.New(sha512.New, []byte(p.ipnSec))
	mac.Write(rawBody)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return PaymentEvent{}, errors.New("np: webhook signature invalid")
	}
	var ipn npIpn
	if err := json.Unmarshal(rawBody, &ipn); err != nil {
		return PaymentEvent{}, errors.New("np: parse ipn failed")
	}
	ev := PaymentEvent{OrderNo: ipn.OrderID, TxID: strconv.FormatInt(ipn.PaymentID, 10),
		Paid: ipn.PaymentStatus == "finished", Currency: "USD"}
	if ipn.PaymentStatus == "finished" {
		// 金额以计价币种（price_amount，与订单币种一致）为准；缺失时退回 pay_amount。
		// 支付侧舍入/折算误差由 biz 的 amountOK（±1 分容差）兜底。
		ev.AmountCents = npCents(ipn.PriceAmount)
		if ev.AmountCents <= 0 {
			ev.AmountCents = npCents(ipn.PayAmount)
		}
	}
	return ev, nil
}

// do 直调 NOWPayments REST；auth=true 时带 x-api-key。
func (p *NPProvider) do(ctx context.Context, method, path string, body any, out any, auth bool) error {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.nowpayments.io"+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth {
		req.Header.Set("x-api-key", p.apiKey)
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		s := string(b)
		if len(s) > 200 {
			s = s[:200]
		}
		return fmt.Errorf("np %s %s: %s", method, path, s)
	}
	if out != nil && len(b) > 0 {
		return json.Unmarshal(b, out)
	}
	return nil
}

// npCents 解析 NP 的金额字段（数字或字符串）为整数分，禁止浮点比较。
func npCents(raw json.RawMessage) int64 {
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int64(math.Round(f * 100))
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if f, err = strconv.ParseFloat(s, 64); err == nil {
			return int64(math.Round(f * 100))
		}
	}
	return 0
}
