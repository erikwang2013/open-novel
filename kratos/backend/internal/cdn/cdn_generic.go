package cdn

// Generic adapter（§五 5.5）：旧 CDN_PURGE_URL webhook 兼容（灰度期测试通道）。
// 逐 key 单请求，URL 模板 {key} 替换；1 key/请求保证 SetChapterStatus 旧断言（2 lang → 2 请求）不变。
// 退役时机：阶段 0 灰度完成（Cloudflare 单厂商稳定 1 周）后删除（§九决议 4）。

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
