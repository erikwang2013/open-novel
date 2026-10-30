package middleware

// API 版本协商：版本不写在 URL，通过请求头 X-Api-Version 传递（当前仅 v1）。
// gRPC 无版本概念，直接放行。

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"open-novel/backend/internal/pkg"
)

const (
	apiVersionHeader = "X-Api-Version"
	supportedVersion = "v1"
)

// isWebhookPath 渠道回调路径：第三方不会携带 X-Api-Version，跳过版本校验。
func isWebhookPath(path string) bool {
	return strings.HasPrefix(path, "/api/payments/webhook/")
}

// ApiVersion 校验 HTTP 请求头 X-Api-Version；缺失或值不支持时拒绝。
func ApiVersion() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			ht, isHTTP := tr.(*khttp.Transport)
			if !isHTTP {
				return next(ctx, req)
			}
			if isWebhookPath(ht.PathTemplate()) || isWebhookPath(ht.Request().URL.Path) {
				return next(ctx, req)
			}
			if v := tr.RequestHeader().Get(apiVersionHeader); v != supportedVersion {
				return nil, pkg.ErrApiVersion
			}
			return next(ctx, req)
		}
	}
}
