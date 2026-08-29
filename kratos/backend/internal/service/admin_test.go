package service

import (
	"context"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"

	adminv1 "open-novel/backend/api/admin/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func newTestAdmin(t *testing.T) (*AdminService, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewAdminService(biz.NewAdminUsecase(d, nil)), d
}

// TestListAuditLogs: requireAdmin → 分页/条件筛选 → created_at 倒序 → 字段透出。
func TestListAuditLogs(t *testing.T) {
	s, d := newTestAdmin(t)
	ctx := context.Background()
	adminCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 3})

	// 隔离 fixture：独立 action/target_type + detail 标记，避免命中真实数据
	const marker = "coder72b_test"
	now := time.Now()
	uid1, uid2 := int64(11), int64(22)
	fixtures := []data.AuditLog{
		{UserID: &uid1, Action: marker + "_a", TargetType: marker, TargetID: "1", Detail: marker, CreatedAt: now.Add(-3 * time.Hour)},
		{UserID: &uid2, Action: marker + "_b", TargetType: marker, TargetID: "2", Detail: marker, CreatedAt: now.Add(-2 * time.Hour)},
		{UserID: &uid1, Action: marker + "_b", TargetType: marker, TargetID: "3", Detail: marker, CreatedAt: now.Add(-1 * time.Hour)},
	}
	for i := range fixtures {
		if err := d.DB.WithContext(ctx).Create(&fixtures[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	ids := []int64{int64(fixtures[0].ID), int64(fixtures[1].ID), int64(fixtures[2].ID)}
	t.Cleanup(func() { d.DB.WithContext(ctx).Delete(&data.AuditLog{}, ids) })

	// 匿名 140401；普通用户 180401
	if _, err := s.ListAuditLogs(ctx, &adminv1.ListAuditLogsReq{}); err == nil || kerrors.FromError(err).Code != 140401 {
		t.Fatalf("anonymous: want 140401, got %v", err)
	}
	readerCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 1})
	if _, err := s.ListAuditLogs(readerCtx, &adminv1.ListAuditLogsReq{}); err == nil || kerrors.FromError(err).Code != 180401 {
		t.Fatalf("reader: want 180401, got %v", err)
	}

	// 无筛选：total >= 3，created_at 倒序
	all, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if all.Total < 3 {
		t.Fatalf("total: want >=3, got %d", all.Total)
	}
	found := map[int64]bool{}
	for _, a := range all.List {
		if a.Detail == marker {
			found[a.Id] = true
			if a.TargetId != 0 && a.CreatedAt == "" {
				t.Fatalf("fixture %d: target_id/created_at not filled", a.Id)
			}
		}
	}
	for _, id := range ids {
		if !found[id] {
			t.Fatalf("fixture %d missing in no-filter list", id)
		}
	}
	for i := 1; i < len(all.List); i++ {
		if all.List[i-1].CreatedAt < all.List[i].CreatedAt {
			t.Fatal("list not sorted by created_at DESC")
		}
	}

	// user_id + action 组合筛选
	ua, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{UserId: uid1, Action: marker + "_b"})
	if err != nil {
		t.Fatal(err)
	}
	if ua.Total != 1 || len(ua.List) != 1 || ua.List[0].TargetId != 3 {
		t.Fatalf("user+action filter: want 1 row target 3, got total=%d target=%d", ua.Total, ua.List[0].TargetId)
	}

	// target_type + 时间范围（起止包住中间一条）
	from := now.Add(-150 * time.Minute).Format(time.RFC3339)
	to := now.Add(-90 * time.Minute).Format(time.RFC3339)
	tr, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{TargetType: marker, StartTime: from, EndTime: to})
	if err != nil {
		t.Fatal(err)
	}
	if tr.Total != 1 || tr.List[0].TargetId != 2 {
		t.Fatalf("time range: want target 2, got total=%d target=%d", tr.Total, tr.List[0].TargetId)
	}

	// 分页：page_size=2 仅返回 2 条，且第一页与第二页不重叠
	pg, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{TargetType: marker, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(pg.List) != 2 {
		t.Fatalf("page1: want 2 rows, got %d", len(pg.List))
	}
	pg2, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{TargetType: marker, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(pg2.List) != 1 {
		t.Fatalf("page2: want 1 row, got %d", len(pg2.List))
	}
	if pg.List[0].Id == pg2.List[0].Id || pg.List[1].Id == pg2.List[0].Id {
		t.Fatal("pages overlap")
	}
}
