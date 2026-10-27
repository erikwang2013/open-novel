package cdn

// Generic adapter：唯一的 CDN 失效通道，由 CDN_PURGE_URL 的 {key} 模板驱动。
// 逐 key 单请求，1 key/请求保证 SetChapterStatus 的 2 lang → 2 请求语义。

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type genericProvider struct{ template string }

// NewGeneric factory（统一签名 §3.3）：url_template 必须含 {key}。
func NewGeneric(cfg map[string]any) (Provider, error) {
	tpl := cfgString(cfg, "url_template")
	if !strings.Contains(tpl, "{key}") {
		return nil, fmt.Errorf("generic: url_template must contain {key}")
	}
	return &genericProvider{template: tpl}, nil
}

func (p *genericProvider) Name() string { return "generic" }

func (p *genericProvider) Purge(ctx context.Context, keys []string) error {
	// 单 key 失败即中止剩余（合批后语义）：上层按厂商整批重试，best-effort 可接受。
	for _, k := range keys {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			strings.ReplaceAll(p.template, "{key}", k), nil)
		if err != nil {
			return err
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			return non2xx(resp.StatusCode)
		}
	}
	return nil
}
