package service

// service 层共享工具：lang 解析、客户端 IP、鉴权声明提取。

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/transport"

	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

// pickLang 优先 ?lang=，其次 Accept-Language 首个语言标签；都缺省返回空。
func pickLang(ctx context.Context, lang string) string {
	if lang != "" {
		return lang
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		if h := tr.RequestHeader().Get("Accept-Language"); h != "" {
			return strings.TrimSpace(strings.Split(h, ",")[0])
		}
	}
	return ""
}

// userAgent 取请求 UA（审计日志，best-effort）。
func userAgent(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("User-Agent")
	}
	return ""
}

// clientIP 从 X-Forwarded-For 取首个 IP（日志 best-effort，缺省空串）。
func clientIP(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if xff := tr.RequestHeader().Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.Split(xff, ",")[0])
		}
	}
	return ""
}

// claims 取 ctx 中的声明（middleware 已解析，未认证为零值）。
func claims(ctx context.Context) pkg.Claims { return pkg.ClaimsFrom(ctx) }

// auth 要求已认证，否则 140401。
func auth(ctx context.Context) (pkg.Claims, error) {
	c := pkg.ClaimsFrom(ctx)
	if c.UID == 0 {
		return c, pkg.ErrUnauth
	}
	return c, nil
}

// requireAdmin 要求管理员角色（kratos 无按路由挂中间件，管理操作在 service 层校验），否则 180401。
func requireAdmin(ctx context.Context) (pkg.Claims, error) {
	c, err := auth(ctx)
	if err != nil {
		return c, err
	}
	if err := pkg.RequireAdmin(c); err != nil {
		return c, err
	}
	return c, nil
}

// proto int64 ↔ 模型 uint64 转换（列均为无符号）。
func u64(v int64) uint64     { return uint64(v) }
func i64(v uint64) int64     { return int64(v) }

func u64s(v []int64) []uint64 {
	out := make([]uint64, len(v))
	for i, x := range v {
		out[i] = uint64(x)
	}
	return out
}

// langOrDefault 缺省 zh-CN。
func langOrDefault(ctx context.Context, lang string) string {
	if l := pickLang(ctx, lang); l != "" {
		return l
	}
	return biz.LangFromAccept("")
}
