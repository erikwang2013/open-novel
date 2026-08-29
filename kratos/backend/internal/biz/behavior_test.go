package biz

// 阅读行为聚合测试：真实 MySQL（对齐 recommend_test.go 模式）。
// 断言不依赖具体时刻：仅验证当天/7 天边界、排序与分布长度。

import (
	"context"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func newTestBehaviorBiz(t *testing.T) (*BehaviorUsecase, *data.Data) {
	t.Helper()
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewBehaviorUsecase(d), d
}

// ensureReadingLogTable 幂等建表（对齐 sql/init.sql 的 DDL），避免依赖初始化脚本已执行。
func ensureReadingLogTable(t *testing.T, d *data.Data) {
	t.Helper()
	if err := d.DB.Exec(`CREATE TABLE IF NOT EXISTS novel_reading_log (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  book_id    BIGINT UNSIGNED NOT NULL,
  chapter_id BIGINT UNSIGNED NOT NULL,
  lang       CHAR(5) NOT NULL DEFAULT 'zh-CN',
  position   INT UNSIGNED NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_time (user_id, created_at),
  KEY idx_book_time (book_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		t.Fatal(err)
	}
}

// TestBehaviorStats: 当日/7 天活跃读者、热门书籍排序与本地化书名、24 小时分布。
func TestBehaviorStats(t *testing.T) {
	uc, d := newTestBehaviorBiz(t)
	ensureReadingLogTable(t, d)
	ctx := context.Background()
	now := time.Now() // 本机为 UTC+8，与 biz 的 cst 口径一致

	marker := "coder72b_bhv"
	uToday := []int64{99101, 99102}
	u7d := int64(99103)
	uOld := int64(99104)

	books := []data.Book{
		{Title: marker + "_A", Author: marker, Lang: "zh-CN", Status: 1},
		{Title: marker + "_B", Author: marker, Lang: "zh-CN", Status: 1},
		{Title: marker + "_C", Author: marker, Lang: "zh-CN", Status: 1},
	}
	for i := range books {
		if err := d.DB.Create(&books[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	b1, b2, b3 := books[0].ID, books[1].ID, books[2].ID
	// en 本地化书名只给 A：验证书名回退逻辑
	if err := d.DB.Create(&data.BookTranslation{BookID: b1, Lang: "en", Title: marker + "_A_en"}).Error; err != nil {
		t.Fatal(err)
	}

	// 当日：u1 读 A×3、B×2；u2 读 A×1、C×1 → 当日事件 7，活跃 2，热度 A=4 B=2 C=1
	today := []data.ReadingLog{
		{UserID: uint64(uToday[0]), BookID: b1, ChapterID: 1, Lang: "zh-CN"},
		{UserID: uint64(uToday[0]), BookID: b1, ChapterID: 2, Lang: "zh-CN"},
		{UserID: uint64(uToday[0]), BookID: b1, ChapterID: 3, Lang: "zh-CN"},
		{UserID: uint64(uToday[0]), BookID: b2, ChapterID: 1, Lang: "zh-CN"},
		{UserID: uint64(uToday[0]), BookID: b2, ChapterID: 2, Lang: "zh-CN"},
		{UserID: uint64(uToday[1]), BookID: b1, ChapterID: 4, Lang: "zh-CN"},
		{UserID: uint64(uToday[1]), BookID: b3, ChapterID: 1, Lang: "zh-CN"},
	}
	for i := range today {
		today[i].CreatedAt = now
		if err := d.DB.Create(&today[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	// 7 天内（第 6 天）：u3 读 B×1 → 7 天活跃 3
	if err := d.DB.Create(&data.ReadingLog{UserID: uint64(u7d), BookID: b2, ChapterID: 1, CreatedAt: now.AddDate(0, 0, -6)}).Error; err != nil {
		t.Fatal(err)
	}
	// 8 天前：u4 → 应被排除
	if err := d.DB.Create(&data.ReadingLog{UserID: uint64(uOld), BookID: b3, ChapterID: 1, CreatedAt: now.AddDate(0, 0, -8)}).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ids := []uint64{b1, b2, b3}
		d.DB.Where("book_id IN ?", ids).Delete(&data.BookTranslation{})
		d.DB.Where("user_id IN ?", append(uToday, u7d, uOld)).Delete(&data.ReadingLog{})
		d.DB.Delete(&data.Book{}, ids)
	})

	st, err := uc.Stats(ctx, "en")
	if err != nil {
		t.Fatal(err)
	}
	if st.ActiveReaders != 2 {
		t.Fatalf("active_readers: want 2, got %d", st.ActiveReaders)
	}
	if st.Readers7d != 3 {
		t.Fatalf("readers_7d: want 3, got %d", st.Readers7d)
	}
	// 热度排序 + en 书名本地化
	// 热度口径为近 7 天：B 含 6 天前那次，故 B=3
	want := []struct {
		book  uint64
		count int64
		title string
	}{{b1, 4, marker + "_A_en"}, {b2, 3, marker + "_B"}, {b3, 1, marker + "_C"}}
	if len(st.HotBooks) != 3 {
		t.Fatalf("hot_books: want 3, got %d", len(st.HotBooks))
	}
	for i, w := range want {
		hb := st.HotBooks[i]
		if hb.BookID != w.book || hb.Count != w.count || hb.Title != w.title {
			t.Fatalf("hot_books[%d]: want (%d,%d,%s), got (%d,%d,%s)",
				i, w.book, w.count, w.title, hb.BookID, hb.Count, hb.Title)
		}
	}
	// 24 小时分布：长度恒 24，当日事件总数 = 7
	var sum int64
	for i, v := range st.Hourly {
		if i >= 24 {
			t.Fatalf("hourly length > 24")
		}
		sum += v
	}
	if len(st.Hourly) != 24 || sum != 7 {
		t.Fatalf("hourly: want len 24 sum 7, got len %d sum %d", len(st.Hourly), sum)
	}
}
