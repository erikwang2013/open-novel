package service

// 支付服务适配层：proto 消息 ↔ biz 用例。
// CreateOrder/GetOrder 需登录；ListMethods/Webhook 公开（验签在 biz）。

import (
	"context"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport"

	paymentv1 "open-novel/backend/api/payment/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/middleware"
	"open-novel/backend/internal/pkg"
)

type PaymentService struct {
	uc *biz.PaymentUsecase
	paymentv1.UnimplementedPaymentServer
}

func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService { return &PaymentService{uc: uc} }

func (s *PaymentService) CreateOrder(ctx context.Context, req *paymentv1.CreateOrderReq) (*paymentv1.CreateOrderReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	o, err := s.uc.CreateOrder(ctx, uint64(c.UID), req.Plan, pickLang(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	return &paymentv1.CreateOrderReply{OrderNo: o.OrderNo, Amount: o.AmountCents,
		Currency: o.Currency, CheckoutUrl: o.CheckoutURL, Provider: o.Provider}, nil
}

func (s *PaymentService) GetOrder(ctx context.Context, req *paymentv1.GetOrderReq) (*paymentv1.GetOrderReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	o, err := s.uc.GetOrder(ctx, uint64(c.UID), req.OrderNo)
	if err != nil {
		return nil, err
	}
	r := &paymentv1.GetOrderReply{OrderNo: o.OrderNo, Status: int32(o.Status),
		Amount: o.AmountCents, Currency: o.Currency, Provider: o.Provider, TxId: o.TxID}
	if o.PaidAt != nil {
		r.PaidAt = o.PaidAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return r, nil
}

// ListPublicPlans 公开套餐列表（无鉴权）：仅 status=1，sort 升序。
func (s *PaymentService) ListPublicPlans(ctx context.Context, req *paymentv1.ListPublicPlansReq) (*paymentv1.ListPublicPlansReply, error) {
	items, err := s.uc.ListPublicPlans(ctx)
	if err != nil {
		return nil, err
	}
	r := &paymentv1.ListPublicPlansReply{}
	for _, it := range items {
		r.List = append(r.List, &paymentv1.PublicPlanItem{PlanCode: it.PlanCode, Days: int64(it.Days),
			Amount: it.AmountCents, Currency: it.Currency, Label: it.Label})
	}
	return r, nil
}

// VipStatus 登录态：active（vip_expires_at > now）+ 到期时间（RFC3339，非会员为空）。
func (s *PaymentService) VipStatus(ctx context.Context, req *paymentv1.VipStatusReq) (*paymentv1.VipStatusReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	active, at := s.uc.IsVipActive(ctx, uint64(c.UID))
	r := &paymentv1.VipStatusReply{Active: active}
	if !at.IsZero() {
		r.VipExpiresAt = at.Format(time.RFC3339)
	}
	return r, nil
}

func (s *PaymentService) ListMethods(ctx context.Context, req *paymentv1.ListMethodsReq) (*paymentv1.ListMethodsReply, error) {
	items, err := s.uc.ListMethods(ctx, pickLang(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	r := &paymentv1.ListMethodsReply{}
	for _, it := range items {
		r.List = append(r.List, &paymentv1.MethodItem{Code: it.Code, Lang: it.Lang,
			Region: it.Region, Sort: it.Sort})
	}
	return r, nil
}

func (s *PaymentService) Webhook(ctx context.Context, req *paymentv1.WebhookReq) (*paymentv1.EmptyReply, error) {
	// 无鉴权端点：验签在 biz；原始 body 由 middleware.RawBody 预读（消息绑定会丢原始字节）
	raw := middleware.RawBodyFrom(ctx)
	headers := map[string]string{}
	if tr, ok := transport.FromServerContext(ctx); ok {
		h := tr.RequestHeader()
		for _, name := range []string{
			"Stripe-Signature", "X-Nowpayments-Sig",
			"X-Razorpay-Signature", "X-KOMOJU-SIGNATURE", "X-IAMPORT-TOKEN",
			"X-CALLBACK-TOKEN",
			"PayPal-Transmission-Id", "PayPal-Transmission-Time",
			"PayPal-Transmission-Sig", "PayPal-Webhook-Id",
		} {
			headers[name] = h.Get(name)
		}
	}
	if err := s.uc.Webhook(ctx, req.Provider, raw, headers); err != nil {
		return nil, err
	}
	return &paymentv1.EmptyReply{}, nil
}

// ---- 管理端（T-P-09~13，全部 requireAdmin → 180401）----

func (s *PaymentService) ListOrders(ctx context.Context, req *paymentv1.ListOrdersReq) (*paymentv1.ListOrdersReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	page := pkg.ParsePage(req.Page, req.PageSize)
	var status *int8
	if req.Status != -1 { // -1 = 全部
		if req.Status < 0 || req.Status > 3 {
			return nil, pkg.ErrInvalidArgument
		}
		v := int8(req.Status)
		status = &v
	}
	if req.UserId < 0 {
		return nil, pkg.ErrInvalidArgument
	}
	start, err := parseRangeTime(req.StartAt)
	if err != nil {
		return nil, err
	}
	end, err := parseRangeEnd(req.EndAt)
	if err != nil {
		return nil, err
	}
	q := biz.OrderQuery{UserID: u64(req.UserId), Provider: req.Provider,
		Status: status, StartAt: start, EndAt: end}
	items, total, err := s.uc.ListOrders(ctx, q, page)
	if err != nil {
		return nil, err
	}
	r := &paymentv1.ListOrdersReply{Total: total}
	for _, it := range items {
		item := &paymentv1.OrderItem{OrderNo: it.OrderNo, UserId: i64(it.UserID),
			Amount: it.AmountCents, Currency: it.Currency, Provider: it.Provider,
			Status: int32(it.Status), TxId: it.TxID,
			CreatedAt: it.CreatedAt.Format(time.RFC3339)}
		if it.PaidAt != nil {
			item.PaidAt = it.PaidAt.Format(time.RFC3339)
		}
		r.List = append(r.List, item)
	}
	return r, nil
}

func (s *PaymentService) OrderStats(ctx context.Context, req *paymentv1.OrderStatsReq) (*paymentv1.OrderStatsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	start, err := parseRangeTime(req.StartAt)
	if err != nil {
		return nil, err
	}
	end, err := parseRangeEnd(req.EndAt)
	if err != nil {
		return nil, err
	}
	st, err := s.uc.OrderStats(ctx, biz.OrderQuery{StartAt: start, EndAt: end})
	if err != nil {
		return nil, err
	}
	return &paymentv1.OrderStatsReply{Total: st.Total, Paid: st.Paid, Pending: st.Pending,
		Failed: st.Failed, Closed: st.Closed, Amount: st.Amount}, nil
}

func (s *PaymentService) ListProviders(ctx context.Context, req *paymentv1.ListProvidersReq) (*paymentv1.ListProvidersReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	r := &paymentv1.ListProvidersReply{Total: int64(len(items))}
	for _, it := range items {
		r.List = append(r.List, toProviderReply(it))
	}
	return r, nil
}

func (s *PaymentService) CreateProvider(ctx context.Context, req *paymentv1.CreateProviderReq) (*paymentv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.CreateProvider(ctx, c.UID, biz.ProviderUpsert{
		Code: req.Code, Lang: req.Lang, Region: req.Region,
		Sort: int(req.Sort), Config: req.Config,
	})
	if err != nil {
		return nil, err
	}
	return toProviderReply(biz.ProviderInfoFrom(p)), nil
}

func (s *PaymentService) UpdateProvider(ctx context.Context, req *paymentv1.UpdateProviderReq) (*paymentv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var lang, region *string
	if req.Lang != nil {
		lang = req.Lang
	}
	if req.Region != nil {
		region = req.Region
	}
	var sort *int
	if req.Sort != nil {
		v := int(*req.Sort)
		sort = &v
	}
	var enabled *int8
	if req.Enabled != nil {
		v := int8(*req.Enabled)
		enabled = &v
	}
	p, err := s.uc.UpdateProvider(ctx, c.UID, u64(req.Id), biz.ProviderUpdate{
		Lang: lang, Region: region, Sort: sort, Enabled: enabled, Config: req.Config,
	})
	if err != nil {
		return nil, err
	}
	return toProviderReply(biz.ProviderInfoFrom(p)), nil
}

func (s *PaymentService) DeleteProvider(ctx context.Context, req *paymentv1.DeleteProviderReq) (*paymentv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteProvider(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &paymentv1.EmptyReply{}, nil
}

func (s *PaymentService) ToggleProvider(ctx context.Context, req *paymentv1.ToggleProviderReq) (*paymentv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.ToggleProvider(ctx, c.UID, u64(req.Id))
	if err != nil {
		return nil, err
	}
	return toProviderReply(biz.ProviderInfoFrom(p)), nil
}

func (s *PaymentService) ListPlans(ctx context.Context, req *paymentv1.ListPlansReq) (*paymentv1.ListPlansReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	r := &paymentv1.ListPlansReply{Total: int64(len(items))}
	for i := range items {
		r.List = append(r.List, toPlanReply(&items[i]))
	}
	return r, nil
}

func (s *PaymentService) CreatePlan(ctx context.Context, req *paymentv1.CreatePlanReq) (*paymentv1.PlanReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.uc.CreatePlan(ctx, c.UID, biz.PlanUpsert{
		PlanCode: req.PlanCode, Days: int(req.Days), AmountCents: req.Amount,
		Currency: req.Currency, Label: req.Label, Sort: int(req.Sort),
	})
	if err != nil {
		return nil, err
	}
	return toPlanReply(p), nil
}

func (s *PaymentService) UpdatePlan(ctx context.Context, req *paymentv1.UpdatePlanReq) (*paymentv1.PlanReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var label *string
	if req.Label != nil {
		label = req.Label
	}
	var days *int
	if req.Days != nil {
		v := int(*req.Days)
		days = &v
	}
	var amount *int64
	if req.Amount != nil {
		amount = req.Amount
	}
	var currency *string
	if req.Currency != nil {
		currency = req.Currency
	}
	var sort *int
	if req.Sort != nil {
		v := int(*req.Sort)
		sort = &v
	}
	var status *int8
	if req.Status != nil {
		v := int8(*req.Status)
		status = &v
	}
	p, err := s.uc.UpdatePlan(ctx, c.UID, u64(req.Id), biz.PlanUpdate{
		Label: label, Days: days, Amount: amount, Currency: currency, Sort: sort, Status: status,
	})
	if err != nil {
		return nil, err
	}
	return toPlanReply(p), nil
}

func (s *PaymentService) DeletePlan(ctx context.Context, req *paymentv1.DeletePlanReq) (*paymentv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeletePlan(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &paymentv1.EmptyReply{}, nil
}

// toProviderReply / toPlanReply：管理端视图，config 只出是否已配置标志。
func toProviderReply(p biz.ProviderInfo) *paymentv1.ProviderReply {
	return &paymentv1.ProviderReply{Id: i64(p.ID), Code: p.Code, Lang: p.Lang,
		Region: p.Region, Enabled: int32(p.Enabled), Sort: int32(p.Sort),
		ConfigConfigured: p.Configured,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339)}
}

func toPlanReply(p *data.VipPlan) *paymentv1.PlanReply {
	return &paymentv1.PlanReply{Id: i64(p.ID), PlanCode: p.PlanCode, Days: int64(p.Days),
		Amount: p.AmountCents, Currency: p.Currency, Label: p.Label,
		Sort: int32(p.Sort), Status: int32(p.Status)}
}

// parseRangeTime 解析时间筛选：RFC3339 或 2006-01-02。
func parseRangeTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {
		return &t, nil
	}
	return nil, pkg.ErrInvalidArgument
}

// parseRangeEnd 与 parseRangeTime 相同；日期型（无 T）表示含当天，外扩 24h。
func parseRangeEnd(s string) (*time.Time, error) {
	t, err := parseRangeTime(s)
	if err != nil || t == nil {
		return t, err
	}
	if !strings.Contains(s, "T") {
		v := t.Add(24 * time.Hour)
		t = &v
	}
	return t, nil
}
