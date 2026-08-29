package biz

// 推荐用例：热门/最新榜单（任务 #15）+ AI 本地启发式个性化（Phase 5 本地版，无第三方 API）。
// 逻辑从 svc1 移植：hot 按评论数排名 / new 按 id 倒序，缓存 recommend:{strategy}:{page}:{page_size}；
// ai 按用户行为画像评分（纯本地计算，第三方离线训练预留，见 rankAI 注释），缓存 rec:ai:{uid}。

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"

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

// List 榜单；strategy = hot | new | ai。
func (uc *RecommendUsecase) List(ctx context.Context, strategy string, p pkg.Page) ([]RecommendItem, int64, error) {
	if strategy == "" {
		strategy = "hot"
	}
	if strategy != "hot" && strategy != "new" && strategy != "ai" {
		return nil, 0, pkg.ErrRecommendArg
	}
	key := "recommend:" + strategy + ":" + strconv.Itoa(p.Page) + ":" + strconv.Itoa(p.PageSize)
	rankFn := func() ([]RecommendItem, int64, error) { return uc.rank(ctx, strategy, p) }
	if strategy == "ai" {
		// 按用户画像缓存；lang 无请求字段，语言规则同 hot 不额外过滤。
		// TTL 走 GetOrLoad 5min±30s 抖动，短 TTL 兜底行为变化。
		uid := pkg.ClaimsFrom(ctx).UID
		key = "rec:ai:" + strconv.FormatInt(uid, 10)
		rankFn = func() ([]RecommendItem, int64, error) { return uc.rankAI(ctx, uid, p) }
	}
	payload, err := uc.cache.GetOrLoad(ctx, key, func() (string, error) {
		items, total, err := rankFn()
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

// aiHotExpr AI 热度排序依据（与 hot 策略同源）：书籍评论数（暂无浏览量列）。
const aiHotExpr = "(SELECT COUNT(*) FROM novel_comment c WHERE c.book_id = b.id)"

// rankAI ai 策略：基于用户行为画像的启发式个性化推荐，纯本地计算（不接任何第三方 API）。
// 候选评分 = 画像分类/标签重叠数 + 热度；排除已收藏/已读/书架的书；无画像或候选 <5 回退 hot。
// ponytail: 按分类/标签重叠的启发式评分，未来接第三方离线训练时仅替换 score 函数；
// 关键词子串匹配字典表（分类/标签行数少，全量加载），表增长后换分词/FTS；候选上限 500。
func (uc *RecommendUsecase) rankAI(ctx context.Context, uid int64, p pkg.Page) ([]RecommendItem, int64, error) {
	if uid <= 0 {
		return uc.rank(ctx, "hot", p)
	}
	cats, tags, err := uc.profile(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	if len(cats) == 0 && len(tags) == 0 {
		return uc.rank(ctx, "hot", p)
	}
	catIDs := make([]uint64, 0, len(cats))
	for id := range cats {
		catIDs = append(catIDs, id)
	}
	tagIDs := make([]uint64, 0, len(tags))
	for id := range tags {
		tagIDs = append(tagIDs, id)
	}
	rows, err := uc.db.WithContext(ctx).Table("novel_book b").
		Select("b.id AS book_id, b.title, b.author, b.summary, b.cover, b.lang, " +
			"(COUNT(DISTINCT bc.category_id) + COUNT(DISTINCT bt.tag_id)) AS score, " + aiHotExpr + " AS hot").
		Joins("LEFT JOIN novel_book_category bc ON bc.book_id = b.id AND bc.category_id IN ?", catIDs).
		Joins("LEFT JOIN novel_book_tag bt ON bt.book_id = b.id AND bt.tag_id IN ?", tagIDs).
		Where("b.status = 1").
		Where("b.id NOT IN (SELECT book_id FROM novel_favorite WHERE user_id = ?)", uid).
		Where("b.id NOT IN (SELECT book_id FROM novel_reading_progress WHERE user_id = ?)", uid).
		Where("b.id NOT IN (SELECT book_id FROM novel_bookshelf WHERE user_id = ?)", uid).
		Group("b.id").
		Having("score > 0").
		Order("score DESC, hot DESC, b.id DESC").
		Limit(500).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	all := make([]RecommendItem, 0, 64)
	for rows.Next() {
		var it RecommendItem
		var score, hot int
		if err := rows.Scan(&it.BookID, &it.Title, &it.Author, &it.Summary, &it.Cover, &it.Lang, &score, &hot); err != nil {
			return nil, 0, err
		}
		it.Strategy = "ai"
		all = append(all, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(all) < 5 {
		return uc.rank(ctx, "hot", p) // 候选不足回退 hot（空数据也走这里，不报错）
	}
	start := p.Offset()
	if start > len(all) {
		start = len(all)
	}
	end := start + p.PageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], int64(len(all)), nil
}

// profile 构建用户画像：最近搜索词匹配分类/标签；行为书籍（曝光/收藏/阅读/书架）取分类/标签。
func (uc *RecommendUsecase) profile(ctx context.Context, uid int64) (map[uint64]bool, map[uint64]bool, error) {
	cats := make(map[uint64]bool)
	tags := make(map[uint64]bool)

	var kws []string
	if err := uc.db.WithContext(ctx).Model(&data.SearchLog{}).
		Where("user_id = ?", uid).Order("id DESC").Limit(20).Pluck("keyword", &kws).Error; err != nil {
		return nil, nil, err
	}
	if len(kws) > 0 {
		// ponytail: 字典表小（几百行内），全量匹配比逐词 LIKE 简单
		var cs []data.Category
		if err := uc.db.WithContext(ctx).Where("status = 1").Find(&cs).Error; err != nil {
			return nil, nil, err
		}
		for _, c := range cs {
			for _, kw := range kws {
				if kw != "" && strings.Contains(c.Name, kw) {
					cats[c.ID] = true
					break
				}
			}
		}
		var ts []data.Tag
		if err := uc.db.WithContext(ctx).Where("status = 1").Find(&ts).Error; err != nil {
			return nil, nil, err
		}
		for _, t := range ts {
			for _, kw := range kws {
				if kw != "" && strings.Contains(t.Name, kw) {
					tags[t.ID] = true
					break
				}
			}
		}
	}

	seen := make(map[uint64]bool)
	var bookIDs []uint64
	for _, tbl := range []string{"novel_recommend_log", "novel_reading_progress", "novel_bookshelf", "novel_favorite"} {
		var ids []uint64
		if err := uc.db.WithContext(ctx).Table(tbl).
			Where("user_id = ?", uid).Order("id DESC").Limit(50).Pluck("book_id", &ids).Error; err != nil {
			return nil, nil, err
		}
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				bookIDs = append(bookIDs, id)
			}
		}
	}
	if len(bookIDs) > 0 {
		var cids []uint64
		if err := uc.db.WithContext(ctx).Table("novel_book_category").
			Where("book_id IN ?", bookIDs).Distinct().Pluck("category_id", &cids).Error; err != nil {
			return nil, nil, err
		}
		for _, id := range cids {
			cats[id] = true
		}
		var tids []uint64
		if err := uc.db.WithContext(ctx).Table("novel_book_tag").
			Where("book_id IN ?", bookIDs).Distinct().Pluck("tag_id", &tids).Error; err != nil {
			return nil, nil, err
		}
		for _, id := range tids {
			tags[id] = true
		}
	}
	return cats, tags, nil
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

