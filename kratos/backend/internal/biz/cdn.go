package biz

// CDN 章节静态化机制（默认关闭）：
//   - 门控：env CDN_BASE_URL 非空才启用；为空时服务行为与未实现完全一致。
//   - 缓存：免费章节 content 响应带 Cache-Control: public, s-maxage=3600（CDN 回源缓存，
//     回源写 CDN，无显式写操作）；VIP 章节强制 no-store（鉴权内容不静态化）。
//   - 失效：env CDN_PURGE_URL 非空时，章节创建/状态变更后对该章节发起 purge，
//     URL 模板 {key} 替换为 CDN key：chapter/{id}?lang={lang}（先只做章节级 key，
//     书籍级预留：book/{id}?lang={lang}）。
//   - purge 为 fire-and-forget goroutine + 5s 超时，失败仅记日志，不阻塞、不对外报错。

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// cdnLog 与 cmd 入口同一 stdout logger；purge 为 best-effort，无需注入依赖。
var cdnLog = log.NewStdLogger(os.Stdout)

// CdnEnabled CDN_BASE_URL 非空 = 设置缓存头 + 参与失效；空 = 完全禁用。
func CdnEnabled() bool { return os.Getenv("CDN_BASE_URL") != "" }

// ChapterCacheControl 免费章节可共享缓存 1h；VIP 章节禁止缓存（鉴权内容）。
func ChapterCacheControl(isVip bool) string {
	if isVip {
		return "no-store"
	}
	return "public, s-maxage=3600"
}

// ChapterKey CDN 对象 key 约定：chapter/{id}?lang={lang}。
func ChapterKey(id uint64, lang string) string { return fmt.Sprintf("chapter/%d?lang=%s", id, lang) }

// PurgeChapterAsync 章节级 CDN 失效：fire-and-forget goroutine + 5s 超时，失败仅记日志。
// ponytail: best-effort——purge 延迟/丢失只让旧内容多缓存至多 1h（s-maxage 到期自然过期）；
// CDN 未启用或未配 CDN_PURGE_URL 时直接跳过；如需可靠失效（消费队列+重试）再升级。
func PurgeChapterAsync(chapterID uint64, lang string) {
	if !CdnEnabled() {
		return
	}
	url := os.Getenv("CDN_PURGE_URL")
	if url == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		key := ChapterKey(chapterID, lang)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.ReplaceAll(url, "{key}", key), nil)
		if err != nil {
			cdnLog.Log(log.LevelWarn, "msg", "cdn purge build request failed", "key", key, "err", err.Error())
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cdnLog.Log(log.LevelWarn, "msg", "cdn purge failed", "key", key, "err", err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			cdnLog.Log(log.LevelWarn, "msg", "cdn purge non-2xx", "key", key, "status", resp.StatusCode)
		}
	}()
}
