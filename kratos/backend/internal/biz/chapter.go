package biz

// 章节用例：章节 / 正文 / 阅读进度 / 书架（任务 #16）。
// 逻辑从 svc1 移植：章节+正文同事务按 rune 计数、缓存 chapter:list:/chapter:content:。

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type ChapterUsecase struct {
	db    *gorm.DB
	cache *data.Cache
}

func NewChapterUsecase(d *data.Data) *ChapterUsecase {
	return &ChapterUsecase{db: d.DB, cache: d.Cache}
}

type ChapterItem struct {
	ID        uint64
	BookID    uint64
	ChapterNo uint32
	Title     string
	WordCount uint32
	IsVip     uint8
	Status    uint8
	CreatedAt string
}

// CreateChapter 新建章节（需登录）；章节 + 正文同事务，按 rune 计数。
func (uc *ChapterUsecase) CreateChapter(ctx context.Context, bookID uint64, chapterNo uint32, title, content, lang string) (*ChapterItem, error) {
	if bookID == 0 || chapterNo == 0 || title == "" {
		return nil, pkg.ErrInvalidArgument
	}
	if lang == "" {
		lang = "zh-CN"
	}
	ch := data.Chapter{BookID: bookID, ChapterNo: chapterNo, Title: title, Status: 1}
	// FORCE_MASTER: 建章后立即可读列表，避免主从延迟读不到
	err := uc.db.Clauses(gormdb.Write).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ch).Error; err != nil {
			return err
		}
		ct := data.ChapterContent{ChapterID: ch.ID, Lang: lang, Content: content}
		if err := tx.Create(&ct).Error; err != nil {
			return err
		}
		ch.WordCount = uint32(len([]rune(content)))
		return tx.Model(&ch).Update("word_count", ch.WordCount).Error
	})
	if err != nil {
		return nil, pkg.ErrChapterCflt
	}
	// CDN 失效：新建章节清掉同名旧 key（若存在）；未启用或未配 purge 端点时为空操作
	PurgeChapterAsync(ch.ID, lang)
	return &ChapterItem{ID: ch.ID, BookID: ch.BookID, ChapterNo: ch.ChapterNo, Title: ch.Title,
		WordCount: ch.WordCount, IsVip: ch.IsVip, Status: ch.Status,
		CreatedAt: ch.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}, nil
}

func (uc *ChapterUsecase) ListChapters(ctx context.Context, bookID uint64, p pkg.Page) ([]ChapterItem, int64, error) {
	if bookID == 0 {
		return nil, 0, pkg.ErrInvalidArgument
	}
	key := fmt.Sprintf("chapter:list:%d:%d:%d", bookID, p.Page, p.PageSize)
	payload, err := uc.cache.GetOrLoad(ctx, key, func() (string, error) {
		var total int64
		if err := uc.db.WithContext(ctx).Model(&data.Chapter{}).
			Where("book_id = ? AND status = 1", bookID).Count(&total).Error; err != nil {
			return "", err
		}
		var list []data.Chapter
		if err := uc.db.WithContext(ctx).Where("book_id = ? AND status = 1", bookID).
			Order("chapter_no").Limit(p.PageSize).Offset(p.Offset()).Find(&list).Error; err != nil {
			return "", err
		}
		items := make([]ChapterItem, 0, len(list))
		for _, c := range list {
			items = append(items, ChapterItem{ID: c.ID, BookID: c.BookID, ChapterNo: c.ChapterNo,
				Title: c.Title, WordCount: c.WordCount, IsVip: c.IsVip, Status: c.Status,
				CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00")})
		}
		b, err := json.Marshal(struct {
			List []ChapterItem `json:"list"`
			Total int64        `json:"total"`
			Page  int          `json:"page"`
			PageSize int       `json:"page_size"`
		}{items, total, p.Page, p.PageSize})
		return string(b), err
	})
	if err != nil {
		return nil, 0, pkg.ErrChapterDB
	}
	var out struct {
		List     []ChapterItem `json:"list"`
		Total    int64         `json:"total"`
		Page     int           `json:"page"`
		PageSize int           `json:"page_size"`
	}
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, 0, pkg.ErrChapterDB
	}
	return out.List, out.Total, nil
}

type ChapterContentItem struct {
	ID        uint64
	ChapterID uint64
	Lang      string
	Content   string
	IsVip     uint8 // CDN 缓存头判定（VIP 章节禁止静态化）
}

// GetChapterContent 正文；缓存 key 含 book_id+chapter_id+lang；无内容缓存空值 60s。
func (uc *ChapterUsecase) GetChapterContent(ctx context.Context, chapterID uint64, lang string) (*ChapterContentItem, error) {
	if chapterID == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	var ch data.Chapter
	if err := uc.db.WithContext(ctx).First(&ch, chapterID).Error; err != nil {
		return nil, pkg.ErrChapterNF
	}
	if ch.Status == 0 {
		return nil, pkg.ErrChapterDisabled // 禁用章节正文不可读
	}
	key := fmt.Sprintf("chapter:content:%d:%d:%s", ch.BookID, chapterID, lang)
	payload, err := uc.cache.GetOrLoad(ctx, key, func() (string, error) {
		var ct data.ChapterContent
		if err := uc.db.WithContext(ctx).
			Where("chapter_id = ? AND lang = ?", chapterID, lang).First(&ct).Error; err != nil {
			return "", nil // empty -> 60s 空值缓存
		}
		b, err := json.Marshal(ct)
		return string(b), err
	})
	if err != nil {
		return nil, pkg.ErrChapterDB
	}
	if payload == data.EmptyMarker {
		return nil, pkg.ErrChapterContent
	}
	var ct data.ChapterContent
	if err := json.Unmarshal([]byte(payload), &ct); err != nil {
		return nil, pkg.ErrChapterDB
	}
	return &ChapterContentItem{ID: ct.ID, ChapterID: ct.ChapterID, Lang: ct.Lang, Content: ct.Content, IsVip: ch.IsVip}, nil
}

func (uc *ChapterUsecase) GetProgress(ctx context.Context, uid int64, bookID uint64) (*data.ReadingProgress, error) {
	if bookID == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	var pr data.ReadingProgress
	// FORCE_MASTER: 更新进度后读进度走主库，避免主从延迟读到旧值
	err := uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("book_id = ? AND user_id = ?", bookID, uid).
		Order("updated_at DESC").First(&pr).Error
	if err != nil {
		return nil, pkg.ErrNoProgress
	}
	return &pr, nil
}

func (uc *ChapterUsecase) UpdateProgress(ctx context.Context, uid int64, bookID uint64, chapterID uint64, position uint32, lang string) (*data.ReadingProgress, error) {
	if bookID == 0 || chapterID == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	if lang == "" {
		lang = "zh-CN"
	}
	pr := data.ReadingProgress{UserID: uint64(uid), BookID: bookID, ChapterID: chapterID, Position: position}
	// FORCE_MASTER: 写完进度后立刻可读，避免主从延迟
	err := uc.db.Clauses(gormdb.Write).Transaction(func(tx *gorm.DB) error {
		var existing data.ReadingProgress
		if e := tx.Where("book_id = ? AND chapter_id = ? AND user_id = ?", bookID, chapterID, uid).
			First(&existing).Error; e == nil {
			return tx.Model(&existing).Update("position", position).Error
		}
		return tx.Create(&pr).Error
	})
	if err != nil {
		return nil, pkg.ErrChapterDB
	}
	// 阅读事件日志：尽力而为，Insert 失败仅记日志不阻塞进度保存。
	// ponytail: 不建队列/异步，量级与 search_log 同级可接受；若写入延迟成为瓶颈再引入批量。
	if e := uc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&data.ReadingLog{
		UserID: uint64(uid), BookID: bookID, ChapterID: chapterID, Lang: lang, Position: position,
	}).Error; e != nil {
		cdnLog.Log(log.LevelWarn, "msg", "write reading log failed", "err", e.Error())
	}
	return &pr, nil
}

func (uc *ChapterUsecase) AddToBookshelf(ctx context.Context, uid int64, bookID uint64) (*data.Bookshelf, error) {
	if bookID == 0 {
		return nil, pkg.ErrInvalidArgument
	}
	sh := data.Bookshelf{UserID: uint64(uid), BookID: bookID}
	if err := uc.db.Clauses(gormdb.Write).Create(&sh).Error; err != nil {
		return nil, pkg.ErrConflict
	}
	return &sh, nil
}

func (uc *ChapterUsecase) RemoveFromBookshelf(ctx context.Context, uid int64, bookID uint64) error {
	if bookID == 0 {
		return pkg.ErrInvalidArgument
	}
	uc.db.Clauses(gormdb.Write).Where("user_id = ? AND book_id = ?", uid, bookID).Delete(&data.Bookshelf{})
	return nil
}

// SetChapterStatus 启用/禁用章节（仅管理员；0 禁用 1 启用）；失效该书籍章节列表缓存。
func (uc *ChapterUsecase) SetChapterStatus(ctx context.Context, adminID int64, id uint64, status uint8) error {
	if id == 0 || (status != 0 && status != 1) {
		return pkg.ErrInvalidArgument
	}
	var ch data.Chapter
	if err := uc.db.WithContext(ctx).First(&ch, id).Error; err != nil {
		return pkg.ErrChapterNF
	}
	res := uc.db.Clauses(gormdb.Write).Model(&data.Chapter{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return pkg.ErrChapterDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrChapterNF
	}
	uc.cache.DelPattern(ctx, fmt.Sprintf("chapter:list:%d:*", ch.BookID))
	// CDN 失效：状态变更影响所有语言版本，收集 langs 单次多 key 广播（§4.1 合批）
	var langs []string
	uc.db.WithContext(ctx).Model(&data.ChapterContent{}).Where("chapter_id = ?", id).Distinct().Pluck("lang", &langs)
	PurgeChaptersAsync(id, langs)
	data.WriteAudit(uc.db, ctx, adminID, "chapter_status", "chapter", strconv.FormatUint(id, 10), fmt.Sprintf("status=%d", status))
	return nil
}

func (uc *ChapterUsecase) ListBookshelf(ctx context.Context, uid int64, p pkg.Page) ([]data.Bookshelf, int64, error) {
	var total int64
	uc.db.WithContext(ctx).Model(&data.Bookshelf{}).Where("user_id = ?", uid).Count(&total)
	var list []data.Bookshelf
	uc.db.WithContext(ctx).Where("user_id = ?", uid).Order("sort_order, id DESC").
		Limit(p.PageSize).Offset(p.Offset()).Find(&list)
	return list, total, nil
}

// LangFromAccept 取 Accept-Language 首个语言标签，缺省 zh-CN。
func LangFromAccept(h string) string {
	for _, part := range strings.Split(h, ",") {
		lang := strings.TrimSpace(strings.Split(part, ";")[0])
		if lang != "" {
			return lang
		}
	}
	return "zh-CN"
}
