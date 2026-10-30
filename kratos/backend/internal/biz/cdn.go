package biz

// CDN 门面：章节静态化 + 单一 generic purge 端点（CDN_BASE_URL / CDN_PURGE_URL）。
// 多厂商 provider 管理已移除；切换厂商在 DNS/边缘侧完成，这里只保留通用失效模板。

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"open-novel/backend/internal/cdn"
)

// cdnLog 章节用例共用的 stdout logger（阅读日志 / purge 失败均 best-effort 记录）。
var cdnLog = log.NewStdLogger(os.Stdout)

// currentManager 由 CDN_PURGE_URL 构造 generic Manager；未配置时返回 nil（失效为空操作）。
func currentManager() *cdn.Manager {
	u := os.Getenv("CDN_PURGE_URL")
	if u == "" {
		return nil
	}
	p, err := cdn.NewGeneric(map[string]any{"url_template": u})
	if err != nil {
		return nil
	}
	return cdn.NewManager([]cdn.Provider{p})
}

// CdnEnabled 门控：配置了 CDN base URL 或 purge 端点即启用。
func CdnEnabled() bool {
	return os.Getenv("CDN_BASE_URL") != "" || currentManager() != nil
}

// ChapterCacheControl 免费章节可共享缓存 1h；VIP 章节禁止缓存（鉴权内容）。
func ChapterCacheControl(isVip bool) string {
	if isVip {
		return "no-store"
	}
	return "public, s-maxage=3600"
}

// ChapterKey CDN 对象 key 约定：chapter/{id}?lang={lang}。
func ChapterKey(id uint64, lang string) string { return fmt.Sprintf("chapter/%d?lang=%s", id, lang) }

// PurgeChapterAsync 章节级失效：单 lang 单 key。
func PurgeChapterAsync(chapterID uint64, lang string) {
	PurgeChaptersAsync(chapterID, []string{lang})
}

// PurgeChaptersAsync 一次广播多 lang 章节 key；fire-and-forget + 5s 超时，失败仅记日志。
func PurgeChaptersAsync(chapterID uint64, langs []string) {
	if len(langs) == 0 {
		return
	}
	keys := make([]string, 0, len(langs))
	for _, l := range langs {
		keys = append(keys, ChapterKey(chapterID, l))
	}
	m := currentManager()
	if m == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.Purge(ctx, keys)
	}()
}
