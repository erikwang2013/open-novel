package service

// 管理端服务适配层：仪表盘统计 / 分类 / 标签（全部 requireAdmin，T-A-14/16）。

import (
	"context"
	"time"

	adminv1 "open-novel/backend/api/admin/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/data"
)

type AdminService struct {
	uc *biz.AdminUsecase
	adminv1.UnimplementedAdminServer
}

func NewAdminService(uc *biz.AdminUsecase) *AdminService { return &AdminService{uc: uc} }

func (s *AdminService) GetStats(ctx context.Context, req *adminv1.GetStatsReq) (*adminv1.GetStatsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	st, err := s.uc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	books := make([]*adminv1.HotBook, 0, len(st.HotBooks))
	for _, b := range st.HotBooks {
		books = append(books, &adminv1.HotBook{BookId: i64(b.BookID), Title: b.Title, Hot: b.Hot})
	}
	words := make([]*adminv1.HotKeyword, 0, len(st.HotKeywords))
	for _, w := range st.HotKeywords {
		words = append(words, &adminv1.HotKeyword{Keyword: w.Keyword, Count: w.Count})
	}
	return &adminv1.GetStatsReply{
		BookCount: st.BookCount, UserCount: st.UserCount, CommentCount: st.CommentCount, Dau: st.DAU,
		HotBooks: books, HotKeywords: words,
	}, nil
}

func (s *AdminService) ListCategories(ctx context.Context, req *adminv1.ListCategoriesReq) (*adminv1.ListCategoriesReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.CategoryReply, 0, len(items))
	for _, c := range items {
		list = append(list, toCategoryReply(&c))
	}
	return &adminv1.ListCategoriesReply{List: list, Total: int64(len(list))}, nil
}

func (s *AdminService) CreateCategory(ctx context.Context, req *adminv1.CreateCategoryReq) (*adminv1.CategoryReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	c, err := s.uc.CreateCategory(ctx, req.Name, u64(req.ParentId), int(req.SortOrder))
	if err != nil {
		return nil, err
	}
	return toCategoryReply(c), nil
}

func (s *AdminService) UpdateCategory(ctx context.Context, req *adminv1.UpdateCategoryReq) (*adminv1.CategoryReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var parentID *uint64
	if req.ParentId != nil {
		v := u64(*req.ParentId)
		parentID = &v
	}
	var sortOrder *int
	if req.SortOrder != nil {
		v := int(*req.SortOrder)
		sortOrder = &v
	}
	var status *int8
	if req.Status != nil {
		v := int8(*req.Status)
		status = &v
	}
	cat, err := s.uc.UpdateCategory(ctx, c.UID, u64(req.Id), biz.CategoryUpdate{
		Name: req.Name, ParentID: parentID, SortOrder: sortOrder, Status: status,
	})
	if err != nil {
		return nil, err
	}
	return toCategoryReply(cat), nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, req *adminv1.DeleteCategoryReq) (*adminv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCategory(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &adminv1.EmptyReply{}, nil
}

func (s *AdminService) ListTags(ctx context.Context, req *adminv1.ListTagsReq) (*adminv1.ListTagsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*adminv1.TagReply, 0, len(items))
	for _, t := range items {
		list = append(list, toTagReply(&t))
	}
	return &adminv1.ListTagsReply{List: list, Total: int64(len(list))}, nil
}

func (s *AdminService) CreateTag(ctx context.Context, req *adminv1.CreateTagReq) (*adminv1.TagReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	t, err := s.uc.CreateTag(ctx, req.Name, req.Lang)
	if err != nil {
		return nil, err
	}
	return toTagReply(t), nil
}

func (s *AdminService) UpdateTag(ctx context.Context, req *adminv1.UpdateTagReq) (*adminv1.TagReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var status *int8
	if req.Status != nil {
		v := int8(*req.Status)
		status = &v
	}
	t, err := s.uc.UpdateTag(ctx, c.UID, u64(req.Id), biz.TagUpdate{Name: req.Name, Lang: req.Lang, Status: status})
	if err != nil {
		return nil, err
	}
	return toTagReply(t), nil
}

func (s *AdminService) DeleteTag(ctx context.Context, req *adminv1.DeleteTagReq) (*adminv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteTag(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &adminv1.EmptyReply{}, nil
}

func toCategoryReply(c *data.Category) *adminv1.CategoryReply {
	return &adminv1.CategoryReply{
		Id: i64(c.ID), Name: c.Name, ParentId: i64(c.ParentID),
		SortOrder: int32(c.SortOrder), Status: int32(c.Status),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func toTagReply(t *data.Tag) *adminv1.TagReply {
	return &adminv1.TagReply{
		Id: i64(t.ID), Name: t.Name, Lang: t.Lang, Status: int32(t.Status),
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}
