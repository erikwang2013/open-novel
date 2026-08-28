package service

// 支付服务适配层：proto 消息 ↔ biz 用例。
// CreateOrder/GetOrder 需登录；ListMethods/Webhook 公开（验签在 biz）。

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"

	paymentv1 "open-novel/backend/api/payment/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/middleware"
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
		headers["Stripe-Signature"] = h.Get("Stripe-Signature")
		headers["X-Nowpayments-Sig"] = h.Get("X-Nowpayments-Sig")
	}
	if err := s.uc.Webhook(ctx, req.Provider, raw, headers); err != nil {
		return nil, err
	}
	return &paymentv1.EmptyReply{}, nil
}
