package cdn

// 阿里云 CDN adapter（§五 5.3）：RPC form（公共参数 + Action=RefreshObjectCaches），
// HMAC-SHA1 签名（RFC3986 编码，StringToSign = "GET&%2F&" + 编码查询串）。
// 批量 ≤1000/批，限速 50qps（token bucket），每日 10000 URL 预警（8000 起 warn）。

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const aliyunVersion = "2014-11-11"

type aliyunProvider struct {
	endpoint  string
	accessKey string
	secretKey string
	batchSize int
	bucket    *tokenBucket
	counter   *dailyCounter
}

// NewAliyun factory（统一签名 §3.3）：缺 access_key_id/access_key_secret 返回 error。
func NewAliyun(cfg map[string]any) (Provider, error) {
	ak := cfgString(cfg, "access_key_id")
	sk := cfgString(cfg, "access_key_secret")
	if ak == "" || sk == "" {
		return nil, fmt.Errorf("aliyun: access_key_id/access_key_secret required")
	}
	endpoint := "https://cdn.aliyuncs.com/"
	if v := cfgString(cfg, "endpoint"); v != "" { // 测试端点旋钮
		endpoint = v
	}
	return &aliyunProvider{endpoint: endpoint, accessKey: ak, secretKey: sk,
		batchSize: cfgInt(cfg, "batch_size", 1000),
		bucket:    newTokenBucket(cfgFloat(cfg, "rate_limit_qps", 50)),
		counter:   newDailyCounter(8000, warnLog)}, nil
}

func (p *aliyunProvider) Name() string { return "aliyun" }

func (p *aliyunProvider) Purge(ctx context.Context, keys []string) error {
	for _, batch := range Split(keys, p.batchSize) {
		if err := p.bucket.Wait(ctx); err != nil {
			return err
		}
		if err := p.purgeBatch(ctx, batch); err != nil {
			return err
		}
		p.counter.Add(len(batch))
	}
	return nil
}

func (p *aliyunProvider) purgeBatch(ctx context.Context, keys []string) error {
	params := map[string]string{
		"AccessKeyId":      p.accessKey,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   uuid.NewString(),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Format":           "JSON",
		"Version":          aliyunVersion,
		"Action":           "RefreshObjectCaches",
		"ObjectType":       "File",
		"ObjectPath":       strings.Join(keys, "\n"),
	}
	params["Signature"] = aliyunSign(p.secretKey, params)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return non2xx(resp.StatusCode)
	}
	return nil
}

// aliyunSign RPC 签名（§五 5.3）：
// 参数名 ASCII 排序 → RFC3986 percent-encode → StringToSign = "GET&%2F&" + 编码结果
// → HMAC-SHA1(secret + "&") → base64。
func aliyunSign(secret string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(escapeRFC3986(k))
		sb.WriteString("=")
		sb.WriteString(escapeRFC3986(params[k]))
	}
	strToSign := "GET&%2F&" + escapeRFC3986(sb.String())
	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(strToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// escapeRFC3986 percent-encode：url.QueryEscape + "+"→"%20"（RFC3986 空格编码）。
func escapeRFC3986(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
