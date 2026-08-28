package service

// 推荐服务适配层：proto 消息 ↔ biz 用例。

import (
	"context"

	"open-novel/backend/api/recommendation/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

type RecommendationService struct {
	uc *biz.RecommendUsecase
	v1.UnimplementedRecommendationServer
}

func NewRecommendationService(uc *biz.RecommendUsecase) *RecommendationService {
	return &RecommendationService{uc: uc}
}

func (s *RecommendationService) GetRecommendations(ctx context.Context, req *v1.GetRecommendationsReq) (*v1.GetRecommendationsReply, error) {
	p := pkg.ParsePage(req.Page, req.PageSize)
	items, total, err := s.uc.List(ctx, req.Strategy, p)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.RecommendItemReply, 0, len(items))
	for _, it := range items {
		list = append(list, &v1.RecommendItemReply{
			BookId: i64(it.BookID), Title: it.Title, Author: it.Author,
			Summary: it.Summary, Cover: it.Cover, Lang: it.Lang, Strategy: it.Strategy,
		})
	}
	// 仅登录用户记录曝光
	s.uc.Log(ctx, claims(ctx).UID, req.Strategy, items)
	return &v1.GetRecommendationsReply{List: list, Total: total, Page: int32(p.Page), PageSize: int32(p.PageSize)}, nil
}
