package service

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	recommendv1 "open-novel/backend/api/recommendation/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func newTestRecommend(t *testing.T) (*RecommendationService, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewRecommendationService(biz.NewRecommendUsecase(d)), d
}

// TestRecommendFlow: insert book → hot/new lists → invalid strategy → log rows (logged-in only).
func TestRecommendFlow(t *testing.T) {
	s, d := newTestRecommend(t)
	ctx := context.Background()

	b := data.Book{Title: "svc3 推荐测试书", Author: "svc3", Summary: "推荐测试", Lang: "zh-CN", Status: 1}
	if err := d.DB.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.Where("book_id = ?", b.ID).Delete(&data.RecommendLog{})
		d.DB.Delete(&data.Book{}, b.ID)
	})
	d.Cache.Del(ctx, "recommend:hot:1:20", "recommend:new:1:20")

	for _, st := range []string{"hot", "new"} {
		rep, err := s.GetRecommendations(ctx, &recommendv1.GetRecommendationsReq{Strategy: st, Page: 1, PageSize: 20})
		if err != nil {
			t.Fatalf("%s: %v", st, err)
		}
		found := false
		for _, it := range rep.List {
			if it.BookId == int64(b.ID) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s list should contain inserted book %d", st, b.ID)
		}
	}

	// invalid strategy → 170400
	if _, err := s.GetRecommendations(ctx, &recommendv1.GetRecommendationsReq{Strategy: "bogus"}); err == nil {
		t.Fatal("expected error for strategy=bogus")
	} else if kerrors.FromError(err).Code != 170400 {
		t.Fatalf("want 170400, got %d", kerrors.FromError(err).Code)
	}

	// anonymous: no log rows; logged-in (UID 42): rows written
	var n int64
	d.DB.Model(&data.RecommendLog{}).Where("book_id = ?", b.ID).Count(&n)
	if n != 0 {
		t.Fatalf("anonymous request must not log impressions, got %d rows", n)
	}
	ctx = pkg.WithClaims(ctx, pkg.Claims{UID: 42})
	d.Cache.Del(ctx, "recommend:hot:1:20")
	if _, err := s.GetRecommendations(ctx, &recommendv1.GetRecommendationsReq{Strategy: "hot", Page: 1, PageSize: 20}); err != nil {
		t.Fatal(err)
	}
	if e := d.DB.Model(&data.RecommendLog{}).Where("book_id = ? AND user_id = 42", b.ID).Count(&n).Error; e != nil || n == 0 {
		t.Fatalf("logged-in impression log missing (n=%d, err=%v)", n, e)
	}
}
