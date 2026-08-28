package service

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	searchv1 "open-novel/backend/api/search/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func newTestSearch(t *testing.T) (*SearchService, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewSearchService(biz.NewSearchUsecase(d)), d
}

// TestSearchFlow: sync → ja/zh search → hot → delete (idempotent) → log row.
func TestSearchFlow(t *testing.T) {
	s, d := newTestSearch(t)
	ctx := context.Background()
	const bid = 900002

	// invalid q → 160400
	if _, err := s.SearchBooks(ctx, &searchv1.SearchBooksReq{Q: "  "}); err == nil {
		t.Fatal("expected error for empty q")
	} else if kerrors.FromError(err).Code != 160400 {
		t.Fatalf("want 160400, got %d", kerrors.FromError(err).Code)
	}

	// index writes require auth + author role (审查 P1)
	req := &searchv1.SyncIndexReq{
		BookId: bid, Lang: "ja", Status: 1, Hot: 1_000_000_000, CreatedAt: "2026-08-28",
		TitleZh: "星海旅人", TitleEn: "Star Sea Wanderer", TitleJa: "星海の旅人",
		SummaryZh: "星际旅行者的冒险", SummaryJa: "仮想世界の冒険譚",
		AuthorZh: "svc3测试", AuthorJa: "svc3テスト",
	}
	if _, err := s.SyncIndex(ctx, req); kerrors.FromError(err).Code != 140401 {
		t.Fatalf("anonymous sync: want 140401, got %d", kerrors.FromError(err).Code)
	}
	userCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 42, Role: 1})
	if _, err := s.SyncIndex(userCtx, req); kerrors.FromError(err).Code != 140403 {
		t.Fatalf("reader sync: want 140403, got %d", kerrors.FromError(err).Code)
	}
	authorCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 42, Role: 2})
	if _, err := s.SyncIndex(authorCtx, req); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.DeleteIndex(authorCtx, &searchv1.DeleteIndexReq{BookId: bid}) })

	// ja search (kuromoji); 索引可能含真实书籍，断言 fixture 命中而非总数
	ja, err := s.SearchBooks(ctx, &searchv1.SearchBooksReq{Q: "仮想世界", Lang: "ja"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range ja.List {
		if b.BookId == bid {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ja search: want %d in hits, got total=%d", bid, ja.Total)
	}

	// zh search (standard tokenizer char matching); 索引可能含真实书籍，断言 fixture 命中而非总数
	zh, err := s.SearchBooks(ctx, &searchv1.SearchBooksReq{Q: "星海旅人"})
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, b := range zh.List {
		if b.BookId == bid {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("zh search: want %d in hits, got total=%d", bid, zh.Total)
	}

	// hot list (uncached, high hot score)
	d.Cache.Del(ctx, "search:hot")
	hot, err := s.HotSearches(ctx, &searchv1.HotSearchesReq{})
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, b := range hot.List {
		if b.BookId == bid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("hot list should contain synced book")
	}

	// search log row written
	var n int64
	if e := d.DB.Model(&data.SearchLog{}).Where("keyword = ?", "仮想世界").Count(&n).Error; e != nil || n == 0 {
		t.Fatalf("search log missing (n=%d, err=%v)", n, e)
	}

	// delete → gone; delete again is idempotent
	if _, err := s.DeleteIndex(authorCtx, &searchv1.DeleteIndexReq{BookId: bid}); err != nil {
		t.Fatal(err)
	}
	after, err := s.SearchBooks(ctx, &searchv1.SearchBooksReq{Q: "仮想世界", Lang: "ja"})
	if err != nil {
		t.Fatal(err)
	}
	if after.Total != 0 {
		t.Fatalf("after delete: want 0 hits, got %d", after.Total)
	}
	if _, err := s.DeleteIndex(authorCtx, &searchv1.DeleteIndexReq{BookId: bid}); err != nil {
		t.Fatalf("second delete should be idempotent: %v", err)
	}
}
