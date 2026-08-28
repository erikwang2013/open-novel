package biz

// Stripe 渠道（T-P-04）：Checkout Session 创建支付链接，webhook 用官方
// ConstructEvent 验签。密钥缺省时返回业务错误，不 panic。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	stripe "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/pkg"
)

type StripeProvider struct {
	secretKey  string
	webhookSec string
}

func newStripeProvider(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error) {
	// 密钥可空：CreateCheckout 需要 secret key，VerifyWebhook 需要 webhook secret，
	// 各自在使用时校验（webhook 验签不依赖 API key）。
	return &StripeProvider{secretKey: pay.StripeSecretKey, webhookSec: pay.StripeWebhookSecret}, nil
}

func (p *StripeProvider) CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (string, error) {
	if p.secretKey == "" {
		return "", fmt.Errorf("stripe: %w", pkg.ErrProviderOn)
	}
	stripe.Key = p.secretKey
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{
			Quantity: stripe.Int64(1),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String(strings.ToLower(currency)),
				UnitAmount: stripe.Int64(amountCents),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(desc),
				},
			},
		}},
		Metadata: map[string]string{"order_no": orderNo},
	}
	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe create checkout: %w", err)
	}
	if s.URL == "" {
		return "", errors.New("stripe: empty checkout url")
	}
	return s.URL, nil
}

func (p *StripeProvider) VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error) {
	if p.webhookSec == "" {
		return PaymentEvent{}, errors.New("stripe: webhook secret not configured")
	}
	event, err := webhook.ConstructEvent(rawBody, headers["Stripe-Signature"], p.webhookSec)
	if err != nil {
		return PaymentEvent{}, errors.New("stripe: webhook signature invalid")
	}
	if event.Type != stripe.EventTypeCheckoutSessionCompleted {
		return PaymentEvent{}, nil
	}
	var s stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
		return PaymentEvent{}, errors.New("stripe: parse session failed")
	}
	txID := ""
	if s.PaymentIntent != nil {
		txID = s.PaymentIntent.ID
	}
	return PaymentEvent{OrderNo: s.Metadata["order_no"], TxID: txID, Paid: true,
		AmountCents: s.AmountTotal, Currency: strings.ToUpper(string(s.Currency))}, nil
}
