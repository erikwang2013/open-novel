package service

// 章节 content 接口 CDN 缓存头测试：免费章节 public s-maxage、VIP 章节 no-store、
// CDN_BASE_URL 为空时不加头（行为与现状一致）。

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	chapterv1 "open-novel/backend/api/chapter/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func newTestChapter(t *testing.T) (*ChapterService, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewChapterService(biz.NewChapterUsecase(d)), d
}

func TestContentCacheControlHeader(t *testing.T) {
	t.Setenv("CDN_BASE_URL", "https://cdn.example.com")
	s, d := newTestChapter(t)
	t.Cleanup(func() { d.Cache.Close() })

	ch := data.Chapter{BookID: 1, ChapterNo: 1, Title: "cdn 头测试", Status: 1}
	if err := d.DB.Create(&ch).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create(&data.ChapterContent{ChapterID: ch.ID, Lang: "zh-CN", Content: "正文"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.Where("chapter_id = ?", ch.ID).Delete(&data.ChapterContent{})
		d.DB.Delete(&data.Chapter{}, ch.ID)
	})

	srv := khttp.NewServer(khttp.Middleware(recovery.Recovery()))
	chapterv1.RegisterChapterHTTPServer(srv, s)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	get := func() string {
		resp, err := http.Get(ts.URL + "/api/chapters/" + strconv.FormatUint(ch.ID, 10) + "/content?lang=zh-CN")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		return resp.Header.Get("Cache-Control")
	}

	// 免费章节 + CDN 开启 → public, s-maxage=3600
	if got := get(); got != "public, s-maxage=3600" {
		t.Fatalf("free+cdn: want public, s-maxage=3600, got %q", got)
	}

	// VIP 章节 + CDN 开启 → no-store
	if err := d.DB.Model(&data.Chapter{}).Where("id = ?", ch.ID).Update("is_vip", 1).Error; err != nil {
		t.Fatal(err)
	}
	if got := get(); got != "no-store" {
		t.Fatalf("vip+cdn: want no-store, got %q", got)
	}

	// CDN 关闭（CDN_BASE_URL 空）→ 不加缓存头，行为与现状一致
	t.Setenv("CDN_BASE_URL", "")
	if got := get(); got != "" {
		t.Fatalf("cdn off: want no Cache-Control header, got %q", got)
	}
}
