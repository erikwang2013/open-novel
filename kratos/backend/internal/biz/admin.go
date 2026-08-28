package biz

// 管理端用例：仪表盘统计 / 分类 / 标签 CRUD（T-A-14/16）。
// 统计走只读副本；写操作 FORCE_MASTER + 审计。

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type AdminUsecase struct {
	db     *gorm.DB
	search *SearchUsecase
}

func NewAdminUsecase(d *data.Data, search *SearchUsecase) *AdminUsecase {
	return &AdminUsecase{db: d.DB, search: search}
}

// Stats 仪表盘统计。
// DAU 口径：当日登录（audit_log action=login）+ 当日搜索（search_log）的去重登录用户，
// 匿名（user_id NULL）无法计入，为近似值。
type Stats struct {
	BookCount    int64
	UserCount    int64
	CommentCount int64
	DAU          int64
	HotBooks     []HotBook
	HotKeywords  []HotKeyword
}

type HotBook struct {
	BookID uint64
	Title  string
	Hot    int64
}

type HotKeyword struct {
	Keyword string
	Count   int64
}

func (uc *AdminUsecase) Stats(ctx context.Context) (Stats, error) {
	var s Stats
	today := time.Now().Format("2006-01-02") // created_at >= '2026-08-29' 即当天 00:00
	for _, c := range []struct {
		dst *int64
		m   any
	}{
		{&s.BookCount, &data.Book{}},
		{&s.UserCount, &data.User{}},
		{&s.CommentCount, &data.Comment{}},
	} {
		if err := uc.db.WithContext(ctx).Model(c.m).Count(c.dst).Error; err != nil {
			return s, pkg.ErrAdminDB
		}
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM (
			SELECT user_id FROM novel_audit_log WHERE user_id IS NOT NULL AND action = 'login' AND created_at >= ?
			UNION
			SELECT user_id FROM novel_search_log WHERE user_id IS NOT NULL AND created_at >= ?
		) t`, today, today).Scan(&s.DAU).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	// 热门书籍：复用搜索热门榜（缓存 search:hot），失败不阻断统计
	if docs, _, err := uc.search.Hot(ctx); err == nil {
		for _, d := range docs {
			if len(s.HotBooks) >= 10 {
				break
			}
			title := d.TitleZh
			if title == "" {
				title = d.TitleEn
			}
			if title == "" {
				title = d.TitleJa
			}
			if title == "" {
				title = d.TitleKo
			}
			s.HotBooks = append(s.HotBooks, HotBook{BookID: d.BookID, Title: title, Hot: d.Hot})
		}
	}
	// 热门搜索词：搜索日志按词聚合
	if err := uc.db.WithContext(ctx).Model(&data.SearchLog{}).
		Select("keyword, COUNT(*) AS count").
		Group("keyword").Order("count DESC").Limit(10).Scan(&s.HotKeywords).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	return s, nil
}

// ListCategories 全量分类（量小不分页），按父级 + 排序返回。
func (uc *AdminUsecase) ListCategories(ctx context.Context) ([]data.Category, error) {
	var list []data.Category
	if err := uc.db.WithContext(ctx).Order("parent_id, sort_order, id").Find(&list).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	return list, nil
}

func (uc *AdminUsecase) CreateCategory(ctx context.Context, name string, parentID uint64, sortOrder int) (*data.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return nil, pkg.ErrInvalidArgument
	}
	if err := uc.checkParent(ctx, parentID); err != nil {
		return nil, err
	}
	c := data.Category{Name: name, ParentID: parentID, SortOrder: sortOrder, Status: 1}
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&c).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	return &c, nil
}

// UpdateCategory 字段可选，仅更新非空项；status 变更写审计。
func (uc *AdminUsecase) UpdateCategory(ctx context.Context, adminID int64, id uint64, req CategoryUpdate) (*data.Category, error) {
	updates := map[string]any{}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" || len(n) > 64 {
			return nil, pkg.ErrInvalidArgument
		}
		updates["name"] = n
	}
	if req.ParentID != nil {
		if err := uc.checkParent(ctx, *req.ParentID); err != nil {
			return nil, err
		}
		updates["parent_id"] = *req.ParentID
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			return nil, pkg.ErrBadState
		}
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.Category{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	if req.Status != nil {
		data.WriteAudit(uc.db, ctx, adminID, "category_status", "category", strconv.FormatUint(id, 10), "status="+strconv.Itoa(int(*req.Status)))
	}
	// FORCE_MASTER: 写后读
	var c data.Category
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	return &c, nil
}

// DeleteCategory 硬删除；存在子分类或未找到时报错，并清理书籍-分类关联。
func (uc *AdminUsecase) DeleteCategory(ctx context.Context, adminID int64, id uint64) error {
	var children int64
	if err := uc.db.WithContext(ctx).Model(&data.Category{}).Where("parent_id = ?", id).Count(&children).Error; err != nil {
		return pkg.ErrAdminDB
	}
	if children > 0 {
		return pkg.ErrBadState // 存在子分类，先处理子分类
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("id = ?", id).Delete(&data.Category{})
	if res.Error != nil {
		return pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("category_id = ?", id).Delete(&data.BookCategory{})
	data.WriteAudit(uc.db, ctx, adminID, "category_delete", "category", strconv.FormatUint(id, 10), "")
	return nil
}

// ListTags 全量标签。
func (uc *AdminUsecase) ListTags(ctx context.Context) ([]data.Tag, error) {
	var list []data.Tag
	if err := uc.db.WithContext(ctx).Order("id").Find(&list).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	return list, nil
}

func (uc *AdminUsecase) CreateTag(ctx context.Context, name, lang string) (*data.Tag, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 64 {
		return nil, pkg.ErrInvalidArgument
	}
	if lang == "" {
		lang = "zh-CN"
	}
	t := data.Tag{Name: name, Lang: lang, Status: 1}
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&t).Error; err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return nil, pkg.ErrConflict // uk_name_lang 重复
		}
		return nil, pkg.ErrAdminDB
	}
	return &t, nil
}

// UpdateTag 字段可选；status 变更写审计。
func (uc *AdminUsecase) UpdateTag(ctx context.Context, adminID int64, id uint64, req TagUpdate) (*data.Tag, error) {
	updates := map[string]any{}
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" || len(n) > 64 {
			return nil, pkg.ErrInvalidArgument
		}
		updates["name"] = n
	}
	if req.Lang != nil {
		if *req.Lang == "" {
			return nil, pkg.ErrInvalidArgument
		}
		updates["lang"] = *req.Lang
	}
	if req.Status != nil {
		if *req.Status != 0 && *req.Status != 1 {
			return nil, pkg.ErrBadState
		}
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.Tag{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		var me *mysql.MySQLError
		if errors.As(res.Error, &me) && me.Number == 1062 {
			return nil, pkg.ErrConflict
		}
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	if req.Status != nil {
		data.WriteAudit(uc.db, ctx, adminID, "tag_status", "tag", strconv.FormatUint(id, 10), "status="+strconv.Itoa(int(*req.Status)))
	}
	// FORCE_MASTER: 写后读
	var t data.Tag
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	return &t, nil
}

// DeleteTag 硬删除并清理书籍-标签关联。
func (uc *AdminUsecase) DeleteTag(ctx context.Context, adminID int64, id uint64) error {
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("id = ?", id).Delete(&data.Tag{})
	if res.Error != nil {
		return pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("tag_id = ?", id).Delete(&data.BookTag{})
	data.WriteAudit(uc.db, ctx, adminID, "tag_delete", "tag", strconv.FormatUint(id, 10), "")
	return nil
}

// checkParent 校验父分类存在（0=一级，无需校验）。
func (uc *AdminUsecase) checkParent(ctx context.Context, parentID uint64) error {
	if parentID == 0 {
		return nil
	}
	var n int64
	if err := uc.db.WithContext(ctx).Model(&data.Category{}).Where("id = ?", parentID).Count(&n).Error; err != nil {
		return pkg.ErrAdminDB
	}
	if n == 0 {
		return pkg.ErrTargetNF
	}
	return nil
}

// CategoryUpdate / TagUpdate 可选更新字段（service 层从 proto optional 转换）。
type CategoryUpdate struct {
	Name      *string
	ParentID  *uint64
	SortOrder *int
	Status    *int8
}

type TagUpdate struct {
	Name   *string
	Lang   *string
	Status *int8
}
