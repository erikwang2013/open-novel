package middleware

// rawbody 中间件单测：经真实 khttp server（RPC 路由才会走 middleware 链）验证
// webhook 路径预读原始 body、非 webhook 路径不干预。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	paymentv1 "open-novel/backend/api/payment/v1"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type rawBodyEcho struct {
	paymentv1.UnimplementedPaymentServer
	captured []byte
}

func (s *rawBodyEcho) Webhook(ctx context.Context, req *paymentv1.WebhookReq) (*paymentv1.EmptyReply, error) {
	s.captured = RawBodyFrom(ctx)
	return &paymentv1.EmptyReply{}, nil
}

func (s *rawBodyEcho) ListPublicPlans(context.Context, *paymentv1.ListPublicPlansReq) (*paymentv1.ListPublicPlansReply, error) {
	return &paymentv1.ListPublicPlansReply{}, nil
}

func (s *rawBodyEcho) VipStatus(context.Context, *paymentv1.VipStatusReq) (*paymentv1.VipStatusReply, error) {
	return &paymentv1.VipStatusReply{}, nil
}

func TestRawBodyMiddleware(t *testing.T) {
	body := `{"payment_id":1,"payment_status":"finished"}`
	srv := khttp.NewServer(khttp.Middleware(RawBody()))
	svc := &rawBodyEcho{}
	paymentv1.RegisterPaymentHTTPServer(srv, svc)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/payments/webhook/stripe", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("webhook status %d", resp.StatusCode)
	}
	if string(svc.captured) != body {
		t.Fatalf("raw body mismatch: %q want %q", svc.captured, body)
	}
}

func TestRawBodySkipsOtherPaths(t *testing.T) {
	srv := khttp.NewServer(khttp.Middleware(RawBody()))
	svc := &rawBodyEcho{}
	paymentv1.RegisterPaymentHTTPServer(srv, svc)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/payments/order", "application/json", strings.NewReader(`{"plan":"monthly"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if svc.captured != nil {
		t.Fatalf("raw body must not be captured for non-webhook path: %q", svc.captured)
	}
}
