package middleware

// 可选鉴权中间件：解析 Bearer token 放入 ctx（未携带/无效 → 零值 Claims）。
// 写操作在 service 层用 auth() 强制校验（kratos 无按路由挂中间件，统一解析更简单）。

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"open-novel/backend/internal/pkg"
)

func OptionalAuth(am *pkg.AuthManager) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if am != nil {
				if tr, ok := transport.FromServerContext(ctx); ok {
					if h := tr.RequestHeader().Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
						if c, err := am.Parse(h[7:]); err == nil {
							ctx = pkg.WithClaims(ctx, c)
						}
					}
				}
			}
			return next(ctx, req)
		}
	}
}
