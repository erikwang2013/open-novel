package service

// 书籍服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/book/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

type BookService struct {
	uc *biz.BookUsecase
	v1.UnimplementedBookServer
}

func NewBookService(uc *biz.BookUsecase) *BookService { return &BookService{uc: uc} }

func (s *BookService) GetBook(ctx context.Context, req *v1.GetBookReq) (*v1.BookReply, error) {
	it, err := s.uc.GetBook(ctx, u64(req.Id), pickLang(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	return toBookReply(it), nil
}

func (s *BookService) ListBooks(ctx context.Context, req *v1.ListBooksReq) (*v1.ListBooksReply, error) {
	p := pkg.ParsePage(req.Page, req.PageSize)
	items, total, err := s.uc.ListBooks(ctx, p, pickLang(ctx, req.Lang), u64(req.CategoryId), u64(req.TagId), req.Status)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.BookReply, 0, len(items))
	for _, it := range items {
		list = append(list, toBookReply(&it))
	}
	return &v1.ListBooksReply{List: list, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func (s *BookService) CreateBook(ctx context.Context, req *v1.CreateBookReq) (*v1.CreateBookReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := pkg.RequireRole(c, 2); err != nil {
		return nil, err
	}
	translations := make([]biz.TranslationParams, 0, len(req.Translations))
	for _, t := range req.Translations {
		translations = append(translations, biz.TranslationParams{Lang: t.Lang, Title: t.Title, Summary: t.Summary, Cover: t.Cover})
	}
	id, err := s.uc.CreateBook(ctx, biz.CreateBookParams{
		Title: req.Title, Author: req.Author, Summary: req.Summary, Cover: req.Cover,
		Lang: req.Lang, IsVip: int8(req.IsVip), CategoryIDs: u64s(req.CategoryIds), TagIDs: u64s(req.TagIds),
		Translations: translations,
	})
	if err != nil {
		return nil, err
	}
	// 以 DB 落盘的 author（req.Author）为准返回，不取 JWT claims，避免响应与库内不一致
	return &v1.CreateBookReply{Id: i64(id), Author: req.Author, Role: c.Role}, nil
}

func (s *BookService) UpsertBookTranslation(ctx context.Context, req *v1.UpsertBookTranslationReq) (*v1.UpsertBookTranslationReply, error) {
	c, err := auth(ctx)
	if err != nil {
		return nil, err
	}
	if err := pkg.RequireRole(c, 2); err != nil {
		return nil, err
	}
	if err := s.uc.UpsertTranslation(ctx, u64(req.Id), req.Lang, req.Title, req.Summary, req.Cover); err != nil {
		return nil, err
	}
	return &v1.UpsertBookTranslationReply{BookId: req.Id, Lang: req.Lang}, nil
}

func (s *BookService) ListCategories(ctx context.Context, req *v1.ListCategoriesReq) (*v1.ListCategoriesReply, error) {
	cats, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.CategoryReply, 0, len(cats))
	for _, c := range cats {
		list = append(list, &v1.CategoryReply{Id: i64(c.ID), Name: c.Name, ParentId: i64(c.ParentID)})
	}
	return &v1.ListCategoriesReply{List: list}, nil
}

func (s *BookService) ListTags(ctx context.Context, req *v1.ListTagsReq) (*v1.ListTagsReply, error) {
	tags, err := s.uc.ListTags(ctx, req.Lang)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.TagReply, 0, len(tags))
	for _, t := range tags {
		list = append(list, &v1.TagReply{Id: i64(t.ID), Name: t.Name, Lang: t.Lang})
	}
	return &v1.ListTagsReply{List: list}, nil
}

func toBookReply(it *biz.BookItem) *v1.BookReply {
	return &v1.BookReply{
		Id: i64(it.ID), Lang: it.Lang, Title: it.Title, Author: it.Author,
		Summary: it.Summary, Cover: it.Cover, IsVip: int32(it.IsVip), Status: int32(it.Status),
	}
}
