package cdn

// CloudFront adapter（§五 5.2）：官方 aws-sdk-go-v2（仅 aws/credentials/cloudfront 三模块）。
// IAM SigV4 由 SDK 处理；CreateInvalidation 批量 ≤3000/批；OAC 回源 + cloudfront:CreateInvalidation 为厂商侧要求。

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	cf "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

const cfBatchDefault = 3000

type cloudFrontProvider struct {
	client    *cf.Client
	distID    string
	batchSize int
}

// NewCloudFront factory（统一签名 §3.3）：缺三必填键返回 error；Region 固定 us-east-1
// （CreateInvalidation 无区域概念，签名区域仅占位）。
func NewCloudFront(cfg map[string]any) (Provider, error) {
	ak := cfgString(cfg, "access_key_id")
	sk := cfgString(cfg, "secret_access_key")
	distID := cfgString(cfg, "distribution_id")
	if ak == "" || sk == "" || distID == "" {
		return nil, fmt.Errorf("cloudfront: access_key_id/secret_access_key/distribution_id required")
	}
	awsCfg := aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider(ak, sk, ""),
	}
	opts := []func(*cf.Options){}
	if v := cfgString(cfg, "base_endpoint"); v != "" { // 测试端点旋钮（httptest mock）
		opts = append(opts, func(o *cf.Options) { o.BaseEndpoint = aws.String(v) })
	}
	return &cloudFrontProvider{client: cf.NewFromConfig(awsCfg, opts...),
		distID: distID, batchSize: cfgInt(cfg, "batch_size", cfBatchDefault)}, nil
}

func (p *cloudFrontProvider) Name() string { return "cloudfront" }

func (p *cloudFrontProvider) Purge(ctx context.Context, keys []string) error {
	paths := make([]string, 0, len(keys))
	for _, k := range keys {
		paths = append(paths, cfPath(k))
	}
	for _, batch := range Split(paths, p.batchSize) {
		if err := p.purgeBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// cfPath 去 query + 补前导 /（invalidation path 不含 query，官方要求）。
func cfPath(key string) string {
	if i := strings.IndexByte(key, '?'); i >= 0 {
		key = key[:i]
	}
	return "/" + strings.TrimLeft(key, "/")
}

// cfCallerRef 批次内 key 排序后 sha256 前 16 hex：同批幂等（重试不重复建 invalidation）。
func cfCallerRef(paths []string) string {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	h := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	return hex.EncodeToString(h[:8])
}

func (p *cloudFrontProvider) purgeBatch(ctx context.Context, paths []string) error {
	_, err := p.client.CreateInvalidation(ctx, &cf.CreateInvalidationInput{
		DistributionId: aws.String(p.distID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(cfCallerRef(paths)),
			Paths: &types.Paths{
				Quantity: aws.Int32(int32(len(paths))),
				Items:    paths,
			},
		},
	})
	// SDK 自带 429/5xx 重试；其余错误记日志由 manager 兜底
	return err
}
