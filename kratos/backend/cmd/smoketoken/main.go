// smoketoken 为 scripts/smoke.sh 签发 access JWT（与服务端 pkg.AuthManager 同算法/claims）。
// 用法：JWT_SECRET=xxx ROLE=2 SMOKE_UID=1001 go run ./cmd/smoketoken
// 注意：不能用 UID，bash 的 UID 是只读变量。
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"open-novel/backend/internal/pkg"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = pkg.DefaultJWTSecret
	}
	role := int64(pkg.RoleAuthor)
	if v := os.Getenv("ROLE"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			fatal("ROLE 必须是整数")
		}
		role = n
	}
	uid := int64(1001)
	if v := os.Getenv("SMOKE_UID"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			fatal("SMOKE_UID 必须是正整数")
		}
		uid = n
	}
	now := time.Now()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      strconv.FormatInt(uid, 10),
		"username": "smoke",
		"role":     role,
		"iat":      now.Unix(),
		"exp":      now.Add(30 * time.Minute).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		fatal(err.Error())
	}
	fmt.Println(token)
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "smoketoken:", msg)
	os.Exit(1)
}
