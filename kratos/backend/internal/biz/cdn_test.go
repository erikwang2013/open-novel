package biz

// CDN 章节静态化机制测试：缓存头判定、env 门控、purge URL 模板、purge 失败不报错、
// 状态变更触发 purge。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func TestChapterCacheControl(t *testing.T) {
	if got := ChapterCacheControl(true); got != "no-store" {
		t.Fatalf("vip: want no-store, got %q", got)
	}
	if got := ChapterCacheControl(false); got != "public, s-maxage=3600" {
		t.Fatalf("free: want public, s-maxage=3600, got %q", got)
	}
}

func TestCdnEnabledEnvGate(t *testing.T) {
	t.Setenv("CDN_BASE_URL", "")
	if CdnEnabled() {
		t.Fatal("CDN_BASE_URL empty must disable CDN")
	}
	t.Setenv("CDN_BASE_URL", "https://cdn.example.com")
	if !CdnEnabled() {
		t.Fatal("CDN_BASE_URL set must enable CDN")
	}
}

// TestPurgeChapterAsyncURLTemplate: purge 端点为 mock，验证 {key} 模板替换为
// chapter/{id}?lang={lang}。
func TestPurgeChapterAsyncURLTemplate(t *testing.T) {
	got := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.URL.RequestURI()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	t.Setenv("CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("CDN_PURGE_URL", ts.URL+"/purge/{key}")

	PurgeChapterAsync(123, "zh-CN")
	select {
	case u := <-got:
		if u != "/purge/chapter/123?lang=zh-CN" {
			t.Fatalf("want /purge/chapter/123?lang=zh-CN, got %q", u)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("purge request not received")
	}
}

// TestPurgeFailureIgnored: purge 端点返回 5xx，不 panic、不阻塞（失败仅记日志）。
func TestPurgeFailureIgnored(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	t.Setenv("CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("CDN_PURGE_URL", ts.URL+"/purge/{key}")

	PurgeChapterAsync(7, "en") // 无返回值、无 error 路径
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
}

// TestSetChapterStatusPurgesChapter: 状态变更对该章节各 lang 逐个 purge。
func TestSetChapterStatusPurgesChapter(t *testing.T) {
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })

	ch := data.Chapter{BookID: 1, ChapterNo: 1, Title: "cdn 状态变更测试", Status: 1}
	if err := d.DB.Create(&ch).Error; err != nil {
		t.Fatal(err)
	}
	for _, l := range []string{"zh-CN", "en"} {
		if err := d.DB.Create(&data.ChapterContent{ChapterID: ch.ID, Lang: l, Content: "x"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		d.DB.Where("chapter_id = ?", ch.ID).Delete(&data.ChapterContent{})
		d.DB.Delete(&data.Chapter{}, ch.ID)
	})

	var mu sync.Mutex
	var hits []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits = append(hits, r.URL.RequestURI())
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	t.Setenv("CDN_BASE_URL", "https://cdn.example.com")
	t.Setenv("CDN_PURGE_URL", ts.URL+"/purge/{key}")

	uc := NewChapterUsecase(d)
	if err := uc.SetChapterStatus(context.Background(), 100, ch.ID, 0); err != nil {
		t.Fatal(err)
	}
	// 异步 purge，轮询等待两个 goroutine 完成
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(hits)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(hits) != 2 {
		t.Fatalf("want 2 purge hits (zh-CN + en), got %v", hits)
	}
}
