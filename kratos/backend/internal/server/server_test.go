package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	"open-novel/backend/internal/pkg"
)

func TestCORSFilterAllowAll(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := corsFilter()(next)

	req := httptest.NewRequest(http.MethodOptions, "/api/users/login", nil)
	req.Header.Set("Origin", "https://web.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS: want 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Allow-Origin: want *, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "PATCH") {
		t.Fatalf("Allow-Methods missing PATCH: %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-Api-Version") {
		t.Fatalf("Allow-Headers missing X-Api-Version: %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("Allow-Credentials must be empty, got %q", got)
	}

	// 普通请求同样带 CORS 头并进入路由
	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("GET: want 200 + *, got %d/%q", rec.Code, rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSFilterSpecificOrigin(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://web.example, https://admin.example")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := corsFilter()(next)

	req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "https://web.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://web.example" {
		t.Fatalf("allowed origin: want echo, got %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary: want Origin, got %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("disallowed origin must not get Allow-Origin, got %q", got)
	}
}

func TestRespEncoderWebhook(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/payments/webhook/stripe", nil)
	rec := httptest.NewRecorder()
	if err := respEncoder(rec, req, nil); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != "success" {
		t.Fatalf("webhook: want 200 success, got %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("webhook content-type: %q", ct)
	}

	// 非 webhook 路径仍走默认 JSON 编码
	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec = httptest.NewRecorder()
	if err := respEncoder(rec, req, map[string]string{"ok": "1"}); err != nil {
		t.Fatal(err)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("default content-type: %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("default body: %q", rec.Body.String())
	}
}

func TestErrEncoderWebhook(t *testing.T) {
	// webhook 失败 → 500，渠道才会重试
	req := httptest.NewRequest(http.MethodPost, "/api/payments/webhook/stripe", nil)
	rec := httptest.NewRecorder()
	errEncoder(rec, req, pkg.ErrPayCreate)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("webhook error: want 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":190401`) {
		t.Fatalf("webhook error body: %q", rec.Body.String())
	}

	// 普通业务错误仍 200 + 业务码
	req = httptest.NewRequest(http.MethodPost, "/api/payments/order", nil)
	rec = httptest.NewRecorder()
	errEncoder(rec, req, pkg.ErrPayCreate)
	if rec.Code != http.StatusOK {
		t.Fatalf("business error: want 200, got %d", rec.Code)
	}

	// 标准 HTTP 错误码原样
	req = httptest.NewRequest(http.MethodGet, "/api/books", nil)
	rec = httptest.NewRecorder()
	errEncoder(rec, req, kerrors.New(http.StatusNotFound, "NOT_FOUND", "not found"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("standard error: want 404, got %d", rec.Code)
	}
}
