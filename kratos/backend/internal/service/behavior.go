package service

// 阅读行为分析服务适配层（requireAdmin，GET /api/stats/behavior）。

import (
	"context"

	behaviorv1 "open-novel/backend/api/behavior/v1"
	"open-novel/backend/internal/biz"
)

type BehaviorService struct {
	uc *biz.BehaviorUsecase
	behaviorv1.UnimplementedBehaviorServer
}

func NewBehaviorService(uc *biz.BehaviorUsecase) *BehaviorService {
	return &BehaviorService{uc: uc}
}

func (s *BehaviorService) GetBehaviorStats(ctx context.Context, req *behaviorv1.GetBehaviorStatsReq) (*behaviorv1.GetBehaviorStatsReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	st, err := s.uc.Stats(ctx, langOrDefault(ctx, req.Lang))
	if err != nil {
		return nil, err
	}
	books := make([]*behaviorv1.HotReadingBook, 0, len(st.HotBooks))
	for _, b := range st.HotBooks {
		books = append(books, &behaviorv1.HotReadingBook{BookId: i64(b.BookID), Title: b.Title, Count: b.Count})
	}
	hourly := make([]int64, 24)
	for i, v := range st.Hourly {
		hourly[i] = v
	}
	return &behaviorv1.GetBehaviorStatsReply{
		ActiveReaders: st.ActiveReaders, Readers_7D: st.Readers7d,
		HotBooks: books, HourlyDistribution: hourly,
	}, nil
}
