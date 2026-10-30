package biz

import (
	"context"
	"testing"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func TestSetLang(t *testing.T) {
	// 与 syncBookIndex 调用顺序一致：基础语言填满四语，翻译逐语覆盖，未知语言忽略。
	var d data.SearchDoc
	setLangBase(&d, "Français", "FR")
	if d.TitleZh != "Français" || d.TitleEn != "Français" || d.TitleJa != "Français" || d.TitleKo != "Français" {
		t.Fatalf("base should fill all languages: %+v", d)
	}
	setLangTranslation(&d, "zh-CN", "中文标题", "中文简介")
	setLangTranslation(&d, "en", "English Title", "English Summary")
	setLangTranslation(&d, "ja", "日本語タイトル", "日本語概要")
	setLangTranslation(&d, "ko", "한국어 제목", "한국어 요약")
	setLangTranslation(&d, "de", "Deutsch", "DE") // 索引不支持 de，忽略
	if d.TitleZh != "中文标题" || d.SummaryZh != "中文简介" {
		t.Fatalf("zh translation should override base, got zh=%q/%q", d.TitleZh, d.SummaryZh)
	}
	if d.TitleEn != "English Title" || d.SummaryEn != "English Summary" {
		t.Fatalf("en wrong: %q/%q", d.TitleEn, d.SummaryEn)
	}
	if d.TitleJa != "日本語タイトル" || d.SummaryJa != "日本語概要" {
		t.Fatalf("ja wrong: %q/%q", d.TitleJa, d.SummaryJa)
	}
	if d.TitleKo != "한국어 제목" || d.SummaryKo != "한국어 요약" {
		t.Fatalf("ko wrong: %q/%q", d.TitleKo, d.SummaryKo)
	}
}

// TestSyncBookIndex 集成：基础语言填满四语、受支持翻译覆盖对应语种、未知语言
// 不覆盖 title_zh、hot=有效评论数（status=1）。需本地 MySQL(3307)/OpenSearch(9200)。
func TestSyncBookIndex(t *testing.T) {
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })
	ctx := context.Background()

	b := data.Book{Title: "基础标题", Summary: "基础简介", Author: "索引作者", Lang: "zh-CN", Status: 1}
	if err := d.DB.WithContext(ctx).Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	trs := []data.BookTranslation{
		{BookID: b.ID, Lang: "zh-CN", Title: "中文标题", Summary: "中文简介"},
		{BookID: b.ID, Lang: "de", Title: "Deutsch Titel", Summary: "DE"}, // 不支持，忽略
		{BookID: b.ID, Lang: "ja", Title: "日本語タイトル", Summary: "日本語概要"},
	}
	if err := d.DB.WithContext(ctx).Create(&trs).Error; err != nil {
		t.Fatal(err)
	}
	comments := []data.Comment{
		{BookID: b.ID, UserID: 1, Content: "c1", Status: 1},
		{BookID: b.ID, UserID: 2, Content: "c2", Status: 1},
		{BookID: b.ID, UserID: 3, Content: "c3", Status: 2}, // 待审不计数
	}
	if err := d.DB.WithContext(ctx).Create(&comments).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = d.ES.Delete(ctx, b.ID)
		d.DB.WithContext(ctx).Where("book_id = ?", b.ID).Delete(&data.Comment{})
		d.DB.WithContext(ctx).Where("book_id = ?", b.ID).Delete(&data.BookTranslation{})
		d.DB.WithContext(ctx).Delete(&data.Book{}, b.ID)
	})

	if err := syncBookIndex(ctx, d.DB, NewSearchUsecase(d), b.ID); err != nil {
		t.Fatal(err)
	}
	docs, _, err := d.ES.Search(ctx, "中文标题", "zh-CN", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	var got *data.SearchDoc
	for i := range docs {
		if docs[i].BookID == b.ID {
			got = &docs[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("synced book %d not found in index", b.ID)
	}
	if got.TitleZh != "中文标题" {
		t.Fatalf("zh translation must win over base and ignored de: %q", got.TitleZh)
	}
	if got.TitleEn != "基础标题" || got.TitleKo != "基础标题" {
		t.Fatalf("untranslated langs should fall back to base: en=%q ko=%q", got.TitleEn, got.TitleKo)
	}
	if got.TitleJa != "日本語タイトル" {
		t.Fatalf("ja translation wrong: %q", got.TitleJa)
	}
	if got.Hot != 2 || got.Status != 1 {
		t.Fatalf("hot must count status=1 comments: hot=%d status=%d", got.Hot, got.Status)
	}
}
