package service

// 搜索服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/search/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

type SearchService struct {
	uc *biz.SearchUsecase
	v1.UnimplementedSearchServer
}

func NewSearchService(uc *biz.SearchUsecase) *SearchService { return &SearchService{uc: uc} }

func (s *SearchService) SearchBooks(ctx context.Context, req *v1.SearchBooksReq) (*v1.SearchBooksReply, error) {
	p := pkg.ParsePage(req.Page, req.PageSize)
	docs, total, err := s.uc.Search(ctx, req.Q, req.Lang, p, claims(ctx).UID, clientIP(ctx))
	if err != nil {
		return nil, err
	}
	return &v1.SearchBooksReply{List: toDocList(docs), Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}

func (s *SearchService) HotSearches(ctx context.Context, req *v1.HotSearchesReq) (*v1.HotSearchesReply, error) {
	docs, total, err := s.uc.Hot(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.HotSearchesReply{List: toDocList(docs), Total: total, Page: 1, PageSize: int32(len(docs))}, nil
}

func (s *SearchService) HotKeywords(ctx context.Context, req *v1.HotKeywordsReq) (*v1.HotKeywordsReply, error) {
	words, err := s.uc.HotKeywords(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.HotKeyword, 0, len(words))
	for _, w := range words {
		list = append(list, &v1.HotKeyword{Keyword: w.Keyword, Count: w.Count})
	}
	return &v1.HotKeywordsReply{List: list}, nil
}

// Suggest 搜索建议（无鉴权，同 HotKeywords；读接口，匿名可访问）。
func (s *SearchService) Suggest(ctx context.Context, req *v1.SuggestReq) (*v1.SuggestReply, error) {
	words, err := s.uc.Suggest(ctx, req.Q)
	if err != nil {
		return nil, err
	}
	return &v1.SuggestReply{Keywords: words}, nil
}

func toDocList(docs []data.SearchDoc) []*v1.BookDoc {
	list := make([]*v1.BookDoc, 0, len(docs))
	for _, d := range docs {
		list = append(list, &v1.BookDoc{
			BookId: i64(d.BookID), Lang: d.Lang, Status: int32(d.Status), Hot: d.Hot, CreatedAt: d.CreatedAt,
			TitleZh: d.TitleZh, TitleEn: d.TitleEn, TitleJa: d.TitleJa, TitleKo: d.TitleKo,
			SummaryZh: d.SummaryZh, SummaryEn: d.SummaryEn, SummaryJa: d.SummaryJa, SummaryKo: d.SummaryKo,
			AuthorZh: d.AuthorZh, AuthorEn: d.AuthorEn, AuthorJa: d.AuthorJa, AuthorKo: d.AuthorKo,
		})
	}
	return list
}
