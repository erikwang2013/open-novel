package biz

import (
	"context"
	"fmt"
	"testing"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func newTestRecommendBiz(t *testing.T) (*RecommendUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewRecommendUsecase(d), d
}

// TestRecommendAI: 画像→评分排序+排除、无画像/候选不足回退 hot、缓存命中。
func TestRecommendAI(t *testing.T) {
	uc, d := newTestRecommendBiz(t)
	ctx := context.Background()

	tagSci := data.Tag{Name: "AI测试科幻", Lang: "zh-CN", Status: 1}
	tagFanta := data.Tag{Name: "AI测试玄幻", Lang: "zh-CN", Status: 1}
	if err := d.DB.Create(&tagSci).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create(&tagFanta).Error; err != nil {
		t.Fatal(err)
	}

	var b []data.Book
	for i := 0; i < 7; i++ {
		book := data.Book{Title: fmt.Sprintf("ai测试书%d", i), Author: "ai-test", Summary: "x", Lang: "zh-CN", Status: 1}
		if err := d.DB.Create(&book).Error; err != nil {
			t.Fatal(err)
		}
		b = append(b, book)
	}
	// A1-A4、A6 科幻；A5 科幻+玄幻（重叠数 2，应排最前）；A7 无标签
	for _, i := range []int{0, 1, 2, 3, 5} {
		if err := d.DB.Create(&data.BookTag{BookID: b[i].ID, TagID: tagSci.ID}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := d.DB.Create(&data.BookTag{BookID: b[4].ID, TagID: tagFanta.ID}).Error; err != nil {
		t.Fatal(err)
	}

	// 用户 999：搜索「AI测试科幻」「AI测试玄幻」+ 收藏 A6（应被排除）
	uid := int64(999)
	if err := d.DB.Create(&data.SearchLog{UserID: &uid, Keyword: "AI测试科幻", Lang: "zh-CN"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create(&data.SearchLog{UserID: &uid, Keyword: "AI测试玄幻", Lang: "zh-CN"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.Create(&data.Favorite{UserID: uint64(uid), BookID: b[5].ID}).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ids := make([]uint64, 0, len(b))
		for _, x := range b {
			ids = append(ids, x.ID)
		}
		d.DB.Where("book_id IN ?", ids).Delete(&data.BookTag{})
		d.DB.Delete(&data.Book{}, ids)
		d.DB.Delete(&data.Tag{}, tagSci.ID, tagFanta.ID)
		d.DB.Where("user_id IN ?", []int64{999, 998, 997}).Delete(&data.SearchLog{})
		d.DB.Where("user_id = ?", uid).Delete(&data.Favorite{})
		d.DB.Where("user_id = ?", uid).Delete(&data.RecommendLog{})
		d.DB.Where("user_id = ?", uid).Delete(&data.ReadingProgress{})
		d.DB.Where("user_id = ?", uid).Delete(&data.Bookshelf{})
	})

	for _, k := range []string{"rec:ai:999", "rec:ai:998", "rec:ai:997", "recommend:hot:1:20"} {
		d.Cache.Del(ctx, k)
	}

	// 1) 画像→评分排序：A5（双标签）在前，A1-A4 随 id 倒序，A6（已收藏）/A7（无重叠）排除
	items, total, err := uc.List(pkg.WithClaims(ctx, pkg.Claims{UID: uid}), "ai", pkg.Page{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	want := []int{4, 3, 2, 1, 0}
	if len(items) != len(want) || total != int64(len(want)) {
		t.Fatalf("personalized: want %d items total=%d, got %d total=%d", len(want), len(want), len(items), total)
	}
	for i, idx := range want {
		if items[i].BookID != b[idx].ID {
			t.Fatalf("personalized[%d]: want book %d, got %d", i, b[idx].ID, items[i].BookID)
		}
	}
	if items[0].Strategy != "ai" {
		t.Fatalf("strategy field: want ai, got %s", items[0].Strategy)
	}

	// 2) 缓存命中：再次调用结果一致，且 rec:ai:999 已写入
	again, _, err := uc.List(pkg.WithClaims(ctx, pkg.Claims{UID: uid}), "ai", pkg.Page{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != len(items) || again[0].BookID != items[0].BookID {
		t.Fatalf("cached: want first %d len %d, got first %d len %d", items[0].BookID, len(items), again[0].BookID, len(again))
	}
	if v, ok := d.Cache.Get(ctx, "rec:ai:999"); !ok || v == "" {
		t.Fatal("rec:ai:999 should be cached after first call")
	}

	// 3) 无画像（新用户）→ 回退 hot，与 hot 榜单完全一致
	hotItems, _, err := uc.List(ctx, "hot", pkg.Page{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	fresh, _, err := uc.List(pkg.WithClaims(ctx, pkg.Claims{UID: 998}), "ai", pkg.Page{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) != len(hotItems) {
		t.Fatalf("no-profile fallback: want len %d, got %d", len(hotItems), len(fresh))
	}
	for i := range fresh {
		if fresh[i].BookID != hotItems[i].BookID {
			t.Fatalf("no-profile fallback[%d]: want %d, got %d", i, hotItems[i].BookID, fresh[i].BookID)
		}
	}

	// 4) 候选不足（<5）：用户 997 仅搜「AI测试玄幻」→ 只命中 A5 → 回退 hot
	u997 := int64(997)
	if err := d.DB.Create(&data.SearchLog{UserID: &u997, Keyword: "AI测试玄幻", Lang: "zh-CN"}).Error; err != nil {
		t.Fatal(err)
	}
	d.Cache.Del(ctx, "rec:ai:997")
	sparse, _, err := uc.List(pkg.WithClaims(ctx, pkg.Claims{UID: 997}), "ai", pkg.Page{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(sparse) != len(hotItems) {
		t.Fatalf("sparse fallback: want len %d, got %d", len(hotItems), len(sparse))
	}
	for i := range sparse {
		if sparse[i].BookID != hotItems[i].BookID {
			t.Fatalf("sparse fallback[%d]: want %d, got %d", i, hotItems[i].BookID, sparse[i].BookID)
		}
	}
}
