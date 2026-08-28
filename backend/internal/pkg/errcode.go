package pkg

import kerrors "github.com/go-kratos/kratos/v2/errors"

// 业务错误码段分配（规划文档 §四）：HTTP 状态与业务码分离
// 格式：1 + 服务号 + 3 位标准码
//   11xxxx User 服务    12xxxx Book 服务    14xxxx Chapter 服务
//   15xxxx Comment 服务 16xxxx Search 服务  17xxxx Recommendation 服务
// 标准码：400 参数 / 401 未认证 / 403 无权限 / 404 未找到 / 409 冲突 / 429 限流 / 500 内部
// 自定义 ErrorEncoder（server/server.go）将业务码（≥100000）映射为 HTTP 200 + {code,reason,message}。

// Biz 构造 kratos 业务错误。
func Biz(code int, reason, msg string) error { return kerrors.New(code, reason, msg) }

var (
	// 通用（14xxxx，与 Chapter 共用）
	ErrInvalidArgument = Biz(140400, "INVALID_ARGUMENT", "invalid argument")
	ErrUnauth          = Biz(140401, "UNAUTHENTICATED", "unauthenticated")
	ErrPermission      = Biz(140403, "PERMISSION_DENIED", "permission denied")
	ErrNotFound        = Biz(140404, "NOT_FOUND", "not found")
	ErrConflict        = Biz(140409, "CONFLICT", "conflict")
	ErrTooMany         = Biz(140429, "TOO_MANY_REQUESTS", "too many requests")

	// User（11xxxx）
	ErrUserArg     = Biz(110400, "INVALID_ARGUMENT", "invalid argument")
	ErrCred        = Biz(110401, "BAD_CREDENTIALS", "incorrect username or password")
	ErrToken       = Biz(110401, "TOKEN_INVALID", "refresh token invalid or expired")
	ErrClosed      = Biz(110403, "ACCOUNT_DISABLED", "account disabled")
	ErrUserExists  = Biz(110409, "ACCOUNT_EXISTS", "username or email already exists")
	ErrUserInternal = Biz(110500, "INTERNAL", "internal error")

	// Book（12xxxx）
	ErrBookArg   = Biz(120400, "INVALID_ARGUMENT", "invalid argument")
	ErrBookNF    = Biz(120404, "BOOK_NOT_FOUND", "book not found")
	ErrBookInternal = Biz(120500, "INTERNAL", "internal error")

	// Chapter（14xxxx）
	ErrChapterCflt   = Biz(140409, "CHAPTER_CONFLICT", "chapter conflict")
	ErrChapterDB     = Biz(140500, "DB_ERROR", "db error")
	ErrChapterNF     = Biz(140404, "CHAPTER_NOT_FOUND", "chapter not found")
	ErrChapterContent = Biz(140404, "CONTENT_NOT_FOUND", "chapter content not found")
	ErrNoProgress    = Biz(140404, "NO_PROGRESS", "no progress yet")

	// Comment（15xxxx）
	ErrCommentArg   = Biz(150400, "INVALID_ARGUMENT", "invalid argument")
	ErrCommentNF    = Biz(150404, "COMMENT_NOT_FOUND", "comment not found")
	ErrCommentCflt  = Biz(150409, "ALREADY_LIKED_OR_FAVORITED", "already liked or favorited")
	ErrCommentDB    = Biz(150500, "DB_ERROR", "db error")

	// Search（16xxxx）
	ErrSearchArg = Biz(160400, "INVALID_ARGUMENT", "invalid argument")
	ErrSearch    = Biz(160500, "SEARCH_FAILED", "search failed")

	// Recommendation（17xxxx）
	ErrRecommendArg = Biz(170400, "INVALID_ARGUMENT", "strategy must be hot or new")
	ErrRecommend    = Biz(170500, "RECOMMEND_FAILED", "recommend failed")
)
