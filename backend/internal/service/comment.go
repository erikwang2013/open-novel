package service

// 评论服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/comment/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

type CommentService struct {
	uc *biz.CommentUsecase
	v1.UnimplementedCommentServer
}

func NewCommentService(uc *biz.CommentUsecase) *CommentService { return &CommentService{uc: uc} }

func (s *CommentService) CreateComment(ctx context.Context, req *v1.CreateCommentReq) (*v1.CommentReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.CreateComment(ctx, c.UID, u64(req.BookId), toUint64Ptr(req.ChapterId), req.Content)
	if err != nil {
		return nil, err
	}
	return toCommentReply(it), nil
}

func (s *CommentService) ListComments(ctx context.Context, req *v1.ListCommentsReq) (*v1.ListCommentsReply, error) {
	p := pkg.ParsePage(req.Page, req.PageSize)
	items, total, err := s.uc.ListComments(ctx, u64(req.BookId), req.ChapterId, p)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.CommentReply, 0, len(items))
	for _, it := range items {
		list = append(list, toCommentReply(&it))
	}
	return &v1.ListCommentsReply{List: list, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func (s *CommentService) LikeComment(ctx context.Context, req *v1.LikeCommentReq) (*v1.EmptyReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.LikeComment(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *CommentService) UnlikeComment(ctx context.Context, req *v1.UnlikeCommentReq) (*v1.EmptyReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.UnlikeComment(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *CommentService) ReportComment(ctx context.Context, req *v1.ReportCommentReq) (*v1.EmptyReply, error) {
	if _, err := auth(ctx); err != nil {
		return nil, err
	}
	if err := s.uc.ReportComment(ctx, u64(req.Id)); err != nil {
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *CommentService) AddFavorite(ctx context.Context, req *v1.AddFavoriteReq) (*v1.FavoriteReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	f, err := s.uc.AddFavorite(ctx, c.UID, u64(req.BookId))
	if err != nil {
		return nil, err
	}
	return &v1.FavoriteReply{Id: i64(f.ID), UserId: int64(f.UserID), BookId: i64(f.BookID)}, nil
}

func (s *CommentService) DelFavorite(ctx context.Context, req *v1.DelFavoriteReq) (*v1.EmptyReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DelFavorite(ctx, c.UID, u64(req.BookId)); err != nil {
		return nil, err
	}
	return &v1.EmptyReply{}, nil
}

func (s *CommentService) ListFavorites(ctx context.Context, req *v1.ListFavoritesReq) (*v1.ListFavoritesReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	p := pkg.ParsePage(req.Page, req.PageSize)
	list, total, err := s.uc.ListFavorites(ctx, c.UID, p)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.FavoriteReply, 0, len(list))
	for _, f := range list {
		items = append(items, &v1.FavoriteReply{Id: i64(f.ID), UserId: int64(f.UserID), BookId: i64(f.BookID)})
	}
	return &v1.ListFavoritesReply{List: items, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func toUint64Ptr(v *int64) *uint64 {
	if v == nil {
		return nil
	}
	u := uint64(*v)
	return &u
}

func toCommentReply(it *biz.CommentItem) *v1.CommentReply {
	return &v1.CommentReply{
		Id: i64(it.ID), BookId: i64(it.BookID), ChapterId: i64(it.ChapterID), UserId: int64(it.UserID),
		ParentId: i64(it.ParentID), Content: it.Content, LikeCount: it.LikeCount,
		ReportCount: it.ReportCount, Status: int32(it.Status), CreatedAt: it.CreatedAt,
	}
}
