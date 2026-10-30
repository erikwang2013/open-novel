package biz

// 章节目录缓存失效测试：新建章节后 chapter:list:{book}:* 必须立即失效，
// 其他书籍的目录缓存不受影响。需本地 MySQL(3307)/Redis(6380)。

import (
	"context"
	"fmt"
	"testing"
	"time"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
)

func TestCreateChapterInvalidatesListCache(t *testing.T) {
	d, err := data.NewData(&conf.Data{
		DbDsn:          "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:      "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })

	uc := NewChapterUsecase(d, nil)
	ctx := context.Background()
	const bookID, otherBook = uint64(990001), uint64(990002)
	key := fmt.Sprintf("chapter:list:%d:1:20", bookID)
	otherKey := fmt.Sprintf("chapter:list:%d:1:20", otherBook)
	d.Cache.Set(ctx, key, "stale", time.Minute)
	d.Cache.Set(ctx, otherKey, "keep", time.Minute)
	t.Cleanup(func() {
		d.Cache.Del(ctx, key, otherKey)
		var ids []uint64
		d.DB.Model(&data.Chapter{}).Where("book_id = ?", bookID).Pluck("id", &ids)
		if len(ids) > 0 {
			d.DB.Where("chapter_id IN ?", ids).Delete(&data.ChapterContent{})
		}
		d.DB.Where("book_id = ?", bookID).Delete(&data.Chapter{})
	})

	if _, err := uc.CreateChapter(ctx, bookID, 1, "缓存失效测试", "正文", "zh-CN"); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Cache.Get(ctx, key); ok {
		t.Fatal("chapter:list cache must be invalidated after CreateChapter")
	}
	if v, ok := d.Cache.Get(ctx, otherKey); !ok || v != "keep" {
		t.Fatalf("unrelated book cache must remain, got %q ok=%v", v, ok)
	}
}
