package cdn

// Cloudflare 适配器（§五 5.1）：Bearer + JSON files ≤30/批；200 且 success=true 才算成功；
// 429/5xx 可重试；缺必填键构造报错。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewCloudflareValidate(t *testing.T) {
	if _, err := NewCloudflare(map[string]any{}); err == nil {
		t.Fatal("empty cfg must error")
	}
	if _, err := NewCloudflare(map[string]any{"zone_id": "z"}); err == nil {
		t.Fatal("missing api_token must error")
	}
}

func TestCloudflarePurge(t *testing.T) {
	var got struct {
		Method string
		Auth   string
		Path   string
		Files  []string
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Auth = r.Header.Get("Authorization")
		got.Path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got.Files)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer ts.Close()

	p, err := NewCloudflare(map[string]any{"zone_id": "zone1", "api_token": "tok", "base_url": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "cloudflare" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost || !strings.HasSuffix(got.Path, "/zones/zone1/purge_cache") {
		t.Fatalf("request mismatch: %+v", got)
	}
	if got.Auth != "Bearer tok" {
		t.Fatalf("auth: %s", got.Auth)
	}
	if len(got.Files) != 2 || got.Files[0] != "chapter/1?lang=zh-CN" {
		t.Fatalf("files: %v", got.Files)
	}
}

// TestCloudflareBatch 超过 30 个 key 切多批（mock 计数）。
func TestCloudflareBatch(t *testing.T) {
	var batches int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches++
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer ts.Close()
	p, err := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL, "batch_size": float64(10)})
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 25)
	for i := range keys {
		keys[i] = "k" + strings.Repeat("x", i)
	}
	if err := p.Purge(t.Context(), keys); err != nil {
		t.Fatal(err)
	}
	if batches != 3 {
		t.Fatalf("want 3 batches, got %d", batches)
	}
}

func TestCloudflareNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	p, _ := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL})
	if err := p.Purge(t.Context(), []string{"k"}); err == nil {
		t.Fatal("401 must error")
	}
}

// TestCloudflareRetriable 429 → 错误标记可重试（manager 会重试一次）。
func TestCloudflareRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()
	p, _ := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}
