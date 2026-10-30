package biz

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// TestListPublicPlansCurrency: 公开套餐只返回 lang 路由到的渠道币种，禁用套餐不返回，不混列多币种。
// 需本地 MySQL(3307)；语言码用一次性 marker，避免与存量渠道路由冲突。
func TestListPublicPlansCurrency(t *testing.T) {
	pc, d := newPaymentTestData(t)
	ctx := context.Background()

	marker := "lp" + strconv.FormatInt(time.Now().UnixNano()%1e10, 10) // 12 字符，lang 列限 16
	t.Cleanup(func() {
		d.DB.Where("plan_code LIKE ?", marker+"%").Delete(&data.VipPlan{})
		d.DB.Where("code LIKE ?", marker+"%").Delete(&data.PaymentProvider{})
	})

	// sort 取最小值，确保 pickProvider 在存在全局渠道时仍选中本用例的 marker 渠道。
	addProvider := func(suffix, lang, currency string) {
		t.Helper()
		enc, err := pc.encryptConfig(map[string]string{"currency": currency})
		if err != nil {
			t.Fatal(err)
		}
		p := data.PaymentProvider{Code: marker + suffix, Lang: lang, Region: "*",
			Enabled: 1, Sort: math.MinInt32, Config: enc}
		if err := d.DB.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
	}
	addProvider("_cny", marker+"zh", "CNY")
	addProvider("_usd", marker+"en", "USD")

	plans := []data.VipPlan{
		{PlanCode: marker + "_cny1", Days: 30, AmountCents: 1000, Currency: "CNY", Sort: 1, Status: 1},
		{PlanCode: marker + "_usd1", Days: 30, AmountCents: 500, Currency: "USD", Sort: 1, Status: 1},
		{PlanCode: marker + "_usd2", Days: 365, AmountCents: 5000, Currency: "USD", Sort: 2, Status: 1},
		{PlanCode: marker + "_usd3", Days: 90, AmountCents: 2000, Currency: "USD", Sort: 3, Status: 0},
	}
	if err := d.DB.Create(&plans).Error; err != nil {
		t.Fatal(err)
	}

	// 只断言本用例 marker 行，避免库中已有正式套餐时误判。
	mine := func(rows []data.VipPlan) []data.VipPlan {
		var out []data.VipPlan
		for _, p := range rows {
			if strings.HasPrefix(p.PlanCode, marker) {
				out = append(out, p)
			}
		}
		return out
	}

	cny, err := pc.ListPublicPlans(ctx, marker+"zh")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range cny {
		if p.Currency != "CNY" {
			t.Fatalf("CNY 渠道混入 %s 套餐", p.Currency)
		}
	}
	if got := mine(cny); len(got) != 1 || got[0].PlanCode != marker+"_cny1" {
		t.Fatalf("CNY 渠道应只返回本用例 1 条启用套餐，got %+v", got)
	}

	usd, err := pc.ListPublicPlans(ctx, marker+"en")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range usd {
		if p.Currency != "USD" {
			t.Fatalf("USD 渠道混入 %s 套餐", p.Currency)
		}
	}
	if got := mine(usd); len(got) != 2 || got[0].PlanCode != marker+"_usd1" || got[1].PlanCode != marker+"_usd2" {
		t.Fatalf("USD 渠道应只返回本用例 2 条启用套餐且按 sort 升序，got %+v", got)
	}
}

// TestPublicPlansFallback: USD 无套餐行回退内置三档；非 USD 无行仍为空；有行时不回退。
func TestPublicPlansFallback(t *testing.T) {
	usd := publicPlansFallback(nil, "USD")
	if len(usd) != 3 {
		t.Fatalf("USD 无套餐行应回退 3 档，got %d", len(usd))
	}
	for i, code := range planOrder {
		p := usd[i]
		if p.PlanCode != code || p.Days != planDays[code] || p.AmountCents != defaultPlans[code] ||
			p.Currency != "USD" || p.Status != 1 || p.Sort != i {
			t.Fatalf("默认套餐 %d 字段不符: %+v", i, p)
		}
	}
	if got := publicPlansFallback(nil, "CNY"); len(got) != 0 {
		t.Fatalf("非 USD 无套餐行应为空，got %+v", got)
	}
	actual := []data.VipPlan{{PlanCode: "monthly", AmountCents: 999, Currency: "USD", Status: 1}}
	if got := publicPlansFallback(actual, "USD"); len(got) != 1 || got[0].AmountCents != 999 {
		t.Fatalf("USD 有套餐行时不应回退，got %+v", got)
	}
}

func newPaymentTestData(t *testing.T) (*PaymentUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })
	return NewPaymentUsecase(d, &conf.Payment{}), d
}

// TestPlanUniquePrice: 同币种同价拒绝创建/更新；不同币种同价允许；禁用态不参与冲突。
func TestPlanUniquePrice(t *testing.T) {
	pc, d := newPaymentTestData(t)
	ctx := context.Background()
	const cur, other = "XQZ", "XQY"
	clean := func() { d.DB.Where("currency IN ?", []string{cur, other}).Delete(&data.VipPlan{}) }
	clean()
	t.Cleanup(clean)

	mk := func(code string, amount int64, currency string) (*data.VipPlan, error) {
		return pc.CreatePlan(ctx, 0, PlanUpsert{PlanCode: code, Days: 30, AmountCents: amount, Currency: currency})
	}
	m, err := mk("monthly", 11111, cur)
	if err != nil {
		t.Fatal(err)
	}
	// 同币种同价（不同 plan_code）拒绝
	if _, err := mk("quarterly", 11111, cur); err != pkg.ErrConflict {
		t.Fatalf("同币种同价创建应冲突，got %v", err)
	}
	// 不同币种同价允许
	if _, err := mk("quarterly", 11111, other); err != nil {
		t.Fatalf("不同币种同价应允许: %v", err)
	}
	// 同币种不同价允许
	y, err := mk("yearly", 22222, cur)
	if err != nil {
		t.Fatal(err)
	}
	// 更新成同币种同价 → 拒绝
	amt := int64(11111)
	if _, err := pc.UpdatePlan(ctx, 0, y.ID, PlanUpdate{Amount: &amt}); err != pkg.ErrConflict {
		t.Fatalf("同币种同价更新应冲突，got %v", err)
	}
	// 更新自身其它字段（排除自身）允许
	label := "test"
	if _, err := pc.UpdatePlan(ctx, 0, m.ID, PlanUpdate{Label: &label}); err != nil {
		t.Fatalf("更新自身应允许: %v", err)
	}
	// 同时置禁用则跳过冲突校验
	off := int8(0)
	if _, err := pc.UpdatePlan(ctx, 0, y.ID, PlanUpdate{Amount: &amt, Status: &off}); err != nil {
		t.Fatalf("禁用态更新应允许: %v", err)
	}
}

type stubPayProvider struct{}

func (stubPayProvider) CreateCheckout(context.Context, string, int64, string, string) (string, error) {
	return "https://pay.test/checkout", nil
}

func (stubPayProvider) VerifyWebhook(context.Context, string, []byte, map[string]string) (PaymentEvent, error) {
	return PaymentEvent{}, nil
}

// TestCreateOrderPlanIdempotency: 复用键含 plan；plan=” 的老订单不复用，不同 plan 不互相复用。
func TestCreateOrderPlanIdempotency(t *testing.T) {
	pc, d := newPaymentTestData(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano() % 1e10
	marker := "lpo" + strconv.FormatInt(suffix, 10)
	uid := uint64(suffix) + 1000000
	const cur = "XQO"
	code := marker + "_stub"
	t.Cleanup(func() {
		d.DB.Where("user_id = ?", uid).Delete(&data.PaymentOrder{})
		d.DB.Where("currency = ?", cur).Delete(&data.VipPlan{})
		d.DB.Where("code = ?", code).Delete(&data.PaymentProvider{})
		delete(providerFactories, code)
	})
	d.DB.Where("currency = ?", cur).Delete(&data.VipPlan{})
	providerFactories[code] = func(map[string]any, *conf.Payment) (PaymentProvider, error) {
		return stubPayProvider{}, nil
	}
	enc, err := pc.encryptConfig(map[string]string{"currency": cur})
	if err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create(&data.PaymentProvider{Code: code, Lang: marker, Region: "*",
		Enabled: 1, Sort: math.MinInt32, Config: enc}).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create([]data.VipPlan{
		{PlanCode: "monthly", Days: 30, AmountCents: 111, Currency: cur, Sort: 1, Status: 1},
		{PlanCode: "quarterly", Days: 90, AmountCents: 222, Currency: cur, Sort: 2, Status: 1},
	}).Error; err != nil {
		t.Fatal(err)
	}

	// 旧订单 plan=''：同 user/provider/currency/amount 但不得复用
	stale := data.PaymentOrder{OrderNo: marker + "_stale", UserID: uid, Amount: 1.11,
		Currency: cur, Provider: code, Plan: "", Status: 0}
	if err := d.DB.Create(&stale).Error; err != nil {
		t.Fatal(err)
	}

	first, err := pc.CreateOrder(ctx, uid, "monthly", marker)
	if err != nil {
		t.Fatal(err)
	}
	if first.OrderNo == stale.OrderNo || first.AmountCents != 111 {
		t.Fatalf("不得复用 plan='' 老订单: %+v", first)
	}
	again, err := pc.CreateOrder(ctx, uid, "monthly", marker)
	if err != nil {
		t.Fatal(err)
	}
	if again.OrderNo != first.OrderNo {
		t.Fatalf("同 plan 应复用订单: %s vs %s", again.OrderNo, first.OrderNo)
	}
	other, err := pc.CreateOrder(ctx, uid, "quarterly", marker)
	if err != nil {
		t.Fatal(err)
	}
	if other.OrderNo == first.OrderNo || other.AmountCents != 222 {
		t.Fatalf("不同 plan 不应复用: %+v", other)
	}
}

// TestWebhookSettlesStoredPlan: 回调按订单快照 plan/days 结算，不按当前套餐表金额反查。
// 快照 days=45 与内置默认 monthly（30 天）不同：若回退金额反查，只会续 30 天。
func TestWebhookSettlesStoredPlan(t *testing.T) {
	pc, d := newPaymentTestData(t)
	pc.payCfg = &conf.Payment{EncryptKey: "test-key", NpIpnSecret: "ipn-secret"}
	ctx := context.Background()
	suffix := time.Now().UnixNano() % 1e10
	marker := "wh" + strconv.FormatInt(suffix, 10)
	orderNo := marker + "_order"

	u := data.User{Username: marker, PasswordHash: "x", Nickname: "wh", Status: 1, Role: 1}
	if err := d.DB.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.Where("user_id = ?", u.ID).Delete(&data.VipOrder{})
		d.DB.Where("user_id = ?", u.ID).Delete(&data.PaymentOrder{})
		d.DB.Where("id = ?", u.ID).Delete(&data.User{})
	})

	o := data.PaymentOrder{OrderNo: orderNo, UserID: u.ID, Amount: 3.00, Currency: "USD",
		Provider: "np_usdt", Plan: "monthly", Days: 45, Status: 0}
	if err := d.DB.Create(&o).Error; err != nil {
		t.Fatal(err)
	}

	body := []byte(fmt.Sprintf(
		`{"payment_id":%d,"payment_status":"finished","order_id":%q,"price_amount":"3.00","price_currency":"USD","pay_amount":"3.00","pay_currency":"usdttrc20"}`,
		suffix, orderNo))
	if err := pc.Webhook(ctx, "np_usdt", body, npSigned(t, "ipn-secret", body)); err != nil {
		t.Fatal(err)
	}

	var got data.PaymentOrder
	if err := d.DB.First(&got, o.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Status != 1 {
		t.Fatalf("订单未结算: status=%d", got.Status)
	}
	var user data.User
	if err := d.DB.First(&user, u.ID).Error; err != nil {
		t.Fatal(err)
	}
	if user.VipExpiresAt == nil {
		t.Fatal("VIP 到期时间未写入")
	}
	if left := time.Until(*user.VipExpiresAt); left < 44*24*time.Hour || left > 46*24*time.Hour {
		t.Fatalf("VIP 应按快照续 45 天，剩余 %v", left)
	}
	var vo data.VipOrder
	if err := d.DB.Where("order_no = ?", orderNo).First(&vo).Error; err != nil {
		t.Fatal(err)
	}
	if vo.Plan != "monthly" || vo.Status != 1 || vo.Currency != "USD" || vo.Amount != 3.00 {
		t.Fatalf("会员订单未按快照写入: %+v", vo)
	}
	if vo.StartAt == nil || vo.EndAt == nil {
		t.Fatal("会员订单起止时间缺失")
	}
	if dur := vo.EndAt.Sub(*vo.StartAt); dur < 45*24*time.Hour-time.Hour || dur > 45*24*time.Hour+time.Hour {
		t.Fatalf("会员订单应按快照 45 天，实际 %v", dur)
	}
}
