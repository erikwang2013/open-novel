package biz

// 评论用例：评论 / 点赞 / 举报 / 收藏（任务 #19）。
// 逻辑从 svc1 移植：评论 ≤2000 字符、点赞去重（uk_user_target）、举报置 status=2 待审核。

import (
	"context"

	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type CommentUsecase struct {
	db *gorm.DB
}

func NewCommentUsecase(d *data.Data) *CommentUsecase {
	return &CommentUsecase{db: d.DB}
}

type CommentItem struct {
	ID          uint64
	BookID      uint64
	ChapterID   uint64 // 0 = 书籍级
	UserID      uint64
	ParentID    uint64
	Content     string
	LikeCount   uint32
	ReportCount uint32
	Status      uint8
	CreatedAt   string
}

func toCommentItem(c data.Comment) CommentItem {
	it := CommentItem{ID: c.ID, BookID: c.BookID, UserID: c.UserID, ParentID: c.ParentID,
		Content: c.Content, LikeCount: c.LikeCount, ReportCount: c.ReportCount,
		Status: c.Status, CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
	if c.ChapterID != nil {
		it.ChapterID = *c.ChapterID
	}
	return it
}

// CreateComment 发表评论（chapterID 为 0 表示书籍级）。
func (uc *CommentUsecase) CreateComment(ctx context.Context, uid int64, bookID uint64, chapterID *uint64, content string) (*CommentItem, error) {
	if bookID == 0 || content == "" || len([]rune(content)) > 2000 {
		return nil, pkg.ErrCommentArg
	}
	c := data.Comment{BookID: bookID, ChapterID: chapterID, UserID: uint64(uid), Content: content, Status: 1}
	if err := uc.db.Clauses(gormdb.Write).Create(&c).Error; err != nil {
		return nil, pkg.ErrCommentDB
	}
	it := toCommentItem(c)
	return &it, nil
}

// ListComments 评论列表；chapterIDOpt：nil=全部，0=仅书籍级，>0=指定章节。
func (uc *CommentUsecase) ListComments(ctx context.Context, bookID uint64, chapterIDOpt *int64, p pkg.Page) ([]CommentItem, int64, error) {
	if bookID == 0 {
		return nil, 0, pkg.ErrCommentArg
	}
	q := uc.db.WithContext(ctx).Where("book_id = ? AND status = 1", bookID)
	if chapterIDOpt != nil {
		if *chapterIDOpt == 0 {
			q = q.Where("chapter_id IS NULL") // 仅书籍级
		} else {
			q = q.Where("chapter_id = ?", *chapterIDOpt)
		}
	}
	var total int64
	q.Model(&data.Comment{}).Count(&total)
	var list []data.Comment
	q.Order("id DESC").Limit(p.PageSize).Offset(p.Offset()).Find(&list)
	items := make([]CommentItem, 0, len(list))
	for _, c := range list {
		items = append(items, toCommentItem(c))
	}
	return items, total, nil
}

func (uc *CommentUsecase) LikeComment(ctx context.Context, uid int64, cid uint64) error {
	if cid == 0 {
		return pkg.ErrCommentArg
	}
	var c data.Comment
	if err := uc.db.WithContext(ctx).First(&c, cid).Error; err != nil || c.Status == 0 {
		return pkg.ErrCommentNF
	}
	err := uc.db.Clauses(gormdb.Write).Transaction(func(tx *gorm.DB) error {
		l := data.Like{UserID: uint64(uid), TargetType: data.TargetComment, TargetID: cid}
		if err := tx.Create(&l).Error; err != nil {
			return err // 重复 -> uk_user_target
		}
		return tx.Model(&data.Comment{}).Where("id = ?", cid).Update("like_count", gorm.Expr("like_count+1")).Error
	})
	if err != nil {
		return pkg.ErrCommentCflt
	}
	return nil
}

func (uc *CommentUsecase) UnlikeComment(ctx context.Context, uid int64, cid uint64) error {
	if cid == 0 {
		return pkg.ErrCommentArg
	}
	err := uc.db.Clauses(gormdb.Write).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, data.TargetComment, cid).Delete(&data.Like{})
		if res.RowsAffected == 0 {
			return pkg.ErrCommentCflt
		}
		return tx.Model(&data.Comment{}).Where("id = ?", cid).
			Update("like_count", gorm.Expr("GREATEST(like_count-1, 0)")).Error
	})
	if err != nil {
		return pkg.ErrCommentCflt
	}
	return nil
}

// ReportComment 举报：置 status=2 待审核。
func (uc *CommentUsecase) ReportComment(ctx context.Context, cid uint64) error {
	if cid == 0 {
		return pkg.ErrCommentArg
	}
	res := uc.db.Clauses(gormdb.Write).Model(&data.Comment{}).
		Where("id = ? AND status = 1", cid).Updates(map[string]any{
		"report_count": gorm.Expr("report_count+1"),
		"status":       2,
	})
	if res.RowsAffected == 0 {
		return pkg.ErrCommentNF
	}
	return nil
}

func (uc *CommentUsecase) AddFavorite(ctx context.Context, uid int64, bookID uint64) (*data.Favorite, error) {
	if bookID == 0 {
		return nil, pkg.ErrCommentArg
	}
	f := data.Favorite{UserID: uint64(uid), BookID: bookID}
	if err := uc.db.Clauses(gormdb.Write).Create(&f).Error; err != nil {
		return nil, pkg.ErrCommentCflt
	}
	return &f, nil
}

func (uc *CommentUsecase) DelFavorite(ctx context.Context, uid int64, bookID uint64) error {
	if bookID == 0 {
		return pkg.ErrCommentArg
	}
	uc.db.Clauses(gormdb.Write).Where("user_id = ? AND book_id = ?", uid, bookID).Delete(&data.Favorite{})
	return nil
}

func (uc *CommentUsecase) ListFavorites(ctx context.Context, uid int64, p pkg.Page) ([]data.Favorite, int64, error) {
	var total int64
	uc.db.WithContext(ctx).Model(&data.Favorite{}).Where("user_id = ?", uid).Count(&total)
	var list []data.Favorite
	uc.db.WithContext(ctx).Where("user_id = ?", uid).Order("id DESC").
		Limit(p.PageSize).Offset(p.Offset()).Find(&list)
	return list, total, nil
}
