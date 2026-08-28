package biz

// 书籍用例：元数据 / 分页列表 / 分类标签 / 多语言翻译（任务 #18）。
// 逻辑从 svc1 移植：GetBook 缓存链 book:{id}:{lang} → DB → 回填（§三.2 / §五）。

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type BookUsecase struct {
	db    *gorm.DB
	cache *data.Cache
}

func NewBookUsecase(d *data.Data) *BookUsecase {
	return &BookUsecase{db: d.DB, cache: d.Cache}
}

type BookItem struct {
	ID      uint64
	Lang    string
	Title   string
	Author  string
	Summary string
	Cover   string
	IsVip   int8
	Status  int8
}

// GetBook：缓存 → DB → 回填（§三.2）。翻译缺失回落原语言；书不存在缓存空值 60s 防穿透。
func (uc *BookUsecase) GetBook(ctx context.Context, id uint64, lang string) (*BookItem, error) {
	if id == 0 {
		return nil, pkg.ErrBookArg
	}
	key := fmt.Sprintf("book:%d:%s", id, lang)
	v, err := uc.cache.GetOrLoad(ctx, key, func() (string, error) {
		var b data.Book
		if err := uc.db.WithContext(ctx).First(&b, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", nil // 空值缓存 60s
			}
			return "", err
		}
		out := BookItem{ID: b.ID, Lang: b.Lang, Title: b.Title, Author: b.Author,
			Summary: b.Summary, Cover: b.Cover, IsVip: b.IsVip, Status: b.Status}
		var t data.BookTranslation
		err := uc.db.WithContext(ctx).
			Where("book_id = ? AND lang = ?", id, lang).First(&t).Error
		if err == nil {
			out.Lang, out.Title, out.Summary, out.Cover = t.Lang, t.Title, t.Summary, t.Cover
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		buf, err := json.Marshal(out)
		return string(buf), err
	})
	if err != nil {
		return nil, pkg.ErrBookInternal
	}
	if v == "" || v == data.EmptyMarker {
		return nil, pkg.ErrBookNF
	}
	var out BookItem
	if err := json.Unmarshal([]byte(v), &out); err != nil {
		return nil, pkg.ErrBookInternal
	}
	return &out, nil
}

// ListBooks 分页列表，支持 lang 本地化标题、category_id/tag_id/status 过滤。
func (uc *BookUsecase) ListBooks(ctx context.Context, p pkg.Page, lang string, categoryID, tagID uint64, status int32) ([]BookItem, int64, error) {
	q := uc.db.WithContext(ctx).Model(&data.Book{})
	if categoryID > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM novel_book_category bc WHERE bc.book_id = novel_book.id AND bc.category_id = ?)", categoryID)
	}
	if tagID > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM novel_book_tag bt WHERE bt.book_id = novel_book.id AND bt.tag_id = ?)", tagID)
	}
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, pkg.ErrBookInternal
	}
	var books []data.Book
	if err := q.Order("id DESC").Offset(p.Offset()).Limit(p.PageSize).Find(&books).Error; err != nil {
		return nil, 0, pkg.ErrBookInternal
	}
	// 批量取翻译做本地化，避免 N+1
	ts := map[uint64]data.BookTranslation{}
	if lang != "" && len(books) > 0 {
		ids := make([]uint64, 0, len(books))
		for _, b := range books {
			ids = append(ids, b.ID)
		}
		var list []data.BookTranslation
		uc.db.WithContext(ctx).Where("book_id IN ? AND lang = ?", ids, lang).Find(&list)
		for _, t := range list {
			ts[t.BookID] = t
		}
	}
	items := make([]BookItem, 0, len(books))
	for _, b := range books {
		it := BookItem{ID: b.ID, Lang: b.Lang, Title: b.Title, Author: b.Author,
			Summary: b.Summary, Cover: b.Cover, IsVip: b.IsVip, Status: b.Status}
		if t, ok := ts[b.ID]; ok {
			it.Lang, it.Title, it.Summary, it.Cover = t.Lang, t.Title, t.Summary, t.Cover
		}
		items = append(items, it)
	}
	return items, total, nil
}

type CreateBookParams struct {
	Title        string
	Author       string
	Summary      string
	Cover        string
	Lang         string
	IsVip        int8
	CategoryIDs  []uint64
	TagIDs       []uint64
	Translations []TranslationParams
}

type TranslationParams struct {
	Lang    string
	Title   string
	Summary string
	Cover   string
}

// CreateBook 建书（作者角色，RBAC 由 service 层校验 claims）。
func (uc *BookUsecase) CreateBook(ctx context.Context, req CreateBookParams) (id uint64, err error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Lang = strings.TrimSpace(req.Lang)
	if req.Title == "" || len(req.Title) > 255 || len(req.Lang) > 5 {
		return 0, pkg.ErrBookArg
	}
	if req.Lang == "" {
		req.Lang = "zh-CN"
	}
	b := data.Book{Title: req.Title, Author: strings.TrimSpace(req.Author), Summary: req.Summary,
		Cover: req.Cover, Lang: req.Lang, IsVip: req.IsVip, Status: 1}
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&b).Error; err != nil {
			return err
		}
		for _, cid := range req.CategoryIDs {
			if err := tx.Create(&data.BookCategory{BookID: b.ID, CategoryID: cid}).Error; err != nil {
				return err
			}
		}
		for _, tid := range req.TagIDs {
			if err := tx.Create(&data.BookTag{BookID: b.ID, TagID: tid}).Error; err != nil {
				return err
			}
		}
		for _, tr := range req.Translations {
			if err := upsertTranslation(tx, b.ID, tr.Lang, tr.Title, tr.Summary, tr.Cover); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, pkg.ErrBookInternal
	}
	return b.ID, nil
}

func (uc *BookUsecase) UpsertTranslation(ctx context.Context, id uint64, lang, title, summary, cover string) error {
	lang = strings.TrimSpace(lang)
	title = strings.TrimSpace(title)
	if id == 0 || len(lang) == 0 || len(lang) > 5 || title == "" || len(title) > 255 {
		return pkg.ErrBookArg
	}
	var cnt int64
	if err := uc.db.WithContext(ctx).Model(&data.Book{}).Where("id = ?", id).Count(&cnt).Error; err != nil || cnt == 0 {
		return pkg.ErrBookNF
	}
	if err := upsertTranslation(uc.db, id, lang, title, summary, cover); err != nil {
		return pkg.ErrBookInternal
	}
	// 翻译更新后立即失效缓存键，避免旧值残留
	uc.cache.Del(ctx, fmt.Sprintf("book:%d:%s", id, lang))
	return nil
}

// upsertTranslation 按 uk_book_lang(book_id, lang) 插入或更新。
func upsertTranslation(db *gorm.DB, bookID uint64, lang, title, summary, cover string) error {
	return db.Model(&data.BookTranslation{}).
		Where("book_id = ? AND lang = ?", bookID, lang).
		Assign(data.BookTranslation{Title: title, Summary: summary, Cover: cover}).
		FirstOrCreate(&data.BookTranslation{BookID: bookID, Lang: lang, Title: title, Summary: summary, Cover: cover}).
		Error
}

type CategoryItem struct {
	ID       uint64
	Name     string
	ParentID uint64
}

func (uc *BookUsecase) ListCategories(ctx context.Context) ([]CategoryItem, error) {
	var cats []data.Category
	if err := uc.db.WithContext(ctx).Where("status = 1").
		Order("sort_order ASC, id ASC").Find(&cats).Error; err != nil {
		return nil, pkg.ErrBookInternal
	}
	items := make([]CategoryItem, 0, len(cats))
	for _, c := range cats {
		items = append(items, CategoryItem{ID: c.ID, Name: c.Name, ParentID: c.ParentID})
	}
	return items, nil
}

type TagItem struct {
	ID   uint64
	Name string
	Lang string
}

func (uc *BookUsecase) ListTags(ctx context.Context, lang string) ([]TagItem, error) {
	q := uc.db.WithContext(ctx).Model(&data.Tag{}).Where("status = 1")
	if lang != "" {
		q = q.Where("lang = ?", lang)
	}
	var tags []data.Tag
	if err := q.Order("id ASC").Find(&tags).Error; err != nil {
		return nil, pkg.ErrBookInternal
	}
	items := make([]TagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, TagItem{ID: t.ID, Name: t.Name, Lang: t.Lang})
	}
	return items, nil
}
