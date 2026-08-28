package biz

import (
	"testing"

	"open-novel/backend/internal/data"
)

func TestSetLang(t *testing.T) {
	// 与 syncIndex 调用顺序一致：先基础语言，再翻译覆盖；未列语言归 zh 位。
	var d data.SearchDoc
	setLang(&d, "fr", "Français", "FR") // 基础语言 fr → zh 位
	setLang(&d, "zh-CN", "中文标题", "中文简介") // 翻译覆盖 zh 位
	setLang(&d, "en", "English Title", "English Summary")
	setLang(&d, "ja", "日本語タイトル", "日本語概要")
	if d.TitleZh != "中文标题" || d.SummaryZh != "中文简介" {
		t.Fatalf("zh translation should override base, got zh=%q/%q", d.TitleZh, d.SummaryZh)
	}
	if d.TitleEn != "English Title" || d.SummaryEn != "English Summary" {
		t.Fatalf("en wrong: %q/%q", d.TitleEn, d.SummaryEn)
	}
	if d.TitleJa != "日本語タイトル" || d.SummaryJa != "日本語概要" {
		t.Fatalf("ja wrong: %q/%q", d.TitleJa, d.SummaryJa)
	}
}
