package middleware

// 按路径限流中间件（§六）：固定窗口按 IP 计数。
// 路由表来自 svc1：login 10/分钟，comment 发布 10/分钟、举报 5/分钟，search 10/分钟。
// 注意：路径不含版本（/api/...），版本经 X-Api-Version 头协商。

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"open-novel/backend/internal/pkg"
)

type rateLimiter struct {
	mu    sync.Mutex
	seen  map[string][]time.Time
	limit int
}

func (l *rateLimiter) allow(ip string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune(ip, now)
	hits := l.seen[ip]
	if len(hits) >= l.limit {
		return false
	}
	l.seen[ip] = append(hits, now)
	if len(l.seen) > 1024 {
		l.sweep(now)
	}
	return true
}

// prune 剔除该 IP 超窗计数，清空后删键，防止 map 随 IP 数无限增长。
func (l *rateLimiter) prune(ip string, now time.Time) {
	ts := l.seen[ip]
	hits := ts[:0]
	for _, t := range ts {
		if now.Sub(t) < time.Minute {
			hits = append(hits, t)
		}
	}
	if len(hits) == 0 {
		delete(l.seen, ip)
		return
	}
	l.seen[ip] = hits
}

// sweep 容量保护：map 超阈值时全量清理过期键（防伪造 X-Forwarded-For 撑爆内存）。
func (l *rateLimiter) sweep(now time.Time) {
	for ip := range l.seen {
		l.prune(ip, now)
	}
}

// RateLimit 按路径限流；limits 形如 {"/api/users/login": 10}，未列出的路径放行。
func RateLimit(limits map[string]int) middleware.Middleware {
	lms := make(map[string]*rateLimiter, len(limits))
	for path, n := range limits {
		lms[path] = &rateLimiter{seen: map[string][]time.Time{}, limit: n}
	}
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			// HTTP 用路由模板（"/api/users/login"）匹配；gRPC 用 operation，未列出的放行。
			key := tr.Operation()
			if ht, isHTTP := tr.(*khttp.Transport); isHTTP {
				key = ht.PathTemplate()
			}
			if lm, hit := lms[key]; hit {
				if !lm.allow(tr.RequestHeader().Get("X-Forwarded-For")) {
					return nil, pkg.ErrTooMany
				}
			}
			return next(ctx, req)
		}
	}
}
