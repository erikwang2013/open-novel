package middleware

// 攻击检测中间件：security-go 对 URL/Query/Headers/Cookies 做注入扫描。
// High/Critical 直接拒绝并计入 IP 黑名单（5 次/60s → 封禁 15min，httpval 默认值，与安全图一致）；
// Low/Medium 仅记日志放行（避免低档误杀）。只扫请求头不读 body，不污染后续 handler 的 body 读取。

import (
	"context"
	"net/http"

	"github.com/erikwang2013/security-go"
	"github.com/erikwang2013/security-go/all"
	"github.com/erikwang2013/security-go/httpval"
	"github.com/erikwang2013/security-go/storage"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"open-novel/backend/internal/pkg"
)

var (
	securityEngine = func() *security.Engine {
		e := security.NewEngine()
		all.RegisterAll(e)
		return e
	}()
	// IP 黑名单走 X-Forwarded-For（与 RateLimit 同源），内存后端。
	ipBlacklist = httpval.NewIPBlacklist(storage.NewMemory())
)

func Security() middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			ht, isHTTP := tr.(*khttp.Transport)
			if !isHTTP {
				return next(ctx, req) // gRPC 不走 HTTP 攻击面，跳过
			}
			if err := checkAttack(ht.Request()); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

// checkAttack 返回 nil 放行；高危/已封禁返回 403。高危命中按攻击源计数。
func checkAttack(r *http.Request) error {
	ip := r.Header.Get("X-Forwarded-For")
	if res := ipBlacklist.Detect(ip); res.Detected {
		log.Warnf("security: %s", res.Message)
		return pkg.ErrPermission
	}
	for _, res := range securityEngine.DetectRequest(r) {
		if !res.Detected {
			continue
		}
		if res.Severity >= security.SeverityHigh {
			// 空 IP 不计数，避免无 XFF 的请求共享同一 key 被集体封禁。
			if ip != "" {
				if banned, err := ipBlacklist.RecordAttack(ip); err == nil && banned {
					log.Warnf("security: IP %s banned after attacks", ip)
				}
			}
			log.Warnf("security: blocked %s (%s)", res.Name, res.Message)
			return pkg.ErrPermission
		}
		log.Infof("security: %s (%s)", res.Name, res.Message)
	}
	return nil
}
