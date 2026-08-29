package service

// 管理端翻译 rpc 测试：requireAdmin 守卫 + 翻译调用（httptest mock DeepL + 真实 DB）。

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	adminv1 "open-novel/backend/api/admin/v1"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

const translateSvcMarker = "coder72b_trsvc"

// TestTranslateBookRPC requireAdmin 守卫（匿名 140401 / 普通用户 180401）+ 成功翻译入库。
func TestTranslateBookRPC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "DeepL-Auth-Key test-key" {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		var req struct {
			Text       []string `json:"text"`
			TargetLang string   `json:"target_lang"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		out := make([]map[string]string, len(req.Text))
		for i, txt := range req.Text {
			out[i] = map[string]string{"text": fmt.Sprintf("[%s] %s", req.TargetLang, txt)}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"translations": out})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("TRANSLATE_BASE_URL", srv.URL)
	t.Setenv("TRANSLATE_API_KEY", "test-key")

	s, d := newTestAdmin(t)
	ctx := context.Background()
	adminCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 3})

	b := data.Book{Title: translateSvcMarker + "书", Summary: translateSvcMarker + "简介", Lang: "zh-CN", Status: 1}
	if err := d.DB.WithContext(ctx).Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.WithContext(ctx).Where("book_id = ?", b.ID).Delete(&data.BookTranslation{})
		d.DB.WithContext(ctx).Delete(&data.Book{}, b.ID)
	})

	// 匿名 140401；普通用户 180401
	if _, err := s.TranslateBook(ctx, &adminv1.TranslateBookReq{BookId: int64(b.ID), Lang: "en"}); err == nil || kerrors.FromError(err).Code != 140401 {
		t.Fatalf("anonymous: want 140401, got %v", err)
	}
	readerCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 1})
	if _, err := s.TranslateBook(readerCtx, &adminv1.TranslateBookReq{BookId: int64(b.ID), Lang: "en"}); err == nil || kerrors.FromError(err).Code != 180401 {
		t.Fatalf("reader: want 180401, got %v", err)
	}

	// 管理员翻译成功，标题/简介带 target_lang 前缀
	reply, err := s.TranslateBook(adminCtx, &adminv1.TranslateBookReq{BookId: int64(b.ID), Lang: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(reply.Title, "[EN]") || !strings.HasPrefix(reply.Summary, "[EN]") {
		t.Fatalf("reply: %+v", reply)
	}
	var row data.BookTranslation
	if err := d.DB.WithContext(ctx).Where("book_id = ? AND lang = ?", b.ID, "en").First(&row).Error; err != nil {
		t.Fatalf("translation not persisted: %v", err)
	}
	if row.Title != reply.Title {
		t.Fatalf("persisted title mismatch: %q vs %q", row.Title, reply.Title)
	}
}
