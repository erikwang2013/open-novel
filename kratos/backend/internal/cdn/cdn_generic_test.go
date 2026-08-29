package cdn

// Generic adapter（§五 5.5）：{key} 模板逐 key 单请求（1 key/请求，§4.3 兼容断言）；
// 仅 DB 无启用厂商且存在 CDN_PURGE_URL 时激活（biz 兜底构造）。

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNewGenericValidate(t *testing.T) {
	if _, err := NewGeneric(map[string]any{"url_template": "http://x/purge"}); err == nil {
		t.Fatal("template without {key} must error")
	}
}

// TestGenericPurgeTemplate 每 key 一个请求，URL 模板 {key} 替换。
func TestGenericPurgeTemplate(t *testing.T) {
	var mu sync.Mutex
	var urls []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		urls = append(urls, r.URL.RequestURI())
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	p, err := NewGeneric(map[string]any{"url_template": ts.URL + "/purge/{key}"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "generic" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/1?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 || urls[0] != "/purge/chapter/1?lang=zh-CN" || urls[1] != "/purge/chapter/1?lang=en" {
		t.Fatalf("urls: %v", urls)
	}
}

// TestGenericNon2xx 500 → 可重试错误（manager 重试一次后记日志）。
func TestGenericNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	p, _ := NewGeneric(map[string]any{"url_template": ts.URL + "/purge/{key}"})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("5xx must be retriable, got %v", err)
	}
}
