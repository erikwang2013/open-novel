package service

// 章节服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/chapter/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type ChapterService struct {
	uc *biz.ChapterUsecase
	v1.UnimplementedChapterServer
}

func NewChapterService(uc *biz.ChapterUsecase) *ChapterService { return &ChapterService{uc: uc} }

func (s *ChapterService) CreateChapter(ctx context.Context, req *v1.CreateChapterReq) (*v1.ChapterReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := pkg.RequireRole(c, 2); err != nil { // 仅作者（2）及以上可建章节，对齐建书
		return nil, err
	}
	it, err := s.uc.CreateChapter(ctx, u64(req.BookId), req.ChapterNo, req.Title, req.Content, req.Lang)
	if err != nil {
		return nil, err
	}
	return toChapterReply(it), nil
}

func (s *ChapterService) ListChapters(ctx context.Context, req *v1.ListChaptersReq) (*v1.ListChaptersReply, error) {
	p := pkg.ParsePage(req.Page, req.PageSize)
	items, total, err := s.uc.ListChapters(ctx, u64(req.BookId), p)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ChapterReply, 0, len(items))
	for _, it := range items {
		list = append(list, toChapterReply(&it))
	}
	return &v1.ListChaptersReply{List: list, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func (s *ChapterService) GetChapterContent(ctx context.Context, req *v1.GetChapterContentReq) (*v1.ChapterContentReply, error) {
	it, err := s.uc.GetChapterContent(ctx, u64(req.Id), langOrDefault(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	return &v1.ChapterContentReply{Id: i64(it.ID), ChapterId: i64(it.ChapterID), Lang: it.Lang, Content: it.Content}, nil
}

func (s *ChapterService) GetReadingProgress(ctx context.Context, req *v1.GetReadingProgressReq) (*v1.ProgressReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	pr, err := s.uc.GetProgress(ctx, c.UID, u64(req.BookId))
	if err != nil {
		return nil, err
	}
	return toProgressReply(pr), nil
}

func (s *ChapterService) UpdateReadingProgress(ctx context.Context, req *v1.UpdateReadingProgressReq) (*v1.ProgressReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	pr, err := s.uc.UpdateProgress(ctx, c.UID, u64(req.BookId), u64(req.ChapterId), req.Position)
	if err != nil {
		return nil, err
	}
	return toProgressReply(pr), nil
}

func (s *ChapterService) AddToBookshelf(ctx context.Context, req *v1.AddToBookshelfReq) (*v1.ShelfReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	sh, err := s.uc.AddToBookshelf(ctx, c.UID, u64(req.BookId))
	if err != nil {
		return nil, err
	}
	return &v1.ShelfReply{Id: i64(sh.ID), UserId: int64(sh.UserID), BookId: i64(sh.BookID), SortOrder: int32(sh.SortOrder)}, nil
}

func (s *ChapterService) RemoveFromBookshelf(ctx context.Context, req *v1.RemoveFromBookshelfReq) (*v1.RemoveFromBookshelfReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.RemoveFromBookshelf(ctx, c.UID, u64(req.BookId)); err != nil {
		return nil, err
	}
	return &v1.RemoveFromBookshelfReply{}, nil
}

func (s *ChapterService) UpdateChapterStatus(ctx context.Context, req *v1.UpdateChapterStatusReq) (*v1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.SetChapterStatus(ctx, c.UID, u64(req.Id), uint8(req.Status)); err != nil {
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *ChapterService) ListBookshelf(ctx context.Context, req *v1.ListBookshelfReq) (*v1.ListBookshelfReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	p := pkg.ParsePage(req.Page, req.PageSize)
	list, total, err := s.uc.ListBookshelf(ctx, c.UID, p)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ShelfReply, 0, len(list))
	for _, sh := range list {
		items = append(items, &v1.ShelfReply{Id: i64(sh.ID), UserId: int64(sh.UserID), BookId: i64(sh.BookID), SortOrder: int32(sh.SortOrder)})
	}
	return &v1.ListBookshelfReply{List: items, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func toChapterReply(it *biz.ChapterItem) *v1.ChapterReply {
	return &v1.ChapterReply{
		Id: i64(it.ID), BookId: i64(it.BookID), ChapterNo: it.ChapterNo, Title: it.Title,
		WordCount: it.WordCount, IsVip: int32(it.IsVip), Status: int32(it.Status), CreatedAt: it.CreatedAt,
	}
}

func toProgressReply(pr *data.ReadingProgress) *v1.ProgressReply {
	return &v1.ProgressReply{
		Id: i64(pr.ID), UserId: int64(pr.UserID), BookId: i64(pr.BookID), ChapterId: i64(pr.ChapterID),
		Position: pr.Position, UpdatedAt: pr.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
