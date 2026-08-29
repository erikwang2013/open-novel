package biz

// DeepL 翻译用例测试：httptest mock DeepL + 真实 DB（fixtures 带 marker，测试后清理）。

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

const translateMarker = "coder72b_tr"

// mockDeepL 返回 DeepL 服务 mock；failTitle 标题含该串的请求返回 500（模拟单章失败）。
func mockDeepL(t *testing.T, failTitle string) (*httptest.Server, *[]deeplReq) {
	t.Helper()
	var got []deeplReq
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "DeepL-Auth-Key test-key" {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		var req deeplReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if failTitle != "" && len(req.Text) > 0 && strings.Contains(req.Text[0], failTitle) {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		got = append(got, req)
		out := make([]map[string]string, len(req.Text))
		for i, txt := range req.Text {
			out[i] = map[string]string{"detected_source_language": "ZH", "text": fmt.Sprintf("[%s] %s", req.TargetLang, txt)}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"translations": out})
	}))
	t.Cleanup(s.Close)
	return s, &got
}

func newTestAdminUc(t *testing.T) (*AdminUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewAdminUsecase(d, nil), d
}

// TestTranslateAuthAndMapping 校验 Authorization header、target_lang 映射、请求体 text 数组。
func TestTranslateAuthAndMapping(t *testing.T) {
	s, got := mockDeepL(t, "")
	tr := NewTranslator(s.URL, "test-key")
	ctx := context.Background()

	for _, c := range []struct{ lang, target string }{
		{"zh-CN", "ZH-HANS"}, {"en", "EN"}, {"ja", "JA"}, {"ko", "KO"},
		{"fr", "FR"}, {"de", "DE"}, {"es", "ES"}, {"ru", "RU"},
		{"pt", "PT"}, {"hi", "HI"}, {"ar", "AR"}, {"id", "ID"},
	} {
		out, err := tr.translate(ctx, c.lang, []string{"你好", "世界"})
		if err != nil {
			t.Fatalf("%s: %v", c.lang, err)
		}
		if len(out) != 2 || !strings.HasPrefix(out[0], "["+c.target+"]") || !strings.HasPrefix(out[1], "["+c.target+"]") {
			t.Fatalf("%s: order mismatch: %v", c.lang, out)
		}
	}
	if len(*got) != 12 {
		t.Fatalf("requests: want 12, got %d", len(*got))
	}
	for _, r := range *got {
		if len(r.Text) != 2 || r.Text[0] != "你好" || r.Text[1] != "世界" {
			t.Fatalf("body text mismatch: %+v", r)
		}
	}
}

func TestTranslateNotConfigured(t *testing.T) {
	tr := NewTranslator("", "")
	if _, err := tr.translate(context.Background(), "en", []string{"hi"}); err != pkg.ErrTranslateNotCfg {
		t.Fatalf("want ErrTranslateNotCfg, got %v", err)
	}
}

func TestTranslateUnsupportedLang(t *testing.T) {
	s, _ := mockDeepL(t, "")
	tr := NewTranslator(s.URL, "test-key")
	if _, err := tr.translate(context.Background(), "bn", []string{"hi"}); err != pkg.ErrTranslateFailed {
		t.Fatalf("want ErrTranslateFailed for bn, got %v", err)
	}
}

func TestTranslateServerError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid key", http.StatusForbidden)
	}))
	t.Cleanup(s.Close)
	tr := NewTranslator(s.URL, "test-key")
	if _, err := tr.translate(context.Background(), "en", []string{"hi"}); err != pkg.ErrTranslateFailed {
		t.Fatalf("want ErrTranslateFailed on 403, got %v", err)
	}
}

func TestTranslateEmptyTranslations(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"translations":[]}`))
	}))
	t.Cleanup(s.Close)
	tr := NewTranslator(s.URL, "test-key")
	if _, err := tr.translate(context.Background(), "en", []string{"hi"}); err != pkg.ErrTranslateFailed {
		t.Fatalf("want ErrTranslateFailed on empty translations, got %v", err)
	}
}

// TestTranslateBookDB 书籍标题/简介翻译 upsert（含重复调用覆盖）。
func TestTranslateBookDB(t *testing.T) {
	s, _ := mockDeepL(t, "")
	uc, d := newTestAdminUc(t)
	uc.tr = NewTranslator(s.URL, "test-key")
	ctx := context.Background()

	b := data.Book{Title: translateMarker + "书", Summary: translateMarker + "简介", Lang: "zh-CN", Status: 1}
	if err := d.DB.WithContext(ctx).Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.WithContext(ctx).Where("book_id = ?", b.ID).Delete(&data.BookTranslation{})
		d.DB.WithContext(ctx).Delete(&data.Book{}, b.ID)
	})

	got, err := uc.TranslateBook(ctx, b.ID, "en")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got.Title, "[EN]") || !strings.HasPrefix(got.Summary, "[EN]") {
		t.Fatalf("translated fields: %+v", got)
	}
	// 再次翻译同语言覆盖更新，不新增行
	if _, err := uc.TranslateBook(ctx, b.ID, "en"); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := d.DB.WithContext(ctx).Model(&data.BookTranslation{}).Where("book_id = ? AND lang = ?", b.ID, "en").Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("upsert: want 1 row, got %d", n)
	}
}

// TestTranslateBookChaptersPartial 批量章节部分失败不中断。
func TestTranslateBookChaptersPartial(t *testing.T) {
	s, _ := mockDeepL(t, "坏章") // 标题含「坏章」的章节翻译失败
	uc, d := newTestAdminUc(t)
	uc.tr = NewTranslator(s.URL, "test-key")
	ctx := context.Background()

	b := data.Book{Title: translateMarker + "书", Lang: "zh-CN", Status: 1}
	if err := d.DB.WithContext(ctx).Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	var chs []data.Chapter
	for _, no := range []uint32{1, 2, 3} {
		title := fmt.Sprintf("%s第%d章", translateMarker, no)
		if no == 2 {
			title = "坏章二号"
		}
		ch := data.Chapter{BookID: b.ID, ChapterNo: no, Title: title, Status: 1}
		if err := d.DB.WithContext(ctx).Create(&ch).Error; err != nil {
			t.Fatal(err)
		}
		chs = append(chs, ch)
		if err := d.DB.WithContext(ctx).Create(&data.ChapterContent{
			ChapterID: ch.ID, Lang: "zh-CN", Content: fmt.Sprintf("%s正文%d", translateMarker, no),
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		ids := []uint64{}
		for _, c := range chs {
			ids = append(ids, c.ID)
		}
		d.DB.WithContext(ctx).Where("chapter_id IN ?", ids).Delete(&data.ChapterContent{})
		d.DB.WithContext(ctx).Delete(&data.Chapter{}, ids)
		d.DB.WithContext(ctx).Where("book_id = ?", b.ID).Delete(&data.BookTranslation{})
		d.DB.WithContext(ctx).Delete(&data.Book{}, b.ID)
	})

	res, err := uc.TranslateBookChapters(ctx, b.ID, "en")
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 || res.Succeeded != 2 || res.Failed != 1 {
		t.Fatalf("want total=3 succeeded=2 failed=1, got %+v", res)
	}
	if len(res.FailedChapters) != 1 || res.FailedChapters[0] != 2 {
		t.Fatalf("failed chapter numbers: %v", res.FailedChapters)
	}
	// 成功章节正文已写入目标语言
	for _, ch := range []data.Chapter{chs[0], chs[2]} {
		var ct data.ChapterContent
		if err := d.DB.WithContext(ctx).Where("chapter_id = ? AND lang = ?", ch.ID, "en").First(&ct).Error; err != nil {
			t.Fatalf("chapter %d missing en content: %v", ch.ChapterNo, err)
		}
		if !strings.HasPrefix(ct.Content, "[EN]") {
			t.Fatalf("chapter %d content not translated: %q", ch.ChapterNo, ct.Content)
		}
	}
	// 失败章节无 en 内容
	var n int64
	if err := d.DB.WithContext(ctx).Model(&data.ChapterContent{}).
		Where("chapter_id = ? AND lang = ?", chs[1].ID, "en").Count(&n).Error; err != nil || n != 0 {
		t.Fatalf("failed chapter should have no en content, n=%d err=%v", n, err)
	}
}
