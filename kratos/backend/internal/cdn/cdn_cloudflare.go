package cdn

// Cloudflare adapter（§五 5.1）：POST /zones/{zone_id}/purge_cache，Bearer 认证。
// 批量 ≤30/请求（config batch_size 可调），多批串行；200 + success=true 才算成功。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const cfDefaultBatch = 30

type cloudflareProvider struct {
	baseURL   string
	zoneID    string
	apiToken  string
	batchSize int
	bucket    *tokenBucket
}

// NewCloudflare factory（统一签名 §3.3）：缺 zone_id/api_token 返回 error。
func NewCloudflare(cfg map[string]any) (Provider, error) {
	zoneID := cfgString(cfg, "zone_id")
	token := cfgString(cfg, "api_token")
	if zoneID == "" || token == "" {
		return nil, fmt.Errorf("cloudflare: zone_id/api_token required")
	}
	base := "https://api.cloudflare.com/client/v4"
	if v := cfgString(cfg, "base_url"); v != "" { // 测试端点旋钮
		base = v
	}
	return &cloudflareProvider{baseURL: base, zoneID: zoneID, apiToken: token,
		batchSize: cfgInt(cfg, "batch_size", cfDefaultBatch),
		bucket:    newTokenBucket(1000)}, nil
}

func (p *cloudflareProvider) Name() string { return "cloudflare" }

func (p *cloudflareProvider) Purge(ctx context.Context, keys []string) error {
	for _, batch := range Split(keys, p.batchSize) {
		if err := p.bucket.Wait(ctx); err != nil {
			return err
		}
		if err := p.purgeBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

func (p *cloudflareProvider) purgeBatch(ctx context.Context, keys []string) error {
	body, _ := json.Marshal(map[string]any{"files": keys})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.baseURL+"/zones/"+p.zoneID+"/purge_cache", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return non2xx(resp.StatusCode)
	}
	var out struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if !out.Success {
		return fmt.Errorf("cloudflare: success=false")
	}
	return nil
}
