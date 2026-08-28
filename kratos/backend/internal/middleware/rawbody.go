package middleware

// Webhook 原始 body 预读：Stripe/NOWPayments 对原始字节做签名，消息绑定会丢失
// 原始内容，故仅对 webhook 路径在绑定前把 body 存入 ctx 并恢复 Body。

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const maxWebhookBody = 1 << 20 // 1MB 上限

type rawBodyKey struct{}

// WithRawBody 存入原始请求体。
func WithRawBody(ctx context.Context, b []byte) context.Context {
	return context.WithValue(ctx, rawBodyKey{}, b)
}

// RawBodyFrom 取出中间件预读的原始请求体。
func RawBodyFrom(ctx context.Context) []byte {
	if b, ok := ctx.Value(rawBodyKey{}).([]byte); ok {
		return b
	}
	return nil
}

// RawBody 预读 webhook 原始 body（限 1MB）存入 ctx，并恢复 Body 供绑定。
func RawBody() middleware.Middleware {
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
			r := ht.Request()
			if r.Body == nil || !strings.HasPrefix(r.URL.Path, "/api/payments/webhook/") {
				return next(ctx, req)
			}
			raw, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
			if err != nil {
				return nil, err
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(raw)) // 恢复给消息绑定
			return next(WithRawBody(ctx, raw), req)
		}
	}
}
