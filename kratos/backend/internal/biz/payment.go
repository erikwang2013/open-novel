package biz

// 支付用例（T-P-03~08）：下单 / 查单 / 支付方式路由 / webhook 验签 + VIP 激活。
// 金额一律整数分比较；读订单一律 FORCE_MASTER（支付结果一致性优先）。

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// 内置默认：套餐天数与金额（分）。provider.config 可覆盖 plans 金额。
var (
	planDays     = map[string]int{"monthly": 30, "quarterly": 90, "yearly": 365}
	defaultPlans = map[string]int64{"monthly": 300, "quarterly": 800, "yearly": 3000}
)

const pendingCloseAfter = 15 * time.Minute

// PaymentEvent 是渠道回调解析出的统一支付事件。
type PaymentEvent struct {
	OrderNo     string
	TxID        string
	Paid        bool
	AmountCents int64
	Currency    string
}

// PaymentProvider 是支付渠道抽象：下单创建跳转 URL，webhook 验签解析事件。
type PaymentProvider interface {
	CreateCheckout(ctx context.Context, orderNo string, amountCents int64, currency, desc string) (checkoutURL string, err error)
	VerifyWebhook(ctx context.Context, provider string, rawBody []byte, headers map[string]string) (PaymentEvent, error)
}

// providerFactory 以解密后的渠道配置实例化 Provider。
type providerFactory func(cfg map[string]any, pay *conf.Payment) (PaymentProvider, error)

var providerFactories = map[string]providerFactory{
	"stripe":  newStripeProvider,
	"np_usdt": newNPProvider,
}

// webhookAlias 把 webhook 路径参数（stripe/nowpayments）归一为渠道码。
func webhookAlias(name string) string {
	if name == "nowpayments" {
		return "np_usdt"
	}
	return name
}

type PaymentUsecase struct {
	db     *gorm.DB
	cr     *pkg.Crypto
	payCfg *conf.Payment
}

func NewPaymentUsecase(d *data.Data, c *conf.Payment) *PaymentUsecase {
	key := c.EncryptKey
	if key == "" {
		key = "dev-encrypt-key-change-me" // 与 config.yaml 默认一致
	}
	cr, _ := pkg.NewCrypto(key) // key 非空必然成功
	return &PaymentUsecase{db: d.DB, cr: cr, payCfg: c}
}

type CreatedOrder struct {
	OrderNo     string
	AmountCents int64
	Currency    string
	CheckoutURL string
	Provider    string
}

type OrderInfo struct {
	OrderNo     string
	Status      int8
	AmountCents int64
	Currency    string
	Provider    string
	TxID        string
	PaidAt      *time.Time
}

type MethodInfo struct {
	Code   string
	Lang   string
	Region string
	Sort   int32
}

// CreateOrder 创建 VIP 订单：plan→(天数,金额)；按 lang 路由支付方式；幂等复用未支付订单。
func (pc *PaymentUsecase) CreateOrder(ctx context.Context, uid uint64, plan, lang string) (CreatedOrder, error) {
	if _, ok := planDays[plan]; !ok {
		return CreatedOrder{}, pkg.ErrInvalidArgument
	}
	rows, err := pc.enabledProviders(ctx)
	if err != nil {
		return CreatedOrder{}, pkg.ErrPayCreate
	}
	row := pickProvider(rows, lang)
	if row == nil {
		return CreatedOrder{}, pkg.ErrProviderOn
	}
	cfg, err := pc.decryptConfig(row.Config)
	if err != nil {
		return CreatedOrder{}, pkg.ErrProviderOn
	}
	currency := "USD"
	if v, _ := cfg["currency"].(string); v != "" {
		currency = v
	}
	amounts, _ := pc.effectivePlans(ctx, currency)
	amountCents := amounts[plan] // 金额↔plan 一一对应由 effectivePlans 保证

	// 幂等：同 user+provider+amount 的未支付订单直接复用。
	// amount↔plan 一一对应（planAmounts 强制金额唯一），等价于 (user_id, plan, status=0)。
	var o data.PaymentOrder
	// FORCE_MASTER: 刚创建的订单立即可见
	err = pc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("user_id = ? AND provider = ? AND amount = ? AND status = 0", uid, row.Code, float64(amountCents)/100).
		Order("id DESC").First(&o).Error
	reuse := err == nil
	if err != nil && err != gorm.ErrRecordNotFound {
		return CreatedOrder{}, pkg.ErrPayCreate
	}
	if !reuse {
		o = data.PaymentOrder{
			OrderNo:  newOrderNo(), UserID: uid,
			Amount: float64(amountCents) / 100, Currency: currency,
			Provider: row.Code, Status: 0,
		}
		if err := pc.db.WithContext(ctx).Create(&o).Error; err != nil {
			return CreatedOrder{}, pkg.ErrPayCreate
		}
	}

	prov, err := providerFactories[row.Code](cfg, pc.payCfg)
	if err != nil {
		return CreatedOrder{}, pkg.ErrPayCreate
	}
	url, err := prov.CreateCheckout(ctx, o.OrderNo, amountCents, currency, "Open Novel VIP "+plan)
	if err != nil {
		return CreatedOrder{}, pkg.ErrPayCreate
	}
	return CreatedOrder{OrderNo: o.OrderNo, AmountCents: amountCents, Currency: currency,
		CheckoutURL: url, Provider: row.Code}, nil
}

// GetOrder 查单：仅本人；未支付超 15 分钟自动置为已关闭（3）。
func (pc *PaymentUsecase) GetOrder(ctx context.Context, uid uint64, orderNo string) (OrderInfo, error) {
	var o data.PaymentOrder
	// FORCE_MASTER: 支付回调后立即可见
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("order_no = ?", orderNo).First(&o).Error; err != nil {
		return OrderInfo{}, pkg.ErrOrderNF
	}
	if o.UserID != uid { // 不泄露他人订单存在性
		return OrderInfo{}, pkg.ErrOrderNF
	}
	if maybeClose(&o, time.Now()) {
		pc.db.WithContext(ctx).Model(&data.PaymentOrder{}).
			Where("id = ? AND status = 0", o.ID).Update("status", 3)
	}
	return orderInfo(o), nil
}

// ListPublicPlans 公开套餐列表（T-P-14）：仅启用，sort 升序。
// 公开读走默认连接（只读不敏感，无需 FORCE_MASTER）。
func (pc *PaymentUsecase) ListPublicPlans(ctx context.Context) ([]data.VipPlan, error) {
	var rows []data.VipPlan
	if err := pc.db.WithContext(ctx).Where("status = 1").Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	return rows, nil
}

// ListMethods 公开接口：enabled 且 lang/region 匹配，sort 升序。
func (pc *PaymentUsecase) ListMethods(ctx context.Context, lang string) ([]MethodInfo, error) {
	rows, err := pc.enabledProviders(ctx)
	if err != nil {
		return nil, pkg.ErrPayCreate
	}
	var out []MethodInfo
	for i := range rows {
		if !matchLangRegion(&rows[i], lang) {
			continue
		}
		out = append(out, MethodInfo{Code: rows[i].Code, Lang: rows[i].Lang,
			Region: rows[i].Region, Sort: int32(rows[i].Sort)})
	}
	return out, nil
}

// Webhook 渠道回调：验签→金额强校验→幂等→status 0→1 + VIP 激活。
func (pc *PaymentUsecase) Webhook(ctx context.Context, providerName string, rawBody []byte, headers map[string]string) error {
	code := webhookAlias(providerName)
	switch code { // 验签仅需 webhook 密钥，不依赖渠道 API key
	case "stripe":
		if pc.payCfg.StripeWebhookSecret == "" {
			return pkg.ErrProviderOn
		}
	case "np_usdt":
		if pc.payCfg.NpIpnSecret == "" {
			return pkg.ErrProviderOn
		}
	default:
		return pkg.ErrProviderOn
	}
	factory, ok := providerFactories[code]
	if !ok {
		return pkg.ErrProviderOn
	}
	prov, err := factory(nil, pc.payCfg)
	if err != nil {
		return pkg.ErrProviderOn
	}
	ev, err := prov.VerifyWebhook(ctx, code, rawBody, headers)
	if err != nil {
		return err
	}
	if !ev.Paid {
		return nil // 非支付成功事件（如 pending）直接 ack
	}
	if ev.OrderNo == "" {
		return pkg.ErrOrderNF
	}
	var o data.PaymentOrder
	// FORCE_MASTER: 回调路径必须读主库
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("order_no = ?", ev.OrderNo).First(&o).Error; err != nil {
		return pkg.ErrOrderNF
	}
	if o.Provider != code {
		return pkg.ErrOrderNF
	}
	if !amountOK(centsOf(o.Amount), ev.AmountCents) || !strings.EqualFold(o.Currency, ev.Currency) {
		return pkg.ErrAmountMism
	}
	if o.Status == 1 { // 幂等：已支付直接 ack
		return nil
	}
	// plan 由金额反查：下单与回调必须同一数据源（DB 套餐 + 内置默认）
	amounts, dayMap := pc.effectivePlans(ctx, o.Currency)
	plan, days, ok := planForAmounts(amounts, dayMap, ev.AmountCents)
	if !ok {
		return pkg.ErrAmountMism
	}
	if err := pc.settle(ctx, &o, ev, plan, days); err != nil {
		return err
	}
	return nil
}

// IsVipActive 供章节/书籍 VIP 校验使用：vip_expires_at > now。
func (pc *PaymentUsecase) IsVipActive(ctx context.Context, uid uint64) (bool, time.Time) {
	var u data.User
	if err := pc.db.WithContext(ctx).First(&u, uid).Error; err != nil {
		return false, time.Time{}
	}
	if u.VipExpiresAt == nil {
		return false, time.Time{}
	}
	return u.VipExpiresAt.After(time.Now()), *u.VipExpiresAt
}

// settle 事务内：订单置已支付 + VIP 激活（可叠加）+ 写会员订单。
// status IN (0,3)：已关闭（超时置 3）订单被真实付款时重新打开；
// RowsAffected==0 说明已被并发回调结算，幂等返回，不得再叠加 VIP/写会员单。
func (pc *PaymentUsecase) settle(ctx context.Context, o *data.PaymentOrder, ev PaymentEvent, plan string, days int) error {
	now := time.Now()
	return pc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&data.PaymentOrder{}).
			Where("id = ? AND status IN (0,3)", o.ID).
			Updates(map[string]any{"status": 1, "tx_id": ev.TxID, "paid_at": now})
		if res.Error != nil {
			return pkg.ErrPayCreate
		}
		if res.RowsAffected == 0 {
			return nil // 已结算（含重复回调）
		}
		var u data.User
		// FORCE_MASTER: 激活写后立即读
		if err := tx.Clauses(gormdb.Write).First(&u, o.UserID).Error; err != nil {
			return pkg.ErrOrderNF
		}
		base := now
		if u.VipExpiresAt != nil && u.VipExpiresAt.After(now) {
			base = *u.VipExpiresAt // 叠加
		}
		end := base.AddDate(0, 0, days)
		if err := tx.Model(&data.User{}).Where("id = ?", u.ID).
			Update("vip_expires_at", end).Error; err != nil {
			return pkg.ErrPayCreate
		}
		vo := data.VipOrder{OrderNo: o.OrderNo, UserID: o.UserID, Plan: plan,
			Amount: o.Amount, Currency: o.Currency, Status: 1,
			StartAt: &now, EndAt: &end, PaidAt: &now}
		if err := tx.Create(&vo).Error; err != nil {
			return pkg.ErrPayCreate
		}
		return nil
	})
}

func (pc *PaymentUsecase) enabledProviders(ctx context.Context) ([]data.PaymentProvider, error) {
	var rows []data.PaymentProvider
	// FORCE_MASTER: 配置刚改立即生效
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("enabled = 1").Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// decryptConfig 解密 provider.config（AES-GCM JSON）；空串返回空配置。
func (pc *PaymentUsecase) decryptConfig(enc string) (map[string]any, error) {
	if enc == "" {
		return map[string]any{}, nil
	}
	plain, err := pc.cr.Decrypt(enc)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(plain), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// pickProvider 按 lang 选默认方式：具体 region 匹配优先，再回退 '*'。
func pickProvider(rows []data.PaymentProvider, lang string) *data.PaymentProvider {
	var wildcard *data.PaymentProvider
	for i := range rows {
		if !matchLangRegion(&rows[i], lang) {
			continue
		}
		if rows[i].Region != "*" {
			return &rows[i]
		}
		if wildcard == nil {
			wildcard = &rows[i]
		}
	}
	return wildcard
}

// matchLangRegion：lang 字段匹配语言（'*' 或语言码），region 匹配地区（'*' 或语言对应地区）。
func matchLangRegion(row *data.PaymentProvider, lang string) bool {
	if row.Lang != "*" && !strings.EqualFold(row.Lang, langBase(lang)) {
		return false
	}
	if row.Region == "*" {
		return true
	}
	return strings.EqualFold(row.Region, langRegion(lang))
}

func langBase(lang string) string {
	l := pkg.NormalizeLang(lang)
	if i := strings.Index(l, "-"); i >= 0 {
		return l[:i]
	}
	return l
}

// langRegion："zh-CN"→CN；"en"→EN。
func langRegion(lang string) string {
	l := pkg.NormalizeLang(lang)
	if i := strings.Index(l, "-"); i >= 0 {
		return l[i+1:]
	}
	return l
}

// planOrder 固定顺序：金额反查与唯一性校验都按此遍历，保证确定性。
var planOrder = []string{"monthly", "quarterly", "yearly"}

// effectivePlans 生效套餐金额/天数（T-P-13）：DB novel_vip_plan 按 currency 匹配优先，
// 表空/无该币种行时回退内置默认；两套餐同价冲突时整体回退默认。
// 下单与 webhook 反查都经此取数，保证两端一致。
func (pc *PaymentUsecase) effectivePlans(ctx context.Context, currency string) (map[string]int64, map[string]int) {
	rows, err := pc.dbPlans(ctx, currency)
	if err != nil {
		return defaultPlans, planDays // 表不存在/查询失败 → 默认
	}
	return plansFromRows(rows)
}

// dbPlans 读启用且币种匹配的套餐（FORCE_MASTER：刚改立即生效）。
func (pc *PaymentUsecase) dbPlans(ctx context.Context, currency string) ([]data.VipPlan, error) {
	var rows []data.VipPlan
	err := pc.db.Clauses(gormdb.Write).WithContext(ctx).
		Where("status = 1 AND currency = ?", currency).Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

// plansFromRows 合并 DB 套餐行与内置默认：plan 缺行回退默认值；同 currency 金额冲突
// （同价两套餐）整体回退默认，保证 amount↔plan 一一对应（幂等复用键与 webhook 反查的前提）。
func plansFromRows(rows []data.VipPlan) (amounts map[string]int64, days map[string]int) {
	if len(rows) == 0 {
		return defaultPlans, planDays
	}
	amounts = map[string]int64{}
	days = map[string]int{}
	for _, p := range planOrder {
		amounts[p] = defaultPlans[p]
		days[p] = planDays[p]
	}
	seen := map[int64]string{}
	for _, r := range rows {
		if _, known := planDays[r.PlanCode]; !known {
			continue // 仅内置三档参与定价
		}
		amounts[r.PlanCode] = r.AmountCents
		days[r.PlanCode] = r.Days
	}
	for _, p := range planOrder {
		a := amounts[p]
		if prev, dup := seen[a]; dup {
			_ = prev
			return defaultPlans, planDays // 冲突属配置错误，回退默认
		}
		seen[a] = p
	}
	return amounts, days
}

// planForAmounts 金额反查套餐：按固定顺序遍历（与下单同一映射，必中且唯一）。
func planForAmounts(amounts map[string]int64, days map[string]int, cents int64) (string, int, bool) {
	for _, p := range planOrder {
		if amounts[p] == cents {
			return p, days[p], true
		}
	}
	return "", 0, false
}

func centsOf(amount float64) int64 { return int64(math.Round(amount * 100)) }

// amountOK 金额强校验：整数分比较，允许支付侧舍入容差 ±1 分
// （ponytail: USDT/币种折算会有分位舍入，真实汇率换算接入时改为按汇率换算后比较）。
func amountOK(orderCents, evCents int64) bool {
	diff := orderCents - evCents
	return diff >= -1 && diff <= 1
}

// maybeClose 未支付超时置为已关闭（3）；返回是否发生关闭。
func maybeClose(o *data.PaymentOrder, now time.Time) bool {
	if o.Status == 0 && now.Sub(o.CreatedAt) > pendingCloseAfter {
		o.Status = 3
		return true
	}
	return false
}

func newOrderNo() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d%s", time.Now().UnixMilli(), hex.EncodeToString(b[:]))
}

func orderInfo(o data.PaymentOrder) OrderInfo {
	return OrderInfo{OrderNo: o.OrderNo, Status: o.Status, AmountCents: centsOf(o.Amount),
		Currency: o.Currency, Provider: o.Provider, TxID: o.TxID, PaidAt: o.PaidAt}
}

// ---- 管理端（T-P-09~13；requireAdmin 在 service 层校验；写操作审计，密钥不落明文）----

// OrderQuery 流水筛选。
type OrderQuery struct {
	UserID   uint64
	Provider string
	Status   *int8
	StartAt  *time.Time
	EndAt    *time.Time
}

// OrderItem 流水条目（管理端视图）。
type OrderItem struct {
	OrderNo     string
	UserID      uint64
	AmountCents int64
	Currency    string
	Provider    string
	Status      int8
	TxID        string
	PaidAt      *time.Time
	CreatedAt   time.Time
}

// OrderStatsInfo 流水汇总。
type OrderStatsInfo struct {
	Total   int64
	Paid    int64
	Pending int64
	Failed  int64
	Closed  int64
	Amount  int64 // 已付总金额（分）
}

// ListOrders 流水分页：FORCE_MASTER（支付结果一致性优先），id 倒序。
func (pc *PaymentUsecase) ListOrders(ctx context.Context, q OrderQuery, page pkg.Page) ([]OrderItem, int64, error) {
	db := pc.applyOrderFilter(pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.PaymentOrder{}), q)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, pkg.ErrAdminDB
	}
	var rows []data.PaymentOrder
	if err := db.Order("id DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, pkg.ErrAdminDB
	}
	out := make([]OrderItem, 0, len(rows))
	for i := range rows {
		out = append(out, OrderItem{OrderNo: rows[i].OrderNo, UserID: rows[i].UserID,
			AmountCents: centsOf(rows[i].Amount), Currency: rows[i].Currency, Provider: rows[i].Provider,
			Status: rows[i].Status, TxID: rows[i].TxID, PaidAt: rows[i].PaidAt, CreatedAt: rows[i].CreatedAt})
	}
	return out, total, nil
}

// OrderStats 流水汇总：总单数 + 按状态计数 + 已付总金额。
func (pc *PaymentUsecase) OrderStats(ctx context.Context, q OrderQuery) (OrderStatsInfo, error) {
	var s OrderStatsInfo
	base := pc.applyOrderFilter(pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.PaymentOrder{}), q)
	if err := base.Session(&gorm.Session{}).Count(&s.Total).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	type grp struct {
		Status int8
		Cnt    int64
		Amt    float64
	}
	var rows []grp
	if err := base.Session(&gorm.Session{}).
		Select("status, COUNT(*) AS cnt, SUM(CASE WHEN status = 1 THEN amount ELSE 0 END) AS amt").
		Group("status").Scan(&rows).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	for _, r := range rows {
		switch r.Status {
		case 1:
			s.Paid, s.Amount = r.Cnt, centsOf(r.Amt)
		case 0:
			s.Pending = r.Cnt
		case 2:
			s.Failed = r.Cnt
		case 3:
			s.Closed = r.Cnt
		}
	}
	return s, nil
}

func (pc *PaymentUsecase) applyOrderFilter(db *gorm.DB, q OrderQuery) *gorm.DB {
	if q.UserID > 0 {
		db = db.Where("user_id = ?", q.UserID)
	}
	if q.Provider != "" {
		db = db.Where("provider = ?", q.Provider)
	}
	if q.Status != nil {
		db = db.Where("status = ?", *q.Status)
	}
	if q.StartAt != nil {
		db = db.Where("created_at >= ?", *q.StartAt)
	}
	if q.EndAt != nil {
		db = db.Where("created_at < ?", *q.EndAt)
	}
	return db
}

// ProviderInfo 管理端支付方式视图（不含 config 明文，仅标志是否已配置）。
type ProviderInfo struct {
	ID         uint64
	Code       string
	Lang       string
	Region     string
	Enabled    int8
	Sort       int
	Configured bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ProviderUpsert / ProviderUpdate 创建/更新入参（service 层从 proto 转换）。
type ProviderUpsert struct {
	Code   string
	Lang   string
	Region string
	Sort   int
	Config map[string]string
}

type ProviderUpdate struct {
	Lang    *string
	Region  *string
	Sort    *int
	Enabled *int8
	Config  map[string]string // 非空则合并重加密；空值键保留原值
}

// ListProviders 全部支付方式（含禁用），sort 升序。
func (pc *PaymentUsecase) ListProviders(ctx context.Context) ([]ProviderInfo, error) {
	var rows []data.PaymentProvider
	if err := pc.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	out := make([]ProviderInfo, 0, len(rows))
	for i := range rows {
		out = append(out, providerInfo(&rows[i]))
	}
	return out, nil
}

func providerInfo(r *data.PaymentProvider) ProviderInfo {
	return ProviderInfo{ID: r.ID, Code: r.Code, Lang: r.Lang, Region: r.Region,
		Enabled: r.Enabled, Sort: r.Sort, Configured: r.Config != "", CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// ProviderInfoFrom 供 service 层将写后读的行转管理端视图。
func ProviderInfoFrom(r *data.PaymentProvider) ProviderInfo { return providerInfo(r) }

// CreateProvider 新建支付方式；config 密钥字段加密后落库。
func (pc *PaymentUsecase) CreateProvider(ctx context.Context, adminID int64, req ProviderUpsert) (*data.PaymentProvider, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" || len(code) > 32 {
		return nil, pkg.ErrInvalidArgument
	}
	if req.Lang == "" {
		req.Lang = "*"
	}
	if req.Region == "" {
		req.Region = "*"
	}
	enc, err := pc.encryptConfig(req.Config)
	if err != nil {
		return nil, pkg.ErrAdminDB
	}
	r := data.PaymentProvider{Code: code, Lang: req.Lang, Region: req.Region,
		Enabled: 1, Sort: req.Sort, Config: enc}
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&r).Error; err != nil {
		if isDup(err) {
			return nil, pkg.ErrConflict
		}
		return nil, pkg.ErrAdminDB
	}
	data.WriteAudit(pc.db, ctx, adminID, "provider_create", "payment_provider", code,
		"lang="+r.Lang+" region="+r.Region+" sort="+strconv.Itoa(r.Sort))
	return &r, nil
}

// UpdateProvider 更新支付方式；config 传入字段合并原值后整体重加密，未传字段保留原值。
func (pc *PaymentUsecase) UpdateProvider(ctx context.Context, adminID int64, id uint64, req ProviderUpdate) (*data.PaymentProvider, error) {
	var r data.PaymentProvider
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	updates := map[string]any{}
	if req.Lang != nil {
		updates["lang"] = *req.Lang
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(req.Config) > 0 {
		cur, err := pc.decryptConfig(r.Config)
		if err != nil {
			return nil, pkg.ErrAdminDB
		}
		merged := map[string]string{}
		for k, v := range cur {
			if s, ok := v.(string); ok {
				merged[k] = s
			}
		}
		for k, v := range req.Config {
			if v != "" { // 空值 = 保留原值
				merged[k] = v
			}
		}
		enc, err := pc.encryptConfig(merged)
		if err != nil {
			return nil, pkg.ErrAdminDB
		}
		updates["config"] = enc
	}
	if len(updates) == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	res := pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.PaymentProvider{}).
		Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	// detail 只列字段名（含 config 标志），绝不含明文
	data.WriteAudit(pc.db, ctx, adminID, "provider_update", "payment_provider", r.Code,
		"fields="+providerAuditKeys(updates))
	// FORCE_MASTER: 写后读
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	return &r, nil
}

// ToggleProvider 启停：enabled 翻转。
func (pc *PaymentUsecase) ToggleProvider(ctx context.Context, adminID int64, id uint64) (*data.PaymentProvider, error) {
	var r data.PaymentProvider
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	next := int8(1)
	if r.Enabled == 1 {
		next = 0
	}
	res := pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.PaymentProvider{}).
		Where("id = ?", id).Update("enabled", next)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	data.WriteAudit(pc.db, ctx, adminID, "provider_toggle", "payment_provider", r.Code,
		"enabled="+strconv.Itoa(int(next)))
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	return &r, nil
}

// DeleteProvider 硬删除支付方式（订单表无外键，仅存渠道码字符串）。
func (pc *PaymentUsecase) DeleteProvider(ctx context.Context, adminID int64, id uint64) error {
	res := pc.db.Clauses(gormdb.Write).WithContext(ctx).Where("id = ?", id).Delete(&data.PaymentProvider{})
	if res.Error != nil {
		return pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	data.WriteAudit(pc.db, ctx, adminID, "provider_delete", "payment_provider", strconv.FormatUint(id, 10), "")
	return nil
}

// PlanUpsert / PlanUpdate 套餐创建/更新入参。
type PlanUpsert struct {
	PlanCode    string
	Days        int
	AmountCents int64
	Currency    string
	Label       string
	Sort        int
}

type PlanUpdate struct {
	Label    *string
	Days     *int
	Amount   *int64
	Currency *string
	Sort     *int
	Status   *int8
}

// ListPlans 全部套餐（含禁用）。
func (pc *PaymentUsecase) ListPlans(ctx context.Context) ([]data.VipPlan, error) {
	var rows []data.VipPlan
	if err := pc.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	return rows, nil
}

func (pc *PaymentUsecase) CreatePlan(ctx context.Context, adminID int64, req PlanUpsert) (*data.VipPlan, error) {
	code := strings.TrimSpace(req.PlanCode)
	if _, ok := planDays[code]; !ok {
		return nil, pkg.ErrInvalidArgument // 仅内置三档可配置
	}
	if req.Days <= 0 || req.AmountCents <= 0 {
		return nil, pkg.ErrInvalidArgument
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}
	p := data.VipPlan{PlanCode: code, Days: req.Days, AmountCents: req.AmountCents,
		Currency: currency, Label: strings.TrimSpace(req.Label), Sort: req.Sort, Status: 1}
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&p).Error; err != nil {
		if isDup(err) {
			return nil, pkg.ErrConflict
		}
		return nil, pkg.ErrAdminDB
	}
	data.WriteAudit(pc.db, ctx, adminID, "plan_create", "vip_plan", code,
		"days="+strconv.Itoa(p.Days)+" amount="+strconv.FormatInt(p.AmountCents, 10)+" currency="+p.Currency)
	return &p, nil
}

// UpdatePlan 更新套餐；字段可选，仅更新非空项。
func (pc *PaymentUsecase) UpdatePlan(ctx context.Context, adminID int64, id uint64, req PlanUpdate) (*data.VipPlan, error) {
	updates := map[string]any{}
	if req.Label != nil {
		updates["label"] = strings.TrimSpace(*req.Label)
	}
	if req.Days != nil {
		if *req.Days <= 0 {
			return nil, pkg.ErrInvalidArgument
		}
		updates["days"] = *req.Days
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, pkg.ErrInvalidArgument
		}
		updates["amount_cents"] = *req.Amount
	}
	if req.Currency != nil {
		c := strings.ToUpper(strings.TrimSpace(*req.Currency))
		if c == "" {
			return nil, pkg.ErrInvalidArgument
		}
		updates["currency"] = c
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			return nil, pkg.ErrBadState
		}
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	res := pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.VipPlan{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	data.WriteAudit(pc.db, ctx, adminID, "plan_update", "vip_plan", strconv.FormatUint(id, 10),
		"fields="+providerAuditKeys(updates))
	// FORCE_MASTER: 写后读
	var p data.VipPlan
	if err := pc.db.Clauses(gormdb.Write).WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	return &p, nil
}

// DeletePlan 软删（status=0）：历史订单按 plan_code 引用，硬删会悬空。
func (pc *PaymentUsecase) DeletePlan(ctx context.Context, adminID int64, id uint64) error {
	res := pc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.VipPlan{}).
		Where("id = ?", id).Update("status", 0)
	if res.Error != nil {
		return pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	data.WriteAudit(pc.db, ctx, adminID, "plan_delete", "vip_plan", strconv.FormatUint(id, 10), "soft")
	return nil
}

// encryptConfig 密钥配置 JSON 加密；空配置返回空串（未配置）。
func (pc *PaymentUsecase) encryptConfig(cfg map[string]string) (string, error) {
	if len(cfg) == 0 {
		return "", nil
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return pc.cr.Encrypt(string(b))
}

// providerAuditKeys 审计 detail 的字段名列表（仅键名，绝不含 config 值）。
func providerAuditKeys(updates map[string]any) string {
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// isDup MySQL 唯一键冲突。
func isDup(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}
