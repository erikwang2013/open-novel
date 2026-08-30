package biz

// 报表用例：订单收入/用户增长/VIP 订阅/内容互动聚合（管理端报表页，requireAdmin）。
// 时间口径 UTC+8（同 behavior.go cst）；金额一律整数分（库中 DECIMAL(10,2) 元×100）。

import (
	"context"
	"math"
	"time"

	"open-novel/backend/internal/pkg"
)

// Report 聚合报表。ByDate 缺日补零，按日期升序。
type Report struct {
	Order   OrderReport
	User    UserReport
	Vip     VipReport
	Content ContentReport
}

type OrderReport struct {
	TotalCount  int64
	TotalAmount int64
	ByDate      []DateAmountPoint
	ByChannel   []ChannelAmountPoint
}

type UserReport struct {
	TotalUsers int64
	ByDate     []DateCountPoint
}

type VipReport struct {
	TotalCount  int64
	TotalAmount int64
	ByDate      []DateAmountPoint
	ByPlan      []PlanAmountPoint
}

type ContentReport struct {
	BooksByDate    []DateCountPoint
	ChaptersByDate []DateCountPoint
	CommentCount   int64
	ReportCount    int64
}

type DateAmountPoint struct {
	Date   string
	Count  int64
	Amount int64
}

type DateCountPoint struct {
	Date  string
	Count int64
}

type ChannelAmountPoint struct {
	Channel string
	Count   int64
	Amount  int64
}

type PlanAmountPoint struct {
	PlanID   int64
	PlanName string
	Count    int64
	Amount   int64
}

// Reports 聚合区间 [start, end]（YYYY-MM-DD，含首尾）。start/end 空串由本函数兜底。
func (uc *AdminUsecase) Reports(ctx context.Context, start, end string) (Report, error) {
	start, end, err := normalizeRange(start, end)
	if err != nil {
		return Report{}, err
	}
	endExcl := end + " 23:59:59" // DATETIME 含当天 23:59:59
	var r Report
	if err := uc.orderReport(ctx, start, endExcl, &r.Order); err != nil {
		return r, err
	}
	if err := uc.userReport(ctx, start, endExcl, &r.User); err != nil {
		return r, err
	}
	if err := uc.vipReport(ctx, start, endExcl, &r.Vip); err != nil {
		return r, err
	}
	if err := uc.contentReport(ctx, start, endExcl, &r.Content); err != nil {
		return r, err
	}
	return r, nil
}

// normalizeRange 缺省区间：均空=近 30 天（含首尾）；单侧空由对侧推导。日期须为 YYYY-MM-DD。
func normalizeRange(start, end string) (string, string, error) {
	parse := func(s string) (time.Time, error) { return time.ParseInLocation("2006-01-02", s, cst) }
	if start == "" && end == "" {
		end = time.Now().In(cst).Format("2006-01-02")
		start = time.Now().In(cst).AddDate(0, 0, -29).Format("2006-01-02")
		return start, end, nil
	}
	if start == "" {
		e, err := parse(end)
		if err != nil {
			return "", "", pkg.ErrInvalidArgument
		}
		return e.AddDate(0, 0, -29).Format("2006-01-02"), end, nil
	}
	if end == "" {
		s, err := parse(start)
		if err != nil {
			return "", "", pkg.ErrInvalidArgument
		}
		return start, s.AddDate(0, 0, 29).Format("2006-01-02"), nil
	}
	if _, err := parse(start); err != nil {
		return "", "", pkg.ErrInvalidArgument
	}
	if _, err := parse(end); err != nil {
		return "", "", pkg.ErrInvalidArgument
	}
	return start, end, nil
}

// eachDate 遍历 [start, end]（含首尾）每日。
func eachDate(start, end string, f func(d string)) {
	for d := start; ; {
		f(d)
		if d == end {
			return
		}
		dt, _ := time.ParseInLocation("2006-01-02", d, cst)
		d = dt.AddDate(0, 0, 1).Format("2006-01-02")
	}
}

// fillAmountDates / fillCountDates 聚合行填成每日一行，缺日补零。
func fillAmountDates(start, end string, rows []DateAmountPoint) []DateAmountPoint {
	m := make(map[string]DateAmountPoint, len(rows))
	for _, r := range rows {
		m[r.Date] = r
	}
	out := make([]DateAmountPoint, 0, 32)
	eachDate(start, end, func(d string) {
		p := m[d]
		p.Date = d
		out = append(out, p)
	})
	return out
}

func fillCountDates(start, end string, rows []DateCountPoint) []DateCountPoint {
	m := make(map[string]DateCountPoint, len(rows))
	for _, r := range rows {
		m[r.Date] = r
	}
	out := make([]DateCountPoint, 0, 32)
	eachDate(start, end, func(d string) {
		p := m[d]
		p.Date = d
		out = append(out, p)
	})
	return out
}

// toCents DECIMAL(10,2) 元 → 整数分（与 models.go“比较一律转整数分”口径一致）。
func toCents(amount float64) int64 { return int64(math.Round(amount * 100)) }

func (uc *AdminUsecase) orderReport(ctx context.Context, start, endExcl string, r *OrderReport) error {
	var total struct {
		Cnt    int64
		Amount float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) AS cnt, COALESCE(SUM(amount),0) AS amount
		 FROM novel_payment_order WHERE status = 1 AND created_at >= ? AND created_at <= ?`,
		start, endExcl).Scan(&total).Error; err != nil {
		return pkg.ErrAdminDB
	}
	r.TotalCount, r.TotalAmount = total.Cnt, toCents(total.Amount)
	var byDate []struct {
		Date   string
		Count  int64
		Amount float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(*) AS count, SUM(amount) AS amount
		 FROM novel_payment_order WHERE status = 1 AND created_at >= ? AND created_at <= ?
		 GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`, start, endExcl).Scan(&byDate).Error; err != nil {
		return pkg.ErrAdminDB
	}
	rows := make([]DateAmountPoint, 0, len(byDate))
	for _, x := range byDate {
		rows = append(rows, DateAmountPoint{Date: x.Date, Count: x.Count, Amount: toCents(x.Amount)})
	}
	r.ByDate = fillAmountDates(start, endExcl[:10], rows)
	var byChannel []struct {
		Channel string
		Count   int64
		Amount  float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT provider AS channel, COUNT(*) AS count, SUM(amount) AS amount
		 FROM novel_payment_order WHERE status = 1 AND created_at >= ? AND created_at <= ?
		 GROUP BY provider ORDER BY count DESC, channel`, start, endExcl).Scan(&byChannel).Error; err != nil {
		return pkg.ErrAdminDB
	}
	for _, x := range byChannel {
		r.ByChannel = append(r.ByChannel, ChannelAmountPoint{Channel: x.Channel, Count: x.Count, Amount: toCents(x.Amount)})
	}
	return nil
}

func (uc *AdminUsecase) userReport(ctx context.Context, start, endExcl string, r *UserReport) error {
	if err := uc.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM novel_user`).Scan(&r.TotalUsers).Error; err != nil {
		return pkg.ErrAdminDB
	}
	var byDate []DateCountPoint
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(*) AS count FROM novel_user
		 WHERE created_at >= ? AND created_at <= ? GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`,
		start, endExcl).Scan(&byDate).Error; err != nil {
		return pkg.ErrAdminDB
	}
	r.ByDate = fillCountDates(start, endExcl[:10], byDate)
	return nil
}

func (uc *AdminUsecase) vipReport(ctx context.Context, start, endExcl string, r *VipReport) error {
	var total struct {
		Cnt    int64
		Amount float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) AS cnt, COALESCE(SUM(amount),0) AS amount
		 FROM novel_vip_order WHERE status = 1 AND created_at >= ? AND created_at <= ?`,
		start, endExcl).Scan(&total).Error; err != nil {
		return pkg.ErrAdminDB
	}
	r.TotalCount, r.TotalAmount = total.Cnt, toCents(total.Amount)
	var byDate []struct {
		Date   string
		Count  int64
		Amount float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(*) AS count, SUM(amount) AS amount
		 FROM novel_vip_order WHERE status = 1 AND created_at >= ? AND created_at <= ?
		 GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`, start, endExcl).Scan(&byDate).Error; err != nil {
		return pkg.ErrAdminDB
	}
	rows := make([]DateAmountPoint, 0, len(byDate))
	for _, x := range byDate {
		rows = append(rows, DateAmountPoint{Date: x.Date, Count: x.Count, Amount: toCents(x.Amount)})
	}
	r.ByDate = fillAmountDates(start, endExcl[:10], rows)
	// 套餐维度：vip_order.plan 存套餐码，join novel_vip_plan 反查 id/label；无套餐行回退码本身
	var byPlan []struct {
		PlanID   int64
		PlanName string
		Count    int64
		Amount   float64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COALESCE(p.id,0) AS plan_id, COALESCE(p.label, v.plan) AS plan_name,
		        COUNT(*) AS count, SUM(v.amount) AS amount
		 FROM novel_vip_order v
		 LEFT JOIN novel_vip_plan p ON p.plan_code = v.plan
		 WHERE v.status = 1 AND v.created_at >= ? AND v.created_at <= ?
		 GROUP BY v.plan, p.id, p.label ORDER BY count DESC, v.plan`,
		start, endExcl).Scan(&byPlan).Error; err != nil {
		return pkg.ErrAdminDB
	}
	for _, x := range byPlan {
		r.ByPlan = append(r.ByPlan, PlanAmountPoint{PlanID: x.PlanID, PlanName: x.PlanName, Count: x.Count, Amount: toCents(x.Amount)})
	}
	return nil
}

func (uc *AdminUsecase) contentReport(ctx context.Context, start, endExcl string, r *ContentReport) error {
	var books, chapters []DateCountPoint
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(*) AS count FROM novel_book
		 WHERE created_at >= ? AND created_at <= ? GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`,
		start, endExcl).Scan(&books).Error; err != nil {
		return pkg.ErrAdminDB
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS date, COUNT(*) AS count FROM novel_chapter
		 WHERE created_at >= ? AND created_at <= ? GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')`,
		start, endExcl).Scan(&chapters).Error; err != nil {
		return pkg.ErrAdminDB
	}
	r.BooksByDate = fillCountDates(start, endExcl[:10], books)
	r.ChaptersByDate = fillCountDates(start, endExcl[:10], chapters)
	if err := uc.db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM novel_comment`).Scan(&r.CommentCount).Error; err != nil {
		return pkg.ErrAdminDB
	}
	// 待处理举报数：status=2 举报待审评论的 report_count 累计（同 GetStats pending_reports 口径）
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COALESCE(SUM(report_count),0) FROM novel_comment WHERE status = 2`).Scan(&r.ReportCount).Error; err != nil {
		return pkg.ErrAdminDB
	}
	return nil
}
