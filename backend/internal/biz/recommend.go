package biz

// 推荐用例：热门/最新榜单（任务 #15）。
// Phase 1 为 MySQL 榜单 stub，Phase 5 接入 AI 推荐（同一契约）。
// 逻辑从 svc1 移植：hot 按评论数排名 / new 按 id 倒序，缓存 recommend:{strategy}:{page}:{page_size}。

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"gorm.io/gorm"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type RecommendUsecase struct {
	db    *gorm.DB
	cache *data.Cache
}

func NewRecommendUsecase(d *data.Data) *RecommendUsecase {
	return &RecommendUsecase{db: d.DB, cache: d.Cache}
}

type RecommendItem struct {
	BookID   uint64
	Title    string
	Author   string
	Summary  string
	Cover    string
	Lang     string
	Strategy string
}

// List 榜单；strategy = hot | new。
func (uc *RecommendUsecase) List(ctx context.Context, strategy string, p pkg.Page) ([]RecommendItem, int64, error) {
	if strategy == "" {
		strategy = "hot"
	}
	if strategy != "hot" && strategy != "new" {
		return nil, 0, pkg.ErrRecommendArg
	}
	payload, err := uc.cache.GetOrLoad(ctx, "recommend:"+strategy+":"+strconv.Itoa(p.Page)+":"+strconv.Itoa(p.PageSize), func() (string, error) {
		items, total, err := uc.rank(ctx, strategy, p)
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(map[string]any{
			"list": items, "total": total, "page": p.Page, "page_size": p.PageSize,
		})
		return string(b), err
	})
	if err != nil {
		return nil, 0, pkg.ErrRecommend
	}
	var out struct {
		List  []RecommendItem `json:"list"`
		Total int64           `json:"total"`
	}
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, 0, pkg.ErrRecommend
	}
	return out.List, out.Total, nil
}

// rank 查 novel_book；hot 按评论数排名（暂无浏览量列），new 按最新。
func (uc *RecommendUsecase) rank(ctx context.Context, strategy string, p pkg.Page) ([]RecommendItem, int64, error) {
	order := "id DESC"
	if strategy == "hot" {
		// ponytail: COUNT 子查询逐行计评论数；表增长后换预计算热度分列
		order = "(SELECT COUNT(*) FROM novel_comment c WHERE c.book_id = novel_book.id) DESC, id DESC"
	}
	var total int64
	if err := uc.db.WithContext(ctx).Model(&data.Book{}).
		Where("status = 1").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows, err := uc.db.WithContext(ctx).Table("novel_book").
		Select("id AS book_id, title, author, summary, cover, lang").
		Where("status = 1").Order(order).Limit(p.PageSize).Offset(p.Offset()).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]RecommendItem, 0, p.PageSize)
	for rows.Next() {
		var it RecommendItem
		if err := rows.Scan(&it.BookID, &it.Title, &it.Author, &it.Summary, &it.Cover, &it.Lang); err != nil {
			return nil, 0, err
		}
		it.Strategy = strategy
		items = append(items, it)
	}
	return items, total, rows.Err()
}

// Log 记录榜单曝光（仅登录用户；best-effort）。
func (uc *RecommendUsecase) Log(ctx context.Context, uid int64, strategy string, items []RecommendItem) {
	if uid <= 0 {
		return
	}
	for i, it := range items {
		rec := data.RecommendLog{UserID: uid, BookID: it.BookID, Strategy: strategy, RankNo: uint32(i + 1)}
		if err := uc.db.WithContext(ctx).Create(&rec).Error; err != nil {
			log.Printf("recommend: log: %v", err)
			return
		}
	}
}

