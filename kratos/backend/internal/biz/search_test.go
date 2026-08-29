package biz

import (
	"context"
	"testing"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func newTestSearchBiz(t *testing.T) (*SearchUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewSearchUsecase(d), d
}

func seedLogs(t *testing.T, d *data.Data, logs []data.SearchLog) {
	t.Helper()
	if err := d.DB.Create(&logs).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.Where("keyword IN ?", []string{"星海旅人", "星海战记", "星云漫游"}).Delete(&data.SearchLog{})
	})
}

// TestSuggest: 空 q / 前缀命中按热度排序 / 无命中返回空。
func TestSuggest(t *testing.T) {
	uc, d := newTestSearchBiz(t)
	ctx := context.Background()
	d.Cache.Del(ctx, "suggest:星海", "suggest:zzzz")

	// 空 q → 空列表，不报错
	words, err := uc.Suggest(ctx, "   ")
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 0 {
		t.Fatalf("empty q: want 0, got %v", words)
	}

	// 造日志：星海旅人 x3、星海战记 x2、星云漫游 x1
	seedLogs(t, d, []data.SearchLog{
		{Keyword: "星海旅人", Lang: "zh-CN", ResultCount: 1}, {Keyword: "星海旅人", Lang: "zh-CN", ResultCount: 1}, {Keyword: "星海旅人", Lang: "zh-CN", ResultCount: 1},
		{Keyword: "星海战记", Lang: "zh-CN", ResultCount: 1}, {Keyword: "星海战记", Lang: "zh-CN", ResultCount: 1},
		{Keyword: "星云漫游", Lang: "zh-CN", ResultCount: 1},
	})

	// 前缀命中：按 count 降序
	words, err = uc.Suggest(ctx, "星海")
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 2 || words[0] != "星海旅人" || words[1] != "星海战记" {
		t.Fatalf("prefix hit: want [星海旅人 星海战记], got %v", words)
	}

	// 缓存命中返回一致结果
	cached, err := uc.Suggest(ctx, "星海")
	if err != nil {
		t.Fatal(err)
	}
	if len(cached) != 2 || cached[0] != words[0] || cached[1] != words[1] {
		t.Fatalf("cached: want %v, got %v", words, cached)
	}

	// 无命中 → 空列表
	words, err = uc.Suggest(ctx, "zzzz")
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 0 {
		t.Fatalf("no hit: want 0, got %v", words)
	}
}
