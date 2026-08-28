package server

// HTTP/gRPC 服务组装：middleware 链（§六 recovery/ratelimit/jwt/validate/tracing）
// + 业务码错误编码（≥100000 → HTTP 200 + {code,reason,message}，§四）。

import (
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"net/http"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	commentv1 "open-novel/backend/api/comment/v1"
	chapterv1 "open-novel/backend/api/chapter/v1"
	bookv1 "open-novel/backend/api/book/v1"
	recommendationv1 "open-novel/backend/api/recommendation/v1"
	searchv1 "open-novel/backend/api/search/v1"
	userv1 "open-novel/backend/api/user/v1"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/middleware"
	"open-novel/backend/internal/pkg"
	"open-novel/backend/internal/service"
)

// 限流路由表（按路径模板，§六）：登录/评论发布/举报/搜索。
var rateLimits = map[string]int{
	"/api/users/login":         10,
	"/api/comments":            10,
	"/api/comments/{id}/report": 5,
	"/api/search":              10,
}

func NewHTTPServer(c *conf.Server, am *pkg.AuthManager, logger log.Logger,
	user *service.UserService, book *service.BookService, chapter *service.ChapterService,
	comment *service.CommentService, search *service.SearchService, rec *service.RecommendationService,
) *khttp.Server {
	opts := []khttp.ServerOption{
		khttp.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			middleware.ApiVersion(),
			middleware.RateLimit(rateLimits),
			middleware.OptionalAuth(am),
		),
		khttp.ErrorEncoder(errEncoder),
	}
	if c.HttpAddr != "" {
		opts = append(opts, khttp.Address(c.HttpAddr))
	}
	srv := khttp.NewServer(opts...)
	userv1.RegisterUserHTTPServer(srv, user)
	bookv1.RegisterBookHTTPServer(srv, book)
	chapterv1.RegisterChapterHTTPServer(srv, chapter)
	commentv1.RegisterCommentHTTPServer(srv, comment)
	searchv1.RegisterSearchHTTPServer(srv, search)
	recommendationv1.RegisterRecommendationHTTPServer(srv, rec)
	return srv
}

func NewGRPCServer(c *conf.Server, am *pkg.AuthManager, logger log.Logger,
	user *service.UserService, book *service.BookService, chapter *service.ChapterService,
	comment *service.CommentService, search *service.SearchService, rec *service.RecommendationService,
) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			middleware.OptionalAuth(am),
		),
	}
	if c.GrpcAddr != "" {
		opts = append(opts, grpc.Address(c.GrpcAddr))
	}
	srv := grpc.NewServer(opts...)
	userv1.RegisterUserServer(srv, user)
	bookv1.RegisterBookServer(srv, book)
	chapterv1.RegisterChapterServer(srv, chapter)
	commentv1.RegisterCommentServer(srv, comment)
	searchv1.RegisterSearchServer(srv, search)
	recommendationv1.RegisterRecommendationServer(srv, rec)
	return srv
}

// errEncoder 业务码（≥100000）→ HTTP 200 + 业务码；标准码（100-599）→ 原样状态。
func errEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := kerrors.FromError(err)
	status := http.StatusInternalServerError
	switch {
	case se.Code >= 100000:
		status = http.StatusOK // §四：HTTP 状态与业务码分离
	case se.Code >= 100 && se.Code < 600:
		status = int(se.Code)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code": se.Code, "reason": se.Reason, "message": se.Message, "detail": nil,
	})
}
