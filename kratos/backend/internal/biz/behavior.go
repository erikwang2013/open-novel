package biz

// 阅读行为分析用例：novel_reading_log 聚合（管理端行为分析页，T-B-xx）。
// 时间口径 UTC+8（与 provider_unionpay txnTimeNow 一致；部署时区即 UTC+8）。

import (
	"context"
	"time"

	"gorm.io/gorm"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// cst 固定 UTC+8；DATETIME 列无时区，MySQL CURRENT_TIMESTAMP 按服务端时区写入。
var cst = time.FixedZone("CST", 8*3600)

type BehaviorUsecase struct {
	db *gorm.DB
}

func NewBehaviorUsecase(d *data.Data) *BehaviorUsecase {
	return &BehaviorUsecase{db: d.DB}
}

// BehaviorStats 阅读行为统计；Hourly 恒为 24 长度。
type BehaviorStats struct {
	ActiveReaders int64           // 当日 DISTINCT 阅读用户数
	Readers7d     int64           // 近 7 天 DISTINCT 阅读用户数
	HotBooks      []HotReadingBook // 近 7 天 TOP10
	Hourly        [24]int64        // 当日按小时阅读事件数
}

type HotReadingBook struct {
	BookID uint64
	Title  string
	Count  int64
}

// Stats 聚合阅读事件。lang 控制热门书籍书名（缺省 zh-CN，service 层兜底）。
func (uc *BehaviorUsecase) Stats(ctx context.Context, lang string) (BehaviorStats, error) {
	now := time.Now().In(cst)
	today := now.Format("2006-01-02")          // created_at >= 当天 00:00（UTC+8）
	weekStart := now.AddDate(0, 0, -6).Format("2006-01-02") // 近 7 天（含今日）
	var s BehaviorStats
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COUNT(DISTINCT user_id) FROM novel_reading_log WHERE created_at >= ?`, today).
		Scan(&s.ActiveReaders).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT COUNT(DISTINCT user_id) FROM novel_reading_log WHERE created_at >= ?`, weekStart).
		Scan(&s.Readers7d).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	// 近 7 天按书聚合；书名优先 lang 翻译，回退主书名，未匹配书行跳过
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT l.book_id, COALESCE(t.title, b.title) AS title, COUNT(*) AS count
		 FROM novel_reading_log l
		 LEFT JOIN novel_book b ON b.id = l.book_id
		 LEFT JOIN novel_book_translation t ON t.book_id = l.book_id AND t.lang = ?
		 WHERE l.created_at >= ? AND b.id IS NOT NULL
		 GROUP BY l.book_id, COALESCE(t.title, b.title)
		 ORDER BY count DESC, l.book_id LIMIT 10`, lang, weekStart).Scan(&s.HotBooks).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	// 当日按小时聚合；扫描填 24 数组，缺省 0
	var rows []struct {
		Hour int
		Cnt  int64
	}
	if err := uc.db.WithContext(ctx).Raw(
		`SELECT HOUR(created_at) AS hour, COUNT(*) AS cnt FROM novel_reading_log
		 WHERE created_at >= ? GROUP BY HOUR(created_at)`, today).Scan(&rows).Error; err != nil {
		return s, pkg.ErrAdminDB
	}
	for _, r := range rows {
		if r.Hour >= 0 && r.Hour < 24 {
			s.Hourly[r.Hour] = r.Cnt
		}
	}
	return s, nil
}
