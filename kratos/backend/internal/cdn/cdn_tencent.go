package cdn

// 腾讯云 CDN adapter（§五 5.4）：JSON POST（X-TC-* 头）+ TC3-HMAC-SHA256 签名。
// 批量 ≤1000/批，限速 20qps（token bucket），每日 10000 URL 预警（8000 起 warn）。

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const tencentVersion = "2018-06-06"

type tencentProvider struct {
	endpoint  string
	secretID  string
	secretKey string
	batchSize int
	bucket    *tokenBucket
	counter   *dailyCounter
}

// NewTencent factory（统一签名 §3.3）：缺 secret_id/secret_key 返回 error。
func NewTencent(cfg map[string]any) (Provider, error) {
	id := cfgString(cfg, "secret_id")
	key := cfgString(cfg, "secret_key")
	if id == "" || key == "" {
		return nil, fmt.Errorf("tencent: secret_id/secret_key required")
	}
	endpoint := "https://cdn.tencentcloudapi.com/"
	if v := cfgString(cfg, "endpoint"); v != "" { // 测试端点旋钮
		endpoint = v
	}
	return &tencentProvider{endpoint: endpoint, secretID: id, secretKey: key,
		batchSize: cfgInt(cfg, "batch_size", 1000),
		bucket:    newTokenBucket(cfgFloat(cfg, "rate_limit_qps", 20)),
		counter:   newDailyCounter(8000, warnLog)}, nil
}

func (p *tencentProvider) Name() string { return "tencent" }

func (p *tencentProvider) Purge(ctx context.Context, keys []string) error {
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

func (p *tencentProvider) purgeBatch(ctx context.Context, keys []string) error {
	body, _ := json.Marshal(map[string]any{"Urls": keys})
	u, err := url.Parse(p.endpoint)
	if err != nil {
		return err
	}
	host := u.Host
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	ts := time.Now().Unix()
	date := time.Now().UTC().Format("2006-01-02")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-TC-Action", "PurgeUrlsCache")
	req.Header.Set("X-TC-Version", tencentVersion)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(ts, 10))
	req.Header.Set("Authorization", tencentSign(p.secretID, p.secretKey, date, ts, body, host))
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

// tencentSign TC3-HMAC-SHA256（§五 5.4）：
// canonicalRequest = "POST\n/\n\ncontent-type;host;x-tc-action\n" + sha256Hex(body)
// credentialScope = "{date}/cdn/tc3_request"
// 4 步密钥派生：kDate=HMAC(secret,date) → kService=HMAC(kDate,"cdn")
//              → kSigning=HMAC(kService,"tc3_request") → Signature=HMAC(kSigning,strToSign)
func tencentSign(secretID, secretKey, date string, ts int64, body []byte, host string) string {
	canonicalRequest := strings.Join([]string{
		"POST", "/", "",
		"content-type;host;x-tc-action",
		sha256Hex(body),
	}, "\n")
	scope := date + "/cdn/tc3_request"
	strToSign := strings.Join([]string{
		"TC3-HMAC-SHA256",
		strconv.FormatInt(ts, 10),
		scope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	kDate := hmacSHA256([]byte(secretKey), date)
	kService := hmacSHA256(kDate, "cdn")
	kSigning := hmacSHA256(kService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, strToSign))
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=content-type;host;x-tc-action, Signature=%s",
		secretID, scope, signature)
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key []byte, msg string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))
	return mac.Sum(nil)
}
