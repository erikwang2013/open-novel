package biz

// 搜索用例：OpenSearch 多语言检索 / 热门 / 索引同步（任务 #21）。
// 逻辑从 svc1 移植：索引消失自动重建重试、搜索日志 best-effort、热门榜缓存 search:hot。

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type SearchUsecase struct {
	db    *gorm.DB
	cache *data.Cache
	es    *data.ES
}

func NewSearchUsecase(d *data.Data) *SearchUsecase {
	return &SearchUsecase{db: d.DB, cache: d.Cache, es: d.ES}
}

// EnsureIndex 启动时建索引（幂等）。
func (uc *SearchUsecase) EnsureIndex(ctx context.Context) error {
	return uc.es.EnsureIndex(ctx)
}

// esSearch 查询失败且因索引消失时，重建索引后重试一次。
func (uc *SearchUsecase) esSearch(f func() ([]data.SearchDoc, int64, error)) ([]data.SearchDoc, int64, error) {
	docs, total, err := f()
	if err == nil || !strings.Contains(err.Error(), "no such index") {
		return docs, total, err
	}
	if e := uc.es.EnsureIndex(context.Background()); e != nil {
		return nil, 0, err
	}
	return f()
}

// Search 检索并写搜索日志（best-effort，失败不阻塞搜索）。
func (uc *SearchUsecase) Search(ctx context.Context, q, lang string, p pkg.Page, uid int64, ip string) ([]data.SearchDoc, int64, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, 0, pkg.ErrSearchArg
	}
	if lang == "" {
		lang = "zh-CN"
	}
	docs, total, err := uc.esSearch(func() ([]data.SearchDoc, int64, error) {
		return uc.es.Search(ctx, q, lang, p.Offset(), p.PageSize)
	})
	if err != nil {
		return nil, 0, pkg.ErrSearch
	}
	var uidPtr *int64
	if uid > 0 {
		uidPtr = &uid
	}
	sl := data.SearchLog{UserID: uidPtr, Keyword: q, Lang: lang, ResultCount: uint32(total), IP: ip}
	if e := uc.db.WithContext(ctx).Create(&sl).Error; e != nil {
		log.Printf("search: log: %v", e)
	}
	return docs, total, nil
}

// Hot 热门榜（缓存 search:hot）。
func (uc *SearchUsecase) Hot(ctx context.Context) ([]data.SearchDoc, int64, error) {
	payload, err := uc.cache.GetOrLoad(ctx, "search:hot", func() (string, error) {
		docs, err := uc.es.Hot(ctx, 50)
		if err != nil {
			return "", err
		}
		b, err := json.Marshal(map[string]any{
			"list": docs, "total": len(docs), "page": 1, "page_size": len(docs),
		})
		return string(b), err
	})
	if err != nil {
		return nil, 0, pkg.ErrSearch
	}
	var out struct {
		List  []data.SearchDoc `json:"list"`
		Total int64            `json:"total"`
	}
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, 0, pkg.ErrSearch
	}
	return out.List, out.Total, nil
}

// HotKeywords 热搜词榜（搜索日志按词聚合，TOP 10，口径同 AdminUsecase.Stats）。
func (uc *SearchUsecase) HotKeywords(ctx context.Context) ([]HotKeyword, error) {
	var out []HotKeyword
	if err := uc.db.WithContext(ctx).Model(&data.SearchLog{}).
		Select("keyword, COUNT(*) AS count").
		Group("keyword").Order("count DESC").Limit(10).Scan(&out).Error; err != nil {
		return nil, pkg.ErrSearch
	}
	return out, nil
}

// Suggest 搜索建议：搜索日志按 keyword 前缀补全（LIKE 参数绑定防注入），
// 按热度 count 降序 TOP 10；缓存 suggest:{q}，TTL 1s。
func (uc *SearchUsecase) Suggest(ctx context.Context, q string) ([]string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	key := "suggest:" + q
	if v, ok := uc.cache.Get(ctx, key); ok {
		var cached []string
		if err := json.Unmarshal([]byte(v), &cached); err == nil {
			return cached, nil
		}
	}
	var rows []HotKeyword
	if err := uc.db.WithContext(ctx).Model(&data.SearchLog{}).
		Where("keyword LIKE ?", q+"%").
		Select("keyword, COUNT(*) AS count").
		Group("keyword").Order("count DESC").Limit(10).Scan(&rows).Error; err != nil {
		return nil, pkg.ErrSearch
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Keyword)
	}
	b, _ := json.Marshal(out)
	uc.cache.Set(ctx, key, string(b), time.Second)
	return out, nil
}

// SyncIndex 同步书籍文档（书籍服务在新建/更新书籍或翻译时调用）。
func (uc *SearchUsecase) SyncIndex(ctx context.Context, doc data.SearchDoc) error {
	if doc.BookID == 0 {
		return pkg.ErrSearchArg
	}
	if doc.CreatedAt == "" {
		doc.CreatedAt = time.Now().Format("2006-01-02")
	}
	if err := uc.es.Upsert(ctx, doc); err != nil {
		return pkg.ErrSearch
	}
	return nil
}

func (uc *SearchUsecase) DeleteIndex(ctx context.Context, bookID uint64) error {
	if bookID == 0 {
		return pkg.ErrSearchArg
	}
	// 幂等：文档已删除（result=not_found）不算错，书籍服务可能重试
	if err := uc.es.Delete(ctx, bookID); err != nil && !strings.Contains(err.Error(), `"result":"not_found"`) && !strings.Contains(err.Error(), `"status":404`) {
		return pkg.ErrSearch
	}
	return nil
}
