package biz

// 支付核心单测：NP HMAC-SHA512 验签向量、Stripe ConstructEvent、金额反查、
// 15min 超时关闭。DB 相关路径依赖真实库，仅测纯逻辑。

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	stripe "github.com/stripe/stripe-go/v81"

	"open-novel/backend/internal/data"
)

// npSigned 构造本地 IPN 测试向量：payload + HMAC-SHA512 签名头。
func npSigned(t *testing.T, secret string, payload []byte) map[string]string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(payload)
	return map[string]string{"X-Nowpayments-Sig": hex.EncodeToString(mac.Sum(nil))}
}

func TestNPVerifyWebhook(t *testing.T) {
	p := &NPProvider{ipnSec: "ipn-test-secret"}
	ctx := context.Background()
	body := []byte(`{"payment_id":123,"payment_status":"finished","order_id":"2026082901","price_amount":"3.00","price_currency":"USD","pay_amount":"3.00000000","actually_paid":"3.00000000","pay_currency":"usdttrc20"}`)

	ev, err := p.VerifyWebhook(ctx, "np_usdt", body, npSigned(t, "ipn-test-secret", body))
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082901" || ev.AmountCents != 300 || ev.TxID != "123" || ev.Currency != "USD" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	// 无 price_amount（旧版 IPN）时退回 pay_amount 计价
	noPrice := []byte(`{"payment_id":124,"payment_status":"finished","order_id":"2026082902","pay_amount":"2.99","pay_currency":"usdttrc20"}`)
	ev2, err := p.VerifyWebhook(ctx, "np_usdt", noPrice, npSigned(t, "ipn-test-secret", noPrice))
	if err != nil {
		t.Fatal(err)
	}
	if ev2.AmountCents != 299 {
		t.Fatalf("fallback pay_amount: got %d want 299", ev2.AmountCents)
	}

	// 篡改 payload → 验签失败
	tampered := append([]byte{}, body...)
	tampered = append(tampered, []byte(" ")...)
	if _, err := p.VerifyWebhook(ctx, "np_usdt", tampered, npSigned(t, "ipn-test-secret", body)); err == nil {
		t.Fatal("tampered payload must fail signature check")
	}
	// 错误密钥 → 失败
	if _, err := p.VerifyWebhook(ctx, "np_usdt", body, npSigned(t, "wrong-secret", body)); err == nil {
		t.Fatal("wrong secret must fail")
	}
}

func TestNPVerifyPending(t *testing.T) {
	p := &NPProvider{ipnSec: "ipn-test-secret"}
	body := []byte(`{"payment_id":9,"payment_status":"waiting","order_id":"o1","pay_amount":"3.00","pay_currency":"usdttrc20"}`)
	ev, err := p.VerifyWebhook(context.Background(), "np_usdt", body, npSigned(t, "ipn-test-secret", body))
	if err != nil {
		t.Fatal(err)
	}
	if ev.Paid {
		t.Fatal("waiting status must not be paid")
	}
}

// stripeSigned 用与 Stripe 相同的 HMAC-SHA256 方案本地构造签名头（测试向量）。
func stripeSigned(t *testing.T, secret string, ts int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{int64Str(ts), string(payload)}, ".")))
	return "t=" + int64Str(ts) + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func int64Str(v int64) string { return strconv.FormatInt(v, 10) }

func TestStripeVerifyWebhook(t *testing.T) {
	p := &StripeProvider{webhookSec: "whsec_test_secret"}
	body := []byte(fmt.Sprintf(`{"id":"evt_1","api_version":%q,"type":"checkout.session.completed","data":{"object":{"id":"cs_1","metadata":{"order_no":"2026082901"},"amount_total":300,"currency":"usd","payment_intent":{"id":"pi_1"}}}}`, stripe.APIVersion))
	hdr := map[string]string{"Stripe-Signature": stripeSigned(t, "whsec_test_secret", time.Now().Unix(), body)}
	ev, err := p.VerifyWebhook(context.Background(), "stripe", body, hdr)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.Paid || ev.OrderNo != "2026082901" || ev.AmountCents != 300 || ev.TxID != "pi_1" || ev.Currency != "USD" {
		t.Fatalf("unexpected event: %+v", ev)
	}

	// 篡改 payload → 签名失败
	if _, err := p.VerifyWebhook(context.Background(), "stripe", append(body, ' '), hdr); err == nil {
		t.Fatal("tampered payload must fail")
	}
	// 非完成事件 → Paid=false 不报错
	pend := []byte(fmt.Sprintf(`{"id":"evt_2","api_version":%q,"type":"checkout.session.expired","data":{"object":{}}}`, stripe.APIVersion))
	hdr2 := map[string]string{"Stripe-Signature": stripeSigned(t, "whsec_test_secret", time.Now().Unix(), pend)}
	ev, err = p.VerifyWebhook(context.Background(), "stripe", pend, hdr2)
	if err != nil || ev.Paid {
		t.Fatalf("non-completed event: err=%v ev=%+v", err, ev)
	}
}

func TestPlanForPrice(t *testing.T) {
	// 默认金额反查
	if p, d, ok := planForPrice(map[string]any{}, 800); !ok || p != "quarterly" || d != 90 {
		t.Fatalf("quarterly default: %v %v %v", p, d, ok)
	}
	// provider 覆盖金额后反查仍一一对应
	cfg := map[string]any{"plans": map[string]any{"monthly": float64(999)}}
	if p, _, ok := planForPrice(cfg, 999); !ok || p != "monthly" {
		t.Fatalf("override monthly: %v %v", p, ok)
	}
	// 未知金额 → 不命中
	if _, _, ok := planForPrice(map[string]any{}, 1); ok {
		t.Fatal("unknown amount must not map to a plan")
	}
	// 配置金额冲突 → 回退默认值，且反查确定性（固定顺序 monthly 优先）
	collide := map[string]any{"plans": map[string]any{"monthly": float64(500), "quarterly": float64(500), "yearly": float64(3000)}}
	amounts := planAmounts(collide)
	if amounts["monthly"] != 300 || amounts["quarterly"] != 800 {
		t.Fatalf("collision must fall back to defaults: %+v", amounts)
	}
	if p, d, ok := planForPrice(collide, 800); !ok || p != "quarterly" || d != 90 {
		t.Fatalf("collision reverse lookup: %v %v %v", p, d, ok)
	}
	for i := 0; i < 20; i++ { // 多次反查结果必须一致
		p, _, _ := planForPrice(collide, 300)
		if p != "monthly" {
			t.Fatalf("non-deterministic reverse lookup: %v", p)
		}
	}
}

func TestAmountOK(t *testing.T) {
	if !amountOK(300, 300) || !amountOK(300, 301) || !amountOK(300, 299) {
		t.Fatal("within ±1 cent must pass")
	}
	if amountOK(300, 302) || amountOK(300, 298) {
		t.Fatal("beyond ±1 cent must fail")
	}
}

func TestCentsOf(t *testing.T) {
	if centsOf(3.0) != 300 || centsOf(3.145) != 315 || centsOf(0.1) != 10 {
		t.Fatal("centsOf rounding wrong")
	}
}

func TestMaybeClose(t *testing.T) {
	now := time.Now()
	fresh := &data.PaymentOrder{Status: 0, CreatedAt: now}
	if maybeClose(fresh, now.Add(14*time.Minute)) {
		t.Fatal("14min pending must not close")
	}
	stale := &data.PaymentOrder{Status: 0, CreatedAt: now}
	if !maybeClose(stale, now.Add(16*time.Minute)) {
		t.Fatal("16min pending must close")
	}
	paid := &data.PaymentOrder{Status: 1, CreatedAt: now.Add(-time.Hour)}
	if maybeClose(paid, now) {
		t.Fatal("paid order must not close")
	}
}
