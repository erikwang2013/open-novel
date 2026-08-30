package biz

// 报表聚合测试：真实 MySQL（对齐 behavior_test.go 模式）。
// 断言窗口用 2015 年——平台数据均为 2026 年，窗口内只有本测试 fixture，保证精确断言；
// 全局量（total_users/comment_count/report_count）与其他测试共享库，用 >= 断言。

import (
	"context"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func newTestAdminBiz(t *testing.T) (*AdminUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewAdminUsecase(d, nil), d
}

// TestReportsAggregation: 日期范围/缺日补零/channel 分组/金额累加(分)/套餐 join。
func TestReportsAggregation(t *testing.T) {
	uc, d := newTestAdminBiz(t)
	ctx := context.Background()
	marker := "rpt_biz_test"
	// 2015 年窗口：5 天，fixture 落在 01-02/01-04
	const start, end = "2015-01-01", "2015-01-05"
	dt := func(day int) time.Time { return time.Date(2015, 1, day, 12, 0, 0, 0, cst) }

	// 支付订单：01-02 两笔已支付(12.34+5.00)，01-04 一笔(3.50)；01-02 一笔未支付；窗口外 01-06/2014-12-31 各一笔
	orders := []data.PaymentOrder{
		{OrderNo: marker + "_o1", UserID: 1, Amount: 12.34, Currency: "USD", Provider: marker + "_a", Status: 1, CreatedAt: dt(2)},
		{OrderNo: marker + "_o2", UserID: 1, Amount: 5.00, Currency: "USD", Provider: marker + "_a", Status: 1, CreatedAt: dt(2)},
		{OrderNo: marker + "_o3", UserID: 1, Amount: 3.50, Currency: "USD", Provider: marker + "_b", Status: 1, CreatedAt: dt(4)},
		{OrderNo: marker + "_o4", UserID: 1, Amount: 9.99, Currency: "USD", Provider: marker + "_a", Status: 0, CreatedAt: dt(2)},
		{OrderNo: marker + "_o5", UserID: 1, Amount: 1.00, Currency: "USD", Provider: marker + "_a", Status: 1, CreatedAt: dt(6)},
		{OrderNo: marker + "_o6", UserID: 1, Amount: 2.00, Currency: "USD", Provider: marker + "_a", Status: 1, CreatedAt: time.Date(2014, 12, 31, 12, 0, 0, 0, cst)},
	}
	// VIP 订单：01-02 marker+_noplan 9.99（无套餐行→回退 plan 码），01-04 marker 套餐 19.99（join 到套餐行）
	vipOrders := []data.VipOrder{
		{OrderNo: marker + "_v1", UserID: 1, Plan: marker + "_noplan", Amount: 9.99, Currency: "USD", Status: 1, CreatedAt: dt(2)},
		{OrderNo: marker + "_v2", UserID: 1, Plan: marker + "_plan", Amount: 19.99, Currency: "USD", Status: 1, CreatedAt: dt(4)},
		{OrderNo: marker + "_v3", UserID: 1, Plan: marker + "_plan", Amount: 1.00, Currency: "USD", Status: 0, CreatedAt: dt(4)},
	}
	plans := []data.VipPlan{
		{PlanCode: marker + "_plan", Days: 30, AmountCents: 1999, Currency: "USD", Label: marker + "_label", Status: 1},
	}
	users := []data.User{
		{Username: marker + "_u1", Email: marker + "_u1@t.co", PasswordHash: "x", Status: 1, CreatedAt: dt(2)},
		{Username: marker + "_u2", Email: marker + "_u2@t.co", PasswordHash: "x", Status: 1, CreatedAt: dt(3)},
		{Username: marker + "_u3", Email: marker + "_u3@t.co", PasswordHash: "x", Status: 1, CreatedAt: time.Date(2014, 12, 31, 12, 0, 0, 0, cst)},
	}
	books := []data.Book{
		{Title: marker + "_b1", Author: marker, Lang: "zh-CN", Status: 1, CreatedAt: dt(2)},
		{Title: marker + "_b2", Author: marker, Lang: "zh-CN", Status: 1, CreatedAt: dt(5)},
		{Title: marker + "_chbook", Author: marker, Lang: "zh-CN", Status: 1, CreatedAt: dt(3)},
	}
	comments := []data.Comment{
		{BookID: 1, UserID: 1, Content: marker + "_c1", Status: 2, ReportCount: 3, CreatedAt: dt(2)},
		{BookID: 1, UserID: 1, Content: marker + "_c2", Status: 1, ReportCount: 1, CreatedAt: dt(2)},
	}

	var ids struct {
		orders   []uint64
		vip      []uint64
		plans    []uint64
		users    []uint64
		books    []uint64
		chapters []uint64
		comments []uint64
	}
	for i := range orders {
		if err := d.DB.Create(&orders[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.orders = append(ids.orders, orders[i].ID)
	}
	for i := range vipOrders {
		if err := d.DB.Create(&vipOrders[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.vip = append(ids.vip, vipOrders[i].ID)
	}
	for i := range plans {
		if err := d.DB.Create(&plans[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.plans = append(ids.plans, plans[i].ID)
	}
	for i := range users {
		if err := d.DB.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.users = append(ids.users, users[i].ID)
	}
	for i := range books {
		if err := d.DB.Create(&books[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.books = append(ids.books, books[i].ID)
	}
	chapter := data.Chapter{BookID: books[0].ID, ChapterNo: 99001, Title: marker + "_chap", Status: 1, CreatedAt: dt(3)}
	if err := d.DB.Create(&chapter).Error; err != nil {
		t.Fatal(err)
	}
	ids.chapters = append(ids.chapters, chapter.ID)
	for i := range comments {
		if err := d.DB.Create(&comments[i]).Error; err != nil {
			t.Fatal(err)
		}
		ids.comments = append(ids.comments, comments[i].ID)
	}

	t.Cleanup(func() {
		d.DB.Delete(&data.PaymentOrder{}, ids.orders)
		d.DB.Delete(&data.VipOrder{}, ids.vip)
		d.DB.Delete(&data.VipPlan{}, ids.plans)
		d.DB.Delete(&data.User{}, ids.users)
		d.DB.Delete(&data.Chapter{}, ids.chapters)
		d.DB.Delete(&data.Book{}, ids.books)
		d.DB.Delete(&data.Comment{}, ids.comments)
	})

	r, err := uc.Reports(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}

	// 订单：total_count=3（未支付/窗口外排除），total_amount=2084 分（12.34+5.00+3.50）
	if r.Order.TotalCount != 3 || r.Order.TotalAmount != 2084 {
		t.Fatalf("order total: want (3,2084), got (%d,%d)", r.Order.TotalCount, r.Order.TotalAmount)
	}
	// by_date：5 天全量，01-02/01-04 有值，其余补零；金额分
	wantByDate := map[string][2]int64{
		"2015-01-01": {0, 0}, "2015-01-02": {2, 1734}, "2015-01-03": {0, 0},
		"2015-01-04": {1, 350}, "2015-01-05": {0, 0},
	}
	if len(r.Order.ByDate) != 5 {
		t.Fatalf("order by_date: want 5 days, got %d", len(r.Order.ByDate))
	}
	for i, p := range r.Order.ByDate {
		w, ok := wantByDate[p.Date]
		if !ok || p.Count != w[0] || p.Amount != w[1] {
			t.Fatalf("order by_date[%d]=%s: want %v, got count=%d amount=%d", i, p.Date, w, p.Count, p.Amount)
		}
		if i > 0 && !(r.Order.ByDate[i].Date > r.Order.ByDate[i-1].Date) {
			t.Fatalf("order by_date not ascending at %d", i)
		}
	}
	// by_channel：a=2 笔 1734 分，b=1 笔 350 分
	if len(r.Order.ByChannel) != 2 {
		t.Fatalf("order by_channel: want 2, got %d", len(r.Order.ByChannel))
	}
	ch := map[string][2]int64{}
	for _, c := range r.Order.ByChannel {
		ch[c.Channel] = [2]int64{c.Count, c.Amount}
	}
	if ch[marker+"_a"] != [2]int64{2, 1734} || ch[marker+"_b"] != [2]int64{1, 350} {
		t.Fatalf("order by_channel: got %v", ch)
	}

	// 用户：窗口内 01-02/01-03 各 1 人，其余 0；total_users 全局 >= 2
	if r.User.TotalUsers < 2 {
		t.Fatalf("user total: want >= 2, got %d", r.User.TotalUsers)
	}
	if len(r.User.ByDate) != 5 || r.User.ByDate[1].Count != 1 || r.User.ByDate[2].Count != 1 ||
		r.User.ByDate[0].Count != 0 || r.User.ByDate[3].Count != 0 || r.User.ByDate[4].Count != 0 {
		t.Fatalf("user by_date: got %+v", r.User.ByDate)
	}

	// VIP：total=(2, 2998 分=9.99+19.99)；by_plan 两行：套餐行(id>0,label)，monthly 回退(id=0,码)
	if r.Vip.TotalCount != 2 || r.Vip.TotalAmount != 2998 {
		t.Fatalf("vip total: want (2,2998), got (%d,%d)", r.Vip.TotalCount, r.Vip.TotalAmount)
	}
	if len(r.Vip.ByDate) != 5 || r.Vip.ByDate[1].Count != 1 || r.Vip.ByDate[1].Amount != 999 || r.Vip.ByDate[3].Amount != 1999 {
		t.Fatalf("vip by_date: got %+v", r.Vip.ByDate)
	}
	if len(r.Vip.ByPlan) != 2 {
		t.Fatalf("vip by_plan: want 2, got %d", len(r.Vip.ByPlan))
	}
	plansBy := map[string]PlanAmountPoint{}
	for _, p := range r.Vip.ByPlan {
		plansBy[p.PlanName] = p
	}
	if p := plansBy[marker+"_label"]; p.PlanID <= 0 || p.Count != 1 || p.Amount != 1999 {
		t.Fatalf("vip by_plan marker: got %+v", p)
	}
	if p := plansBy[marker+"_noplan"]; p.PlanID != 0 || p.Count != 1 || p.Amount != 999 {
		t.Fatalf("vip by_plan no-plan fallback: got %+v", p)
	}

	// 内容：书籍 01-02/01-05 各 1，章节 01-03 1 条，评论/举报为全局累计（>= fixture 贡献）
	if len(r.Content.BooksByDate) != 5 || r.Content.BooksByDate[1].Count != 1 || r.Content.BooksByDate[4].Count != 1 {
		t.Fatalf("books by_date: got %+v", r.Content.BooksByDate)
	}
	if len(r.Content.ChaptersByDate) != 5 || r.Content.ChaptersByDate[2].Count != 1 {
		t.Fatalf("chapters by_date: got %+v", r.Content.ChaptersByDate)
	}
	if r.Content.CommentCount < 2 {
		t.Fatalf("comment_count: want >= 2, got %d", r.Content.CommentCount)
	}
	// status=2 评论 report_count=3；status=1 的 report_count=1 不计
	if r.Content.ReportCount < 3 {
		t.Fatalf("report_count: want >= 3, got %d", r.Content.ReportCount)
	}
}

// TestNormalizeRange: 日期区间兜底推导（纯逻辑，无 DB）。
func TestNormalizeRange(t *testing.T) {
	today := time.Now().In(cst).Format("2006-01-02")
	wantStart := time.Now().In(cst).AddDate(0, 0, -29).Format("2006-01-02")
	s, e, err := normalizeRange("", "")
	if err != nil || s != wantStart || e != today {
		t.Fatalf("both empty: want (%s,%s), got (%s,%s,%v)", wantStart, today, s, e, err)
	}
	s, e, err = normalizeRange("", "2026-01-31")
	if err != nil || s != "2026-01-02" || e != "2026-01-31" {
		t.Fatalf("start empty: want (2026-01-02,2026-01-31), got (%s,%s,%v)", s, e, err)
	}
	s, e, err = normalizeRange("2026-02-01", "")
	if err != nil || s != "2026-02-01" || e != "2026-03-02" {
		t.Fatalf("end empty: want (2026-02-01,2026-03-02), got (%s,%s,%v)", s, e, err)
	}
	if _, _, err = normalizeRange("2026/01/01", "2026-01-31"); err == nil {
		t.Fatal("invalid start: want error")
	}
	if _, _, err = normalizeRange("2026-01-01", "oops"); err == nil {
		t.Fatal("invalid end: want error")
	}
	if err != pkg.ErrInvalidArgument {
		t.Fatalf("invalid date: want pkg.ErrInvalidArgument, got %v", err)
	}
}
