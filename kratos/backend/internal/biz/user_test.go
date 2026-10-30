package biz

// 邮箱可空唯一测试：空邮箱落 NULL（可多用户共存）、非空重复仍 1062、
// 清空后邮箱释放可被他人绑定。需本地 MySQL(3307)/Redis(6380)。

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func TestNullableEmail(t *testing.T) {
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })
	ctx := context.Background()
	uc := NewUserUsecase(d, nil)

	marker := "emailtest" + strconv.FormatInt(time.Now().UnixNano(), 10)
	t.Cleanup(func() { d.DB.Where("username LIKE ?", marker+"%").Delete(&data.User{}) })

	reg := func(suffix, email string) (UserProfile, error) {
		return uc.Register(ctx, marker+suffix, "password123", email, "")
	}
	// 空邮箱 → NULL，唯一键允许多行
	a, err := reg("a", "")
	if err != nil {
		t.Fatalf("empty email user a: %v", err)
	}
	if _, err := reg("b", ""); err != nil {
		t.Fatalf("second empty email user must be allowed: %v", err)
	}
	mail := marker + "@x.test"
	c, err := reg("c", mail)
	if err != nil {
		t.Fatalf("user c with email: %v", err)
	}
	if _, err := reg("d", mail); !errors.Is(err, pkg.ErrUserExists) {
		t.Fatalf("duplicate non-empty email: want ErrUserExists, got %v", err)
	}
	// 清空邮箱写 NULL
	empty := ""
	if _, err := uc.UpdateMe(ctx, int64(c.ID), UpdateProfile{Email: &empty}, ""); err != nil {
		t.Fatalf("clear email: %v", err)
	}
	var stored sql.NullString
	if err := d.DB.Model(&data.User{}).Where("id = ?", c.ID).Select("email").Scan(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Valid {
		t.Fatalf("cleared email must be NULL, got %q", stored.String)
	}
	// 释放后的邮箱可被他人绑定
	p, err := uc.UpdateMe(ctx, int64(a.ID), UpdateProfile{Email: &mail}, "")
	if err != nil {
		t.Fatalf("bind freed email: %v", err)
	}
	if p.Email != mail {
		t.Fatalf("want email %q, got %q", mail, p.Email)
	}
}
