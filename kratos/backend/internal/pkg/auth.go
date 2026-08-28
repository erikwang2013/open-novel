package pkg

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// 鉴权核心（从 svc1 移植）：access JWT（HS256）+ refresh 随机串 Redis 轮换（7d）。
// JWT_SECRET 环境变量为唯一签名密钥来源，未设置用开发默认值，生产必须注入。

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret-change-me"
}

// Claims 是 access_token 中的业务声明。
type Claims struct {
	UID      int64
	Username string
	Role     int32
}

type ctxKey int

const claimsKey ctxKey = 1

// WithClaims 将声明写入 context（由 middleware/auth.go 解析后调用）。
func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// ClaimsFrom 取出声明；未认证返回零值（UID=0）。
func ClaimsFrom(ctx context.Context) Claims {
	if c, ok := ctx.Value(claimsKey).(Claims); ok {
		return c
	}
	return Claims{}
}

// 角色常量（novel_user.role）。
const (
	RoleReader = 1
	RoleAuthor = 2
	RoleAdmin  = 3
)

// RequireRole 校验角色下限（1普通/2作者/3管理员/4运营），不足返回 140403。
func RequireRole(c Claims, min int32) error {
	if c.Role < min {
		return ErrPermission
	}
	return nil
}

// RequireAdmin 校验管理员（role==3），否则 180401。
func RequireAdmin(c Claims) error {
	if c.Role != RoleAdmin {
		return ErrAdmin
	}
	return nil
}

// AuthManager 签发 access/refresh 双令牌；refresh 存 Redis（轮换制，GETDEL 防重放）。
type AuthManager struct {
	rdb        *redis.Client
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthManager(rdb *redis.Client, accessTTL, refreshTTL time.Duration) *AuthManager {
	if accessTTL <= 0 {
		accessTTL = 30 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &AuthManager{rdb: rdb, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *AuthManager) AccessTTL() time.Duration { return m.accessTTL }

// Issue 签发 access（JWT，含 uid/username/role）+ refresh（32B 随机串）。
func (m *AuthManager) Issue(ctx context.Context, uid int64, username string, role int32) (access, refresh string, err error) {
	now := time.Now()
	access, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      strconv.FormatInt(uid, 10),
		"username": username,
		"role":     role,
		"iat":      now.Unix(),
		"exp":      now.Add(m.accessTTL).Unix(),
	}).SignedString([]byte(jwtSecret()))
	if err != nil {
		return "", "", err
	}
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	refresh = hex.EncodeToString(b)
	m.rdb.Set(ctx, "refresh:"+refresh, strconv.FormatInt(uid, 10), m.refreshTTL)
	return access, refresh, nil
}

// Parse 校验 access_token 并返回声明。
func (m *AuthManager) Parse(token string) (Claims, error) {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauth
		}
		return []byte(jwtSecret()), nil
	})
	if err != nil || !tok.Valid {
		return Claims{}, ErrUnauth
	}
	sub, err := tok.Claims.GetSubject()
	if err != nil {
		return Claims{}, ErrUnauth
	}
	var uid int64
	for _, c := range sub {
		if c < '0' || c > '9' {
			return Claims{}, ErrUnauth
		}
		uid = uid*10 + int64(c-'0')
	}
	if uid == 0 {
		return Claims{}, ErrUnauth
	}
	c := Claims{UID: uid}
	if m, ok := tok.Claims.(jwt.MapClaims); ok {
		if s, ok := m["username"].(string); ok {
			c.Username = s
		}
		if f, ok := m["role"].(float64); ok {
			c.Role = int32(f)
		}
	}
	return c, nil
}

// ConsumeRefresh 用 refresh_token 换 uid；GETDEL 原子轮换（旧 token 重放即失效）。
func (m *AuthManager) ConsumeRefresh(ctx context.Context, token string) (int64, error) {
	uidStr, err := m.rdb.GetDel(ctx, "refresh:"+token).Result()
	if err != nil || uidStr == "" {
		return 0, ErrToken
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil || uid == 0 {
		return 0, ErrToken
	}
	return uid, nil
}
