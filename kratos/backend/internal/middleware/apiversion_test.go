package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

func TestIsWebhookPath(t *testing.T) {
	cases := map[string]bool{
		"/api/payments/webhook/stripe":      true,
		"/api/payments/webhook/nowpayments": true,
		"/api/payments/webhook/{provider}":  true,
		"/api/payments/order":               false,
		"/api/users/login":                  false,
		"/api/payments/webhook":             false, // 无尾斜杠不匹配路由
		"":                                  false,
	}
	for path, want := range cases {
		if got := isWebhookPath(path); got != want {
			t.Errorf("isWebhookPath(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestApiVersionMiddleware: 真实路由模板下，webhook 免版本头；普通接口缺/错版本
// 返回 ErrApiVersion(140426)。防止仅路径判断正确、中间件接线却漏掉豁免。
func TestApiVersionMiddleware(t *testing.T) {
	srv := khttp.NewServer(
		khttp.Middleware(ApiVersion()),
		khttp.ErrorEncoder(func(w http.ResponseWriter, _ *http.Request, err error) {
			w.Header().Set("X-Err-Code", strconv.Itoa(int(kerrors.FromError(err).Code)))
			w.WriteHeader(http.StatusOK)
		}),
	)
	pass := func(ctx khttp.Context) error {
		h := ctx.Middleware(func(context.Context, any) (any, error) { return nil, nil })
		_, err := h(ctx, nil)
		return err
	}
	r := srv.Route("/")
	r.POST("/api/payments/webhook/{provider}", pass)
	r.POST("/api/payments/order", pass)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	do := func(path, version string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, ts.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if version != "" {
			req.Header.Set("X-Api-Version", version)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp
	}

	if resp := do("/api/payments/webhook/stripe", ""); resp.StatusCode != http.StatusOK || resp.Header.Get("X-Err-Code") != "" {
		t.Fatalf("webhook without version must bypass: status=%d err=%q",
			resp.StatusCode, resp.Header.Get("X-Err-Code"))
	}
	if resp := do("/api/payments/order", ""); resp.Header.Get("X-Err-Code") != "140426" {
		t.Fatalf("normal API without version must be 140426, got %q", resp.Header.Get("X-Err-Code"))
	}
	if resp := do("/api/payments/order", "v2"); resp.Header.Get("X-Err-Code") != "140426" {
		t.Fatalf("unsupported version must be 140426, got %q", resp.Header.Get("X-Err-Code"))
	}
	if resp := do("/api/payments/order", "v1"); resp.Header.Get("X-Err-Code") != "" {
		t.Fatalf("v1 must pass, got err=%q", resp.Header.Get("X-Err-Code"))
	}
}
