package service

// requireAdmin 守卫测试：匿名 140401、普通用户 180401（守卫在 uc 调用前，无需 DB）。

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	behaviorv1 "open-novel/backend/api/behavior/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

func TestGetBehaviorStatsGuard(t *testing.T) {
	s := NewBehaviorService((*biz.BehaviorUsecase)(nil))
	ctx := context.Background()
	if _, err := s.GetBehaviorStats(ctx, &behaviorv1.GetBehaviorStatsReq{}); err == nil || kerrors.FromError(err).Code != 140401 {
		t.Fatalf("anonymous: want 140401, got %v", err)
	}
	readerCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 1})
	if _, err := s.GetBehaviorStats(readerCtx, &behaviorv1.GetBehaviorStatsReq{}); err == nil || kerrors.FromError(err).Code != 180401 {
		t.Fatalf("reader: want 180401, got %v", err)
	}
}
