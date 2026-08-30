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

	// 无筛选：total >= 3，created_at 倒序（PageSize 取上限，保证 fixture 在共享库分页内可见）
	all, err := s.ListAuditLogs(adminCtx, &adminv1.ListAuditLogsReq{PageSize: 100})
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

// TestGetReports: requireAdmin → 缺省近 30 天区间补零 → 非法日期 140400 → 字段透出。
func TestGetReports(t *testing.T) {
	s, _ := newTestAdmin(t)
	ctx := context.Background()
	adminCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 3})

	// 匿名 140401；普通用户 180401
	if _, err := s.GetReports(ctx, &adminv1.GetReportsReq{}); err == nil || kerrors.FromError(err).Code != 140401 {
		t.Fatalf("anonymous: want 140401, got %v", err)
	}
	readerCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 1})
	if _, err := s.GetReports(readerCtx, &adminv1.GetReportsReq{}); err == nil || kerrors.FromError(err).Code != 180401 {
		t.Fatalf("reader: want 180401, got %v", err)
	}
	// 非法日期
	if _, err := s.GetReports(adminCtx, &adminv1.GetReportsReq{StartDate: "2026/01/01", EndDate: "2026-01-31"}); err == nil || kerrors.FromError(err).Code != 140400 {
		t.Fatalf("bad date: want 140400, got %v", err)
	}
	// 缺省近 30 天：四组 by_date 均 30 天、升序、末位为今天
	rep, err := s.GetReports(adminCtx, &adminv1.GetReportsReq{})
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	checkDates := func(name string, pts []string) {
		t.Helper()
		if len(pts) != 30 {
			t.Fatalf("%s: want 30 days, got %d", name, len(pts))
		}
		if pts[0] > pts[len(pts)-1] || pts[len(pts)-1] != today {
			t.Fatalf("%s: range end=%s want %s", name, pts[len(pts)-1], today)
		}
	}
	checkDates("order", datesOfAmount(rep.OrderReport.ByDate))
	checkDates("user", datesOfCount(rep.UserReport.ByDate))
	checkDates("vip", datesOfAmount(rep.VipReport.ByDate))
	checkDates("books", datesOfCount(rep.ContentReport.BooksByDate))
	checkDates("chapters", datesOfCount(rep.ContentReport.ChaptersByDate))
	if rep.OrderReport.ByDate[0].Amount < 0 || rep.VipReport.TotalAmount < 0 {
		t.Fatal("amount fields must be >= 0")
	}
}

func datesOfAmount(pts []*adminv1.DateAmountPoint) []string {
	out := make([]string, len(pts))
	for i, p := range pts {
		out[i] = p.Date
	}
	return out
}

func datesOfCount(pts []*adminv1.DateCountPoint) []string {
	out := make([]string, len(pts))
	for i, p := range pts {
		out[i] = p.Date
	}
	return out
}

// TestGetStatsExt: 报表新字段（order/vip/user/评论举报）随 fixture 增量，且共享库下用 >= 断言。
func TestGetStatsExt(t *testing.T) {
	s, d := newTestAdmin(t)
	ctx := context.Background()
	adminCtx := pkg.WithClaims(ctx, pkg.Claims{UID: 100, Role: 3})
	marker := "stats_ext_test"

	base, err := s.GetStats(adminCtx, &adminv1.GetStatsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if base.OrderCount < 0 || base.OrderAmount < 0 || base.VipCount < 0 {
		t.Fatal("stats ext fields must be >= 0")
	}

	now := time.Now()
	exp := now.AddDate(0, 0, 30)
	var u1, u2 data.User
	u1 = data.User{Username: marker + "_u1", Email: marker + "_u1@t.co", PasswordHash: "x", Status: 1, CreatedAt: now}
	u2 = data.User{Username: marker + "_u2", Email: marker + "_u2@t.co", PasswordHash: "x", Status: 1, VipExpiresAt: &exp, CreatedAt: now.Add(-24 * time.Hour)}
	if err := d.DB.WithContext(ctx).Create(&u1).Error; err != nil {
		t.Fatal(err)
	}
	if err := d.DB.WithContext(ctx).Create(&u2).Error; err != nil {
		t.Fatal(err)
	}
	o := data.PaymentOrder{OrderNo: marker, UserID: u1.ID, Amount: 6.66, Currency: "USD", Provider: marker, Status: 1, CreatedAt: now}
	if err := d.DB.WithContext(ctx).Create(&o).Error; err != nil {
		t.Fatal(err)
	}
	c := data.Comment{BookID: 1, UserID: u1.ID, Content: marker, Status: 2, ReportCount: 3, CreatedAt: now}
	if err := d.DB.WithContext(ctx).Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		d.DB.WithContext(ctx).Delete(&data.Comment{}, c.ID)
		d.DB.WithContext(ctx).Delete(&data.PaymentOrder{}, o.ID)
		d.DB.WithContext(ctx).Delete(&data.User{}, []uint64{u1.ID, u2.ID})
	})

	after, err := s.GetStats(adminCtx, &adminv1.GetStatsReq{})
	if err != nil {
		t.Fatal(err)
	}
	if after.TodayNewUsers < base.TodayNewUsers+1 {
		t.Fatalf("today_new_users: want >= %d, got %d", base.TodayNewUsers+1, after.TodayNewUsers)
	}
	if after.VipCount < base.VipCount+1 {
		t.Fatalf("vip_count: want >= %d, got %d", base.VipCount+1, after.VipCount)
	}
	if after.OrderCount < base.OrderCount+1 {
		t.Fatalf("order_count: want >= %d, got %d", base.OrderCount+1, after.OrderCount)
	}
	if after.OrderAmount < base.OrderAmount+666 {
		t.Fatalf("order_amount: want >= %d, got %d", base.OrderAmount+666, after.OrderAmount)
	}
	if after.PendingComments < base.PendingComments+1 {
		t.Fatalf("pending_comments: want >= %d, got %d", base.PendingComments+1, after.PendingComments)
	}
	if after.PendingReports < base.PendingReports+3 {
		t.Fatalf("pending_reports: want >= %d, got %d", base.PendingReports+3, after.PendingReports)
	}
}
