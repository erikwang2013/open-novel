# CDN 多厂商接入实施计划（2026-08-29-cdn-support）

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development

## Goal

将现有「章节 CDN 静态化」（单一 CDN_PURGE_URL webhook）升级为多厂商 CDN 失效编排：
Cloudflare / AWS CloudFront（海外线）+ 阿里云 CDN / 腾讯云 CDN（中国线），配置入 DB（`novel_cdn_provider` 表，AES-GCM 加密 config 列，复用 PAYMENT_ENCRYPT_KEY），管理端可配置管理（增删改启停，热生效无需重启），保留 generic 兼容路径直至灰度退役。设计文档：`/home/wwwroot/open-novel/docs/cdn-support.md`。本轮范围 = 设计文档 §三~§八 1:1；§1.3 非目标与 §10 升级路径一律不做。

## Architecture

- `internal/cdn` 新包（纯 HTTP，零 DB/conf 依赖）：`Provider` 接口 + `Manager` 编排（去重、并发广播、单厂商失败仅记日志、429/5xx 单次重试 1s 退避）+ 5 个 adapter（cloudflare/cloudfront/aliyun/tencent/generic）+ tokenBucket 限速 + dailyCounter 每日限额预警 + 批量切分工具。
- `biz/cdn.go` 改造为薄门面：包级 `cdnRegistry{mu, init, db, cr, finger, manager}`，`InitCdn(d, cr)` 启动时从 DB 读启用厂商行构造默认 Manager；每次 purge 前按「DB 全行指纹（sha256）」检测变更，变更即重建（管理端操作不重启生效）。镜像 payment.go 的 providerFactories / decryptConfig / encryptConfig 模式，不引入新抽象。
- 管理端 API `api/cdn/v1/cdn.proto`（5 个 RPC，路由镜像 payment 的 `/api/payments/admin/providers*`）→ `biz/cdn_admin.go`（CRUD + 键名校验 §3.3 表 + encryptConfig 落库 + WriteAudit 仅键名）→ `service/cdn_admin.go`（requireAdmin）。
- 数据流：管理端 UI 明文表单 → 后端 `encryptConfig` 加密落库 → `buildCdnManager` 解密构造 adapter → 章节变更触发 `PurgeChaptersAsync(id, langs)` 单次多 key 广播全部启用厂商。区域分流在 DNS 层（后端零改动，设计 §二）。

## Tech Stack

| 组件 | 版本/约束 |
| :--- | :--- |
| Go | 1.25.0（go.mod 已定，不升级） |
| Kratos | **锁定 v2.9.2**（go.mod `replace` 指向 ..，绝不改动） |
| 新增依赖 | `github.com/aws/aws-sdk-go-v2` + `credentials` + `service/cloudfront`（仅这 3 个模块进 go.mod；`config` 模块不引用，`go mod tidy` 会自动剔除） |
| MySQL | 已接 gorm（dbresolver FORCE_MASTER 模式沿用） |
| Redis | 仅测试环境用（127.0.0.1:6380，与现有测试一致） |
| 测试 | 仅 stdlib（t.Fatalf / t.Context / httptest），不引 testify |
| proto | protoc v5.28.3 + protoc-gen-go-http（kratos 官方插件）+ protoc-gen-go-grpc v1.6.2（与现有生成文件头版本一致，GOBIN=/home/wwwroot/go/bin） |
| Flutter | 管理端 apps/admin（dio + records 风格页面，镜像 providers_page.dart） |

## 范围对照（设计文档 §三~§八）

| 设计章节 | 落地任务 |
| :--- | :--- |
| §三 Provider 抽象层（3.1 包位置 / 3.2 接口 / 3.3 统一配置结构+DDL+工厂表） | Task 1（DDL+模型）、Task 2（internal/cdn 核心）、Task 8（工厂表+加载）、Task 10（管理端 CRUD） |
| §四 现有代码改造（4.1 biz/cdn.go 演进 / 4.2 缓存头扩展 / 4.3 测试兼容） | Task 8（门面+PathCachePolicy+BookKey）、Task 9（SetChapterStatus 合批）、Task 2-7（兼容性守卫） |
| §五 厂商适配器（5.1~5.5） | Task 3（Cloudflare）、Task 4（Aliyun）、Task 5（Tencent）、Task 6（CloudFront）、Task 7（Generic） |
| §六 配置设计（DB + 管理端 + env 退役表） | Task 1、Task 8、Task 10（管理端操作=§8.4「灰度期 SQL 直插」的替代，本轮即交付管理端，不再 SQL 直插） |
| §七 测试策略 | 各任务 TDD 测试 + Task 12 全量验证 |
| §八 上线检查清单 | Task 12 输出 checklist 引用；§8.4 录入动作由 Task 10 管理端承担 |

明确不做（§1.3 / §10）：VIP 边缘缓存/签名 URL（URLSigner 接口位仅声明不实现）、按 Region 返回 CDN base URL、动态 GET 列表静态化、静态资源加速、可靠失效队列、每日限额指标上报、generic 退役（灰度完成后按 §九决议 4 另立任务）。

## 全局约定

- 所有命令工作目录：`/home/wwwroot/open-novel/kratos/backend`（除非另注）；测试数据库 DSN：`root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local`，Redis `127.0.0.1:6380`，OpenSearch `http://127.0.0.1:9200`（与现有 `biz/cdn_test.go` 一致）。
- 单测文件与实现同包（`package cdn` / `package biz`），直接测内部符号。
- 行数约束：新增文件 ≤500 行；注释从简。
- commit 风格（git log 惯例）：`feat: <中文概要（关键点）>` / `test: ...` / `docs: ...`；每次 commit 只含本任务文件。
- 厂商工厂签名统一（自检项）：`func(cfg map[string]any) (cdn.Provider, error)`，biz 内 `cdnProviderFactories` map 含全部 5 个 code（含 generic）。
- 配置值兼容两种 JSON 形态：SQL/API 录入的字符串（"30"）与既有代码可能的 float64（30）——adapter 一律经 `cfgInt/cfgFloat` 解析。

---

## Task 0 依赖：aws-sdk-go-v2 三模块入 go.mod

**Files**：`go.mod`（仅此文件，Kratos replace 行不动）

步骤：

1. 拉取依赖（版本固定；若某模块版本解析失败，将该模块改为 `@latest` 重试，语义一致）：
```bash
cd /home/wwwroot/open-novel/kratos/backend
go get github.com/aws/aws-sdk-go-v2@v1.36.3 github.com/aws/aws-sdk-go-v2/credentials@v1.17.62 github.com/aws/aws-sdk-go-v2/service/cloudfront@v1.45.5
go mod tidy
```
2. 校验：`go mod tidy` 后 go.mod 中 **不出现** `aws-sdk-go-v2/config`（无引用被剔除——设计 §九决议 1 注释），`replace github.com/go-kratos/kratos/v2 => ..` 原样保留。
3. `go build ./...` 通过。
4. commit：`feat: 引入 aws-sdk-go-v2 三模块（CloudFront 失效认证，kratos v2.9.2 不动）`

---

## Task 1 DDL + 模型 + 模型/加密往返单测

**Files**：
- Modify `/home/wwwroot/open-novel/kratos/backend/sql/init.sql`（文件末尾追加，novel_audit_log 之后）
- Modify `/home/wwwroot/open-novel/kratos/backend/internal/data/models.go`（PaymentProvider 之后插入）
- Create `/home/wwwroot/open-novel/kratos/backend/internal/biz/cdn_db_test.go`

步骤 1（写失败测试）：`internal/biz/cdn_db_test.go` —— 定义供后续所有 DB 测试复用的建表常量与辅助函数，测试「模型读写 + encryptConfig/decryptConfig 往返」：

```go
package biz

// CDN DB 层测试：novel_cdn_provider 模型读写 + 配置加密往返（§三 3.3 密钥复用决策的守卫）。
// 供 cdn_admin_test.go / 热更新验收测试复用 cdnTestDDL / newCdnTestData。

import (
	"context"
	"testing"

	"open-novel/backend/internal/conf"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// cdnTestDDL 测试库幂等建表（与 init.sql 追加段一致，缺列自行补齐）。
const cdnTestDDL = `CREATE TABLE IF NOT EXISTS novel_cdn_provider (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code       VARCHAR(32)    NOT NULL,
  enabled    TINYINT        NOT NULL DEFAULT 0,
  sort       INT            NOT NULL DEFAULT 0,
  config     TEXT           NULL,
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_code (code),
  KEY idx_enabled_sort (enabled, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

const cdnTestDSN = "root:novel_dev_2026@tcp(127.0.0.1:3307)/novel?charset=utf8mb4&parseTime=True&loc=Local"

func newCdnTestData(t *testing.T) *data.Data {
	t.Helper()
	d, err := data.NewData(&conf.Data{DbDsn: cdnTestDSN, RedisAddr: "127.0.0.1:6380",
		OpensearchAddr: "http://127.0.0.1:9200"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Cache.Close() })
	if err := d.DB.Exec(cdnTestDDL).Error; err != nil {
		t.Fatal(err)
	}
	return d
}

func TestCdnProviderModelRoundTrip(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })

	r := data.CdnProvider{Code: "cloudflare", Enabled: 1, Sort: 1,
		Config: "enc-abc"}
	if err := d.DB.Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	var got data.CdnProvider
	if err := d.DB.First(&got, r.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Code != "cloudflare" || got.Enabled != 1 || got.Sort != 1 || got.Config != "enc-abc" {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

// TestCdnConfigEncryptRoundTrip 加密往返：同一密钥下明文配置加密→解密还原。
func TestCdnConfigEncryptRoundTrip(t *testing.T) {
	cr, err := pkg.NewCrypto("dev-encrypt-key-change-me")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := encryptConfig(map[string]string{"zone_id": "z1", "api_token": "t1", "batch_size": "30"})
	if err != nil {
		t.Fatal(err)
	}
	if enc == "" {
		t.Fatal("encryptConfig must not return empty for non-empty config")
	}
	cfg, err := decryptConfig(enc, cr)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["zone_id"] != "z1" || cfg["batch_size"] != "30" {
		t.Fatalf("decrypt mismatch: %+v", cfg)
	}
	_ = context.Background() // 保持 import（后续 DB 用例扩展预留）
}
```

步骤 2（跑失败）：`go test ./internal/biz/... -run TestCdnProviderModelRoundTrip` → 编译失败（模型/函数不存在）。

步骤 3（最小实现）：

3a. `sql/init.sql` 末尾追加（§3.3 DDL 原文）：

```sql
-- ------------------------------------------------------------
-- CDN 厂商表：多厂商 CDN 失效配置（config 列 AES-GCM 加密 JSON）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS novel_cdn_provider (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  code       VARCHAR(32)    NOT NULL COMMENT 'cloudflare/cloudfront/aliyun/tencent',
  enabled    TINYINT        NOT NULL DEFAULT 0 COMMENT '0禁用 1启用',
  sort       INT            NOT NULL DEFAULT 0 COMMENT '广播顺序（升序）',
  config     TEXT           NULL COMMENT 'AES-GCM 加密 JSON（凭据/批量/限速）',
  created_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_code (code),
  KEY idx_enabled_sort (enabled, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CDN 厂商表';
```

3b. `internal/data/models.go` 在 `PaymentProvider`（:242 `func (PaymentProvider) TableName()...` 之后）插入：

```go
// CDN 厂商模型（多厂商接入 §3.3）：config 列 AES-GCM 加密 JSON，镜像 PaymentProvider（无 lang/region）。
type CdnProvider struct {
	ID        uint64    `gorm:"primaryKey;column:id"`
	Code      string    `gorm:"column:code"`
	Enabled   int8      `gorm:"column:enabled"`
	Sort      int       `gorm:"column:sort"`
	Config    string    `gorm:"column:config"` // AES-GCM 加密 JSON
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (CdnProvider) TableName() string { return "novel_cdn_provider" }
```

3c. `internal/biz/cdn.go` 追加两个包级加密工具（镜像 payment.go:918/:357，Task 8 会整体重写该文件，此处先补函数；`TestCdnConfigEncryptRoundTrip` 依赖）：

```go
// encryptConfig 明文配置 JSON 加密；空配置返回空串（未配置）。镜像 payment.go encryptConfig。
func encryptConfig(cfg map[string]string) (string, error) {
	if len(cfg) == 0 {
		return "", nil
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return crEncrypt(b) // Task 8 改为接收 *pkg.Crypto 的签名，此实现先内联
}
```

注意：Task 8 重写时 `encryptConfig`/`decryptConfig` 最终签名固定为：

```go
func encryptConfig(cfg map[string]string) (string, error) // 包级，无需 cr（测试/管理端统一走 biz.NewCdnAdminUsecase 注入的 cr 解密）
```

（见 Task 8 完整实现，勿在 Task 1 提前发散。）

步骤 4（跑过）：`go test ./internal/biz/... -run 'TestCdnProviderModelRoundTrip|TestCdnConfigEncryptRoundTrip'` → 全过。

步骤 5：commit：`feat: novel_cdn_provider 表与模型 + 配置加密往返单测（复用 PAYMENT_ENCRYPT_KEY）`

> 说明：Task 1 步骤 3c 仅为让加密往返测试可编译的最小实现；Task 8 重写 cdn.go 时以 Task 8 代码为准，函数签名可能微调（测试同步微调，测试目标不变）。

---

## Task 2 internal/cdn 核心：Provider / Manager / 工具

**Files**：Create `/home/wwwroot/open-novel/kratos/backend/internal/cdn/cdn.go`、Create `internal/cdn/cdn_test.go`

步骤 1（失败测试）`internal/cdn/cdn_test.go`：

```go
package cdn

// 核心编排测试：去重 / 批量切分 / 并发广播 / 429 重试一次 / 非重试错误不重试 / 令牌桶 / 每日计数。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type fakeProvider struct {
	name string
	hits *int32
	err  error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Purge(ctx context.Context, keys []string) error {
	atomic.AddInt32(f.hits, 1)
	return f.err
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"a", "b", "a", "c", "b"})
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("want [a b c], got %v", got)
	}
}

func TestSplit(t *testing.T) {
	got := Split([]string{"1", "2", "3", "4", "5"}, 2)
	if len(got) != 3 || len(got[0]) != 2 || len(got[2]) != 1 {
		t.Fatalf("split mismatch: %v", got)
	}
	if got := Split(nil, 0); got != nil {
		t.Fatalf("max<=0 must not panic, got %v", got)
	}
}

// TestManagerBroadcast 两个 provider 各收到同一批 key。
func TestManagerBroadcast(t *testing.T) {
	var h1, h2 int32
	m := NewManager([]Provider{&fakeProvider{name: "p1", hits: &h1}, &fakeProvider{name: "p2", hits: &h2}})
	m.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/1?lang=zh-CN", "chapter/1?lang=en"})
	if atomic.LoadInt32(&h1) != 1 || atomic.LoadInt32(&h2) != 1 {
		t.Fatalf("want 1 purge call per provider, got %d/%d", h1, h2)
	}
}

// TestManagerRetryOnce 429 重试一次（1s 退避），第二次成功不再发。
func TestManagerRetryOnce(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	p := &fakeProvider{name: "t", err: &retriableError{status: 429}}
	_ = p
	// 直接走真实 HTTP 路径：用 generic provider 打 mock（generic 适配器 Task 7 才实现，
	// 此处以 manager 级重试语义为准：重试仅针对 retriableError）。
	m := NewManager([]Provider{p})
	start := time.Now()
	m.Purge(t.Context(), []string{"k"})
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("fake provider err path: %d", calls)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Fatal("retry backoff must sleep ~1s")
	}
	// 第二次调用：provider 不再报错 → 无重试无 sleep
	ok := &fakeProvider{name: "ok", hits: new(int32)}
	m2 := NewManager([]Provider{ok})
	m2.Purge(t.Context(), []string{"k"})
	if atomic.LoadInt32(ok.hits) != 1 {
		t.Fatalf("want 1 hit, got %d", atomic.LoadInt32(ok.hits))
	}
}

// TestManagerNonRetriableNoRetry 非重试错误只记日志，不 sleep。
func TestManagerNonRetriableNoRetry(t *testing.T) {
	p := &fakeProvider{name: "x", err: &retriableError{status: 400}, hits: new(int32)}
	m := NewManager([]Provider{p})
	start := time.Now()
	m.Purge(t.Context(), []string{"k"})
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("non-retriable must not sleep")
	}
	if atomic.LoadInt32(p.hits) != 1 {
		t.Fatalf("want 1 call, got %d", atomic.LoadInt32(p.hits))
	}
}

func TestTokenBucketWait(t *testing.T) {
	b := newTokenBucket(100) // 100 qps
	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := b.Wait(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("10 tokens at 100qps must be fast, took %v", time.Since(start))
	}
}

func TestDailyCounterWarnOnce(t *testing.T) {
	var warns int
	c := newDailyCounter(8000, func(string) { warns++ })
	c.Add(7999)
	if warns != 0 {
		t.Fatal("must not warn below threshold")
	}
	c.Add(1)
	if warns != 1 {
		t.Fatal("must warn at threshold")
	}
	c.Add(100)
	if warns != 1 {
		t.Fatal("must warn only once per day")
	}
}
```
步骤 2（跑失败）：`go test ./internal/cdn/...` → 编译失败（包不存在）。

步骤 3（最小实现）`internal/cdn/cdn.go`：

```go
package cdn

// CDN 失效编排（设计 §三）：Provider 接口 + Manager 并发广播 + 批量/限速/每日限额工具。
// 纯 HTTP 标准库，不依赖 DB/conf，可独立测试。

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Provider 厂商适配器：一次调用清理一批缓存对象。
type Provider interface {
	Name() string
	Purge(ctx context.Context, keys []string) error // keys 为全量对象 key（如 chapter/123?lang=zh-CN）
}

// URLSigner 签名 URL 扩展位：本轮不实现（设计 §3.2 预留，VIP 边缘缓存时再落地）。
type URLSigner interface {
	SignURL(url string, expire time.Duration) (string, error)
}

// Manager 编排：由启用厂商列表构造，Purge 时并发广播全部。
type Manager struct{ providers []Provider }

// NewManager 空列表 → 空 manager（全禁用，行为等同现状未启用）。
func NewManager(providers []Provider) *Manager { return &Manager{providers: providers} }

// Purge 并发广播；每厂商失败仅记日志（best-effort，§4.1 ponytail 语义不变）。
func (m *Manager) Purge(ctx context.Context, keys []string) {
	if len(m.providers) == 0 || len(keys) == 0 {
		return
	}
	keys = dedupe(keys)
	var wg sync.WaitGroup
	for _, p := range m.providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			m.purgeOne(ctx, p, keys)
		}(p)
	}
	wg.Wait()
}

// purgeOne 单厂商发送：429/5xx 重试一次（1s 退避），仍失败记 warn。
// ponytail: 固定 1s 退避 + 单次重试；purge 丢失最坏多缓存 1h（s-maxage 到期），可靠失效需队列，暂不做。
func (m *Manager) purgeOne(ctx context.Context, p Provider, keys []string) {
	for attempt := 0; attempt < 2; attempt++ {
		err := p.Purge(ctx, keys)
		if err == nil {
			return
		}
		if attempt == 0 && httpRetriable(err) {
			time.Sleep(time.Second)
			continue
		}
		log.Printf("[cdn] %s purge failed: %v", p.Name(), err)
		return
	}
}

// retriableError 可重试的 HTTP 错误（429/5xx）。
type retriableError struct{ status int }

func (e *retriableError) Error() string { return fmt.Sprintf("http %d", e.status) }

func httpRetriable(err error) bool {
	var re *retriableError
	return errors.As(err, &re) && (re.status == http.StatusTooManyRequests || re.status >= 500)
}

// dedupe 去重保序。
func dedupe(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	return out
}

// Split 按 max 切批；max<=0 时按 1000。
func Split(keys []string, max int) [][]string {
	if max <= 0 {
		max = 1000
	}
	var out [][]string
	for i := 0; i < len(keys); i += max {
		end := min(i+max, len(keys))
		out = append(out, keys[i:end])
	}
	return out
}

// httpClient 厂商 API 共用客户端：5s 超时（沿用现状）。
var httpClient = &http.Client{Timeout: 5 * time.Second}

// cfgString / cfgInt / cfgFloat 配置读取：兼容字符串与 JSON 数值两种形态（§全局约定）。
func cfgString(cfg map[string]any, key string) string {
	s, _ := cfg[key].(string)
	return s
}

func cfgInt(cfg map[string]any, key string, def int) int {
	switch v := cfg[key].(type) {
	case float64:
		return int(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func cfgFloat(cfg map[string]any, key string, def float64) float64 {
	switch v := cfg[key].(type) {
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// tokenBucket 简单令牌桶限速（构造时给定 qps）。
type tokenBucket struct {
	mu     sync.Mutex
	qps    float64
	tokens float64
	last   time.Time
}

func newTokenBucket(qps float64) *tokenBucket {
	return &tokenBucket{qps: qps, tokens: qps, last: time.Now()}
}

func (b *tokenBucket) Take() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	b.tokens = min(b.qps, b.tokens+now.Sub(b.last).Seconds()*b.qps)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Wait 阻塞直到取得令牌或 ctx 结束。
func (b *tokenBucket) Wait(ctx context.Context) error {
	for {
		if b.Take() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// dailyCounter 每日计数：达 warnAt 记一次 warn（按天重置，§5.3 阿里/腾讯 10000 限额预警）。
type dailyCounter struct {
	mu     sync.Mutex
	day    string
	count  int
	warned bool
	warnAt int
	warn   func(string)
}

func newDailyCounter(warnAt int, warn func(string)) *dailyCounter {
	return &dailyCounter{warnAt: warnAt, warn: warn}
}

// Add 累计 n；返回当天累计值；首次越过 warnAt 触发一次 warn。
func (c *dailyCounter) Add(n int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != c.day {
		c.day, c.count, c.warned = today, 0, false
	}
	c.count += n
	if !c.warned && c.count >= c.warnAt {
		c.warned = true
		if c.warn != nil {
			c.warn(fmt.Sprintf("cdn daily purge count %d >= %d, check manually", c.count, c.warnAt))
		}
	}
	return c.count
}

// non2xx 统一判定：429/5xx → retriableError（manager 重试一次）；其余 → 普通错误（仅记日志）。
func non2xx(status int) error {
	if status == http.StatusTooManyRequests || status >= 500 {
		return &retriableError{status: status}
	}
	return fmt.Errorf("http %d", status)
}

// warnLog 包装 log.Printf 适配 dailyCounter 回调。
func warnLog(s string) { log.Printf("[cdn] %s", s) }
```

步骤 4（跑过）：`go test ./internal/cdn/...` → 全过（含 TestManagerRetryOnce 的 ~1s 退避断言）。

步骤 5：commit：`feat: internal/cdn 编排核心（Manager 广播/去重/切批/重试/限速/每日限额）`

---

## Task 3 Cloudflare 适配器

**Files**：Create `internal/cdn/cdn_cloudflare.go`、Create `internal/cdn/cdn_cloudflare_test.go`

步骤 1（失败测试）`internal/cdn/cdn_cloudflare_test.go`：

```go
package cdn

// Cloudflare 适配器（§五 5.1）：Bearer + JSON files ≤30/批；200 且 success=true 才算成功；
// 429/5xx 可重试；缺必填键构造报错。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewCloudflareValidate(t *testing.T) {
	if _, err := NewCloudflare(map[string]any{}); err == nil {
		t.Fatal("empty cfg must error")
	}
	if _, err := NewCloudflare(map[string]any{"zone_id": "z"}); err == nil {
		t.Fatal("missing api_token must error")
	}
}

func TestCloudflarePurge(t *testing.T) {
	var got struct {
		Method string
		Auth   string
		Path   string
		Files  []string
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Auth = r.Header.Get("Authorization")
		got.Path = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got.Files)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer ts.Close()

	p, err := NewCloudflare(map[string]any{"zone_id": "zone1", "api_token": "tok", "base_url": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "cloudflare" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost || !strings.HasSuffix(got.Path, "/zones/zone1/purge_cache") {
		t.Fatalf("request mismatch: %+v", got)
	}
	if got.Auth != "Bearer tok" {
		t.Fatalf("auth: %s", got.Auth)
	}
	if len(got.Files) != 2 || got.Files[0] != "chapter/1?lang=zh-CN" {
		t.Fatalf("files: %v", got.Files)
	}
}

// TestCloudflareBatch 超过 30 个 key 切多批（mock 计数）。
func TestCloudflareBatch(t *testing.T) {
	var batches int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches++
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer ts.Close()
	p, err := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL, "batch_size": float64(10)})
	if err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 25)
	for i := range keys {
		keys[i] = "k" + strings.Repeat("x", i)
	}
	if err := p.Purge(t.Context(), keys); err != nil {
		t.Fatal(err)
	}
	if batches != 3 {
		t.Fatalf("want 3 batches, got %d", batches)
	}
}

func TestCloudflareNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	p, _ := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL})
	if err := p.Purge(t.Context(), []string{"k"}); err == nil {
		t.Fatal("401 must error")
	}
}

// TestCloudflareRetriable 429 → 错误标记可重试（manager 会重试一次）。
func TestCloudflareRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()
	p, _ := NewCloudflare(map[string]any{"zone_id": "z", "api_token": "t", "base_url": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}
```

步骤 2（跑失败）：`go test ./internal/cdn/... -run TestCloudflare` → 编译失败。

步骤 3（最小实现）`internal/cdn/cdn_cloudflare.go`：

```go
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
```

步骤 4（跑过）：`go test ./internal/cdn/... -run 'TestNewCloudflare|TestCloudflare'` → 全过。

步骤 5：commit：`feat: Cloudflare adapter（Bearer purge_cache，≤30/批，success 判定）`

---

## Task 4 阿里云 CDN 适配器（RPC 签名 + 向量守卫）

**Files**：Create `internal/cdn/cdn_aliyun.go`、Create `internal/cdn/cdn_aliyun_test.go`

步骤 1（失败测试）`internal/cdn/cdn_aliyun_test.go`：

```go
package cdn

// 阿里云 adapter（§五 5.3）：RPC form + HMAC-SHA1 签名；ObjectPath 换行分隔；≤1000/批；50qps；
// 每日 10000 URL 预警（8000 起 warn）。签名向量由独立参考实现（Python）计算，防回归。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestAliyunSignVector 固定输入 → 固定签名（独立参考实现计算）：
// AccessKeyId=testid, Action=DescribeCdnService, Format=JSON,
// SignatureMethod=HMAC-SHA1, SignatureNonce=4c33d64d-ee48-4b13-bb81-0a3b0a4f5b7b,
// SignatureVersion=1.0, Timestamp=2016-02-23T12:46:24Z, Version=2014-11-11,
// SecretKey=testsecret。
func TestAliyunSignVector(t *testing.T) {
	params := map[string]string{
		"AccessKeyId":      "testid",
		"Action":           "DescribeCdnService",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   "4c33d64d-ee48-4b13-bb81-0a3b0a4f5b7b",
		"SignatureVersion": "1.0",
		"Timestamp":        "2016-02-23T12:46:24Z",
		"Version":          "2014-11-11",
	}
	got := aliyunSign("testsecret", params)
	want := "1aHx5fy2R2UfxfBMfMvIv624x50="
	if got != want {
		t.Fatalf("signature vector mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestNewAliyunValidate(t *testing.T) {
	if _, err := NewAliyun(map[string]any{"access_key_id": "a"}); err == nil {
		t.Fatal("missing access_key_secret must error")
	}
}

// TestAliyunPurgeForm mock 断言：POST 表单含 ObjectPath（\n 分隔）与签名参数。
func TestAliyunPurgeForm(t *testing.T) {
	var got url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got = r.PostForm
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	p, err := NewAliyun(map[string]any{"access_key_id": "ak", "access_key_secret": "sk",
		"endpoint": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "aliyun" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Get("Action") != "RefreshObjectCaches" || got.Get("ObjectType") != "File" {
		t.Fatalf("action mismatch: %v", got)
	}
	if got.Get("ObjectPath") != "chapter/1?lang=zh-CN\nchapter/2?lang=en" {
		t.Fatalf("object path: %q", got.Get("ObjectPath"))
	}
	if got.Get("Signature") == "" || got.Get("Timestamp") == "" || got.Get("SignatureNonce") == "" {
		t.Fatalf("missing signature params: %v", got)
	}
}

// TestAliyunBatch1000 超 1000 个 key 切多批。
func TestAliyunBatch1000(t *testing.T) {
	var batches int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	p, _ := NewAliyun(map[string]any{"access_key_id": "a", "access_key_secret": "s",
		"endpoint": ts.URL, "batch_size": float64(100)})
	keys := make([]string, 250)
	for i := range keys {
		keys[i] = "k" + strings.Repeat("y", i)
	}
	if err := p.Purge(t.Context(), keys); err != nil {
		t.Fatal(err)
	}
	if batches != 3 {
		t.Fatalf("want 3 batches, got %d", batches)
	}
}

// TestAliyunRetriable 5xx → 可重试错误。
func TestAliyunRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()
	p, _ := NewAliyun(map[string]any{"access_key_id": "a", "access_key_secret": "s", "endpoint": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}

// TestAliyunDailyCounter 日累计计数到达阈值记 warn（当天仅一次）。
func TestAliyunDailyCounter(t *testing.T) {
	var warns int
	c := newDailyCounter(8000, func(string) { warns++ })
	p := &aliyunProvider{counter: c}
	p.counter.Add(7990)
	p.counter.Add(10)
	if warns != 1 {
		t.Fatalf("want 1 warn, got %d", warns)
	}
	p.counter.Add(5)
	if warns != 1 {
		t.Fatalf("want still 1 warn, got %d", warns)
	}
	_ = context.Background()
}
```

步骤 2（跑失败）：`go test ./internal/cdn/... -run TestAliyun` → 编译失败。

步骤 3（最小实现）`internal/cdn/cdn_aliyun.go`：

```go
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
```

步骤 4（跑过）：`go test ./internal/cdn/... -run 'TestAliyun|TestNewAliyun'` → 全过（含签名向量）。

步骤 5：commit：`feat: 阿里云 CDN adapter（RPC 签名 + 签名向量单测 + 限速/每日限额）`
---

## Task 5 腾讯云 CDN 适配器（TC3 签名 + 向量守卫）

**Files**：Create `internal/cdn/cdn_tencent.go`、Create `internal/cdn/cdn_tencent_test.go`

步骤 1（失败测试）`internal/cdn/cdn_tencent_test.go`：

```go
package cdn

// 腾讯云 adapter（§五 5.4）：JSON POST + TC3-HMAC-SHA256；≤1000/批；20qps；每日 10000 预警。
// 签名向量由独立参考实现（Python）计算，防回归。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// TestTencentSignVector 固定输入 → 固定签名（独立参考实现计算）：
// SecretId=AKIDEXAMPLE, SecretKey=Gu5t9xGARNpq86cd98joQYCN3Cozk1qA,
// ts=1788004800（=2026-08-29 UTC）, host=cdn.tencentcloudapi.com,
// body={"Urls":["chapter/123?lang=zh-CN"]}。
func TestTencentSignVector(t *testing.T) {
	body := []byte(`{"Urls":["chapter/123?lang=zh-CN"]}`)
	got := tencentSign("AKIDEXAMPLE", "Gu5t9xGARNpq86cd98joQYCN3Cozk1qA",
		"2026-08-29", 1788004800, body, "cdn.tencentcloudapi.com")
	want := "TC3-HMAC-SHA256 Credential=AKIDEXAMPLE/2026-08-29/cdn/tc3_request, " +
		"SignedHeaders=content-type;host;x-tc-action, " +
		"Signature=f95b1da7d4ff9fa8bf197765b41f091ee58e3a312a5372174f1fe4575f294a09"
	if got != want {
		t.Fatalf("TC3 vector mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestNewTencentValidate(t *testing.T) {
	if _, err := NewTencent(map[string]any{"secret_id": "a"}); err == nil {
		t.Fatal("missing secret_key must error")
	}
}

// TestTencentPurgeHeaders mock 断言：TC3 头 + JSON Urls。
func TestTencentPurgeHeaders(t *testing.T) {
	var got struct {
		Action, Version, Timestamp, Auth, CT string
		Urls                                []string
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Action = r.Header.Get("X-TC-Action")
		got.Version = r.Header.Get("X-TC-Version")
		got.Timestamp = r.Header.Get("X-TC-Timestamp")
		got.Auth = r.Header.Get("Authorization")
		got.CT = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&got.Urls)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	p, err := NewTencent(map[string]any{"secret_id": "sid", "secret_key": "skey",
		"endpoint": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "tencent" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Action != "PurgeUrlsCache" || got.Version != "2018-06-06" {
		t.Fatalf("headers mismatch: %+v", got)
	}
	if got.CT != "application/json; charset=utf-8" {
		t.Fatalf("content-type must match SignedHeaders: %q", got.CT)
	}
	if !strings.HasPrefix(got.Auth, "TC3-HMAC-SHA256 Credential=sid/") ||
		!strings.Contains(got.Auth, "SignedHeaders=content-type;host;x-tc-action") {
		t.Fatalf("auth header: %q", got.Auth)
	}
	if _, err := strconv.ParseInt(got.Timestamp, 10, 64); err != nil {
		t.Fatalf("bad timestamp: %q", got.Timestamp)
	}
	if len(got.Urls) != 2 || got.Urls[0] != "chapter/1?lang=zh-CN" {
		t.Fatalf("urls: %v", got.Urls)
	}
}

// TestTencentRetriable 429 → 可重试。
func TestTencentRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()
	p, _ := NewTencent(map[string]any{"secret_id": "s", "secret_key": "k", "endpoint": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}
```

步骤 2（跑失败）：`go test ./internal/cdn/... -run TestTencent` → 编译失败。

步骤 3（最小实现）`internal/cdn/cdn_tencent.go`：

```go
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

```

步骤 4（跑过）：`go test ./internal/cdn/... -run 'TestTencent|TestNewTencent'` → 全过（含 TC3 向量）。

步骤 5：commit：`feat: 腾讯云 CDN adapter（TC3-HMAC-SHA256 + 签名向量单测 + 限速/每日限额）`

---

## Task 6 CloudFront 适配器（aws-sdk-go-v2）

**Files**：Create `internal/cdn/cdn_cloudfront.go`、Create `internal/cdn/cdn_cloudfront_test.go`

步骤 1（失败测试）`internal/cdn/cdn_cloudfront_test.go`：

```go
package cdn

// CloudFront adapter（§五 5.2）：aws-sdk-go-v2 CreateInvalidation。
// invalidation path 必须去 query + 补前导 /（官方要求，带 query 直接 400）。
// CallerReference = 批次内 key 排序后 sha256 前 16 hex（唯一、幂等）。
// 测试经 BaseEndpoint 指向 mock server，静态假凭据（mock 不验签）。

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCfPath(t *testing.T) {
	cases := map[string]string{
		"chapter/123?lang=zh-CN": "/chapter/123",
		"chapter/123":            "/chapter/123",
		"/book/9":                "/book/9",
	}
	for in, want := range cases {
		if got := cfPath(in); got != want {
			t.Fatalf("cfPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCfCallerRefStable(t *testing.T) {
	a := cfCallerRef([]string{"/b", "/a"})
	b := cfCallerRef([]string{"/a", "/b"})
	if a != b {
		t.Fatalf("caller ref must be order-independent: %s vs %s", a, b)
	}
	if len(a) != 16 {
		t.Fatalf("caller ref must be 16 hex chars, got %d: %s", len(a), a)
	}
}

func TestNewCloudFrontValidate(t *testing.T) {
	if _, err := NewCloudFront(map[string]any{"access_key_id": "a", "secret_access_key": "s"}); err == nil {
		t.Fatal("missing distribution_id must error")
	}
}

// TestCloudFrontPurgeXML mock 断言：POST /distribution/{id}/invalidation，XML body 无 query。
func TestCloudFrontPurgeXML(t *testing.T) {
	got := struct {
		Method, Path string
		Body         string
	}{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Path = r.URL.Path
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		got.Body = string(buf[:n])
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invalidation xmlns="http://cloudfront.amazonaws.com/doc/2020-05-31/">
  <Id>I1</Id><Status>InProgress</Status>
  <CreateTime>2026-08-29T00:00:00.000Z</CreateTime>
  <InvalidationBatch>
    <Paths><Quantity>1</Quantity><Items>/chapter/123</Items></Paths>
    <CallerReference>abcdef1234567890</CallerReference>
  </InvalidationBatch>
</Invalidation>`))
	}))
	defer ts.Close()

	p, err := NewCloudFront(map[string]any{
		"access_key_id": "AK", "secret_access_key": "SK", "distribution_id": "D1",
		"base_endpoint": ts.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "cloudfront" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/123?lang=zh-CN"}); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost || !strings.Contains(got.Path, "/distribution/D1/invalidation") {
		t.Fatalf("request mismatch: %+v", got)
	}
	if strings.Contains(got.Body, "?lang") {
		t.Fatalf("invalidation path must strip query: %s", got.Body)
	}
	if !strings.Contains(got.Body, "/chapter/123") {
		t.Fatalf("body missing path: %s", got.Body)
	}
}
```

步骤 2（跑失败）：`go test ./internal/cdn/... -run 'TestCf|TestCloudFront|TestNewCloudFront'` → 编译失败。

步骤 3（最小实现）`internal/cdn/cdn_cloudfront.go`：

```go
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
```

步骤 4（跑过）：`go test ./internal/cdn/... -run 'TestCf|TestCloudFront|TestNewCloudFront'` → 全过。

步骤 5：commit：`feat: CloudFront adapter（SDK CreateInvalidation + 去 query 路径转换单测）`

---

## Task 7 Generic 适配器（旧 webhook 兼容）

**Files**：Create `internal/cdn/cdn_generic.go`、Create `internal/cdn/cdn_generic_test.go`

步骤 1（失败测试）`internal/cdn/cdn_generic_test.go`：

```go
package cdn

// Generic adapter（§五 5.5）：{key} 模板逐 key 单请求（1 key/请求，§4.3 兼容断言）；
// 仅 DB 无启用厂商且存在 CDN_PURGE_URL 时激活（biz 兜底构造）。

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNewGenericValidate(t *testing.T) {
	if _, err := NewGeneric(map[string]any{"url_template": "http://x/purge"}); err == nil {
		t.Fatal("template without {key} must error")
	}
}

// TestGenericPurgeTemplate 每 key 一个请求，URL 模板 {key} 替换。
func TestGenericPurgeTemplate(t *testing.T) {
	var mu sync.Mutex
	var urls []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		urls = append(urls, r.URL.RequestURI())
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	p, err := NewGeneric(map[string]any{"url_template": ts.URL + "/purge/{key}"})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "generic" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/1?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 || urls[0] != "/purge/chapter/1?lang=zh-CN" || urls[1] != "/purge/chapter/1?lang=en" {
		t.Fatalf("urls: %v", urls)
	}
}

// TestGenericNon2xx 500 → 可重试错误（manager 重试一次后记日志）。
func TestGenericNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	p, _ := NewGeneric(map[string]any{"url_template": ts.URL + "/purge/{key}"})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("5xx must be retriable, got %v", err)
	}
}
```

步骤 2（跑失败）：`go test ./internal/cdn/... -run 'TestGeneric|TestNewGeneric'` → 编译失败。

步骤 3（最小实现）`internal/cdn/cdn_generic.go`：

```go
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
```

步骤 4（跑过）：`go test ./internal/cdn/...` → 全过（核心 + 5 适配器）。

步骤 5：commit：`feat: generic adapter 兼容 CDN_PURGE_URL（{key} 模板，1 key/请求）`
---

## Task 8 biz/cdn.go 门面：DB 加载 + 指纹热更新 + 缓存键/头策略

**Files**：
- Rewrite `/home/wwwroot/open-novel/kratos/backend/internal/biz/cdn.go`（现行 71 行 → 约 260 行，<500 行）
- Create `/home/wwwroot/open-novel/kratos/backend/internal/biz/cdn_registry_test.go`（DB 三态 + 热更新）

前置：Task 1 已加 `data.CdnProvider`、包级 `encryptConfig/decryptConfig`（本任务以本节代码为准定稿）。

步骤 1（失败测试）`internal/biz/cdn_registry_test.go`：

```go
package biz

// CDN 门面测试（§七）：CdnEnabled 门控 / PathCachePolicy 表 / BookKey /
// InitCdn 加载三态（启用行→厂商、无行+env→generic、全无→空）/ 指纹热更新。

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func TestBookKey(t *testing.T) {
	if got := BookKey(9, "en"); got != "book/9?lang=en" {
		t.Fatalf("BookKey: %q", got)
	}
}

func TestPathCachePolicy(t *testing.T) {
	cases := map[string]string{
		"/api/chapters/123/content":       "public, s-maxage=3600",
		"/api/chapters/123":               "",
		"/api/books/1":                    "",
		"/api/chapters/1/content/v2":      "", // 后缀不匹配 content
		"/api/comments":                   "",
	}
	for path, want := range cases {
		if got := PathCachePolicy(path); got != want {
			t.Fatalf("PathCachePolicy(%q) = %q, want %q", path, got, want)
		}
	}
}

// TestInitCdnStates 三态：启用行 → Manager 含厂商（构造期从行配置生成，不对外请求）；
// 无行 + CDN_PURGE_URL → generic 激活；全无 → 空 Manager（CdnEnabled 走 env 门控）。
func TestInitCdnStates(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")

	// 态 1：启用 cloudflare 行（config 含测试端点，构造成功即证明加载）
	if err := d.DB.Create(&data.CdnProvider{Code: "cloudflare", Enabled: 1, Sort: 1,
		Config: mustEnc(t, map[string]string{"zone_id": "z", "api_token": "t"})}).Error; err != nil {
		t.Fatal(err)
	}
	t.Setenv("CDN_PURGE_URL", "")
	InitCdn(d, cr)
	defer SetDefaultManager(nil) // 还原，避免污染同包其他测试
	if CdnEnabled() != true || cdnActiveNames() == "" {
		t.Fatalf("state1: want enabled with provider, got %q", cdnActiveNames())
	}
	if !strings.Contains(cdnActiveNames(), "cloudflare") {
		t.Fatalf("state1: want cloudflare in manager, got %q", cdnActiveNames())
	}
}

// TestCdnRegistryHotReload 热更新（§6 管理端操作不重启生效的验收点）：
// InitCdn → purge 打到 mock → 禁用（DB UPDATE 等价管理端启停）→ 不再打 → 再启用 → 恢复。
func TestCdnRegistryHotReload(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")

	var mu sync.Mutex
	var hits int
	ts := newPurgeMock(t, &mu, &hits)
	t.Setenv("CDN_PURGE_URL", "")

	row := &data.CdnProvider{Code: "cloudflare", Enabled: 1, Sort: 1,
		Config: mustEnc(t, map[string]string{"zone_id": "z", "api_token": "t", "base_url": ts.URL})}
	if err := d.DB.Create(row).Error; err != nil {
		t.Fatal(err)
	}
	InitCdn(d, cr)
	defer SetDefaultManager(nil)

	purge := func() { PurgeChaptersAsync(9, []string{"zh-CN"}) }
	waitHits := func(min int) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			n := hits
			mu.Unlock()
			if n >= min {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("want >= %d hits, got %d", min, hits)
	}

	purge()
	waitHits(1)

	// 管理端禁用（不重启）→ 指纹变化 → 下次 purge 无请求（直接 UPDATE 等价管理端启停）
	if err := d.DB.Model(&data.CdnProvider{}).Where("id = ?", row.ID).
		Update("enabled", 0).Error; err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	purge()
	time.Sleep(300 * time.Millisecond)
	mu.Lock()
	after := hits
	mu.Unlock()
	if after != 1 {
		t.Fatalf("toggle off must stop purging without restart, got %d hits", after)
	}

	// 重新启用 → purge 恢复
	if err := d.DB.Model(&data.CdnProvider{}).Where("id = ?", row.ID).
		Update("enabled", 1).Error; err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	purge()
	waitHits(2)
}

// ---- 测试辅助 ----

func mustEnc(t *testing.T, cfg map[string]string) string {
	t.Helper()
	enc, err := encryptConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return enc
}

func cdnActiveNames() string {
	m := currentManager()
	if m == nil {
		return ""
	}
	var names []string
	for _, p := range m.Providers() {
		names = append(names, p.Name())
	}
	return strings.Join(names, ",")
}

func newPurgeMock(t *testing.T, mu *sync.Mutex, hits *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*hits++
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
}
```

步骤 2（跑失败）：`go test ./internal/biz/... -run 'TestBookKey|TestPathCachePolicy|TestInitCdnStates|TestCdnRegistryHotReload'` → 编译失败（currentManager / httptest.Server / PurgeChaptersAsync / BookKey / PathCachePolicy / InitCdn / SetDefaultManager / encryptConfig 等缺失）。

步骤 3（最小实现）：

3a. **Rewrite `internal/biz/cdn.go`**（整体替换现行文件，保留全部既有导出符号 `CdnEnabled / ChapterCacheControl / ChapterKey / PurgeChapterAsync`）：

```go
package biz

// CDN 门面（设计 §三/§四）：DB 配置加载（novel_cdn_provider）+ 默认 Manager 指纹热更新
// （管理端操作不重启生效）+ 缓存键/头策略。厂商协议在 internal/cdn（纯 HTTP），本文件仅编排。
// 灰度期保留 env 路径（CDN_BASE_URL/CDN_PURGE_URL → generic），退役见 §九决议 4。

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"open-novel/backend/internal/cdn"
	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// cdnLog 与 cmd 入口同一 stdout logger；purge 为 best-effort，无需注入依赖。
var cdnLog = log.NewStdLogger(os.Stdout)

// cdnProviderFactories 厂商工厂表（§3.3，镜像 payment.go providerFactories；统一签名）。
var cdnProviderFactories = map[string]func(map[string]any) (cdn.Provider, error){
	"cloudflare": cdn.NewCloudflare,
	"cloudfront": cdn.NewCloudFront,
	"aliyun":     cdn.NewAliyun,
	"tencent":    cdn.NewTencent,
	"generic":    cdn.NewGeneric,
}

// cdnRegistry 默认 Manager 注册表：启动 InitCdn 装载，之后每次 purge 前按 DB 全行指纹
// 检测变更并重建（管理端 CRUD 热生效）。
type cdnRegistry struct {
	mu      sync.Mutex
	init    bool
	db      *gorm.DB
	cr      *pkg.Crypto
	finger  string
	manager *cdn.Manager
}

var cdnReg cdnRegistry

// InitCdn 启动时初始化：读启用厂商行（ORDER BY sort）构造默认 Manager；
// 无启用行且存在 CDN_PURGE_URL → generic 兜底；全无 → 空 Manager（全禁用）。
func InitCdn(d *data.Data, cr *pkg.Crypto) {
	cdnReg.mu.Lock()
	defer cdnReg.mu.Unlock()
	cdnReg.db = d.DB
	cdnReg.cr = cr
	cdnReg.init = true
	cdnReg.finger = cdnFingerprint(d.DB)
	cdnReg.manager = buildCdnManager(d.DB, cr)
}

// SetDefaultManager 测试注入：设置默认 Manager 并回到未初始化态（nil 清空恢复 env-only）。
func SetDefaultManager(m *cdn.Manager) {
	cdnReg.mu.Lock()
	defer cdnReg.mu.Unlock()
	cdnReg.init = false
	cdnReg.manager = m
}

// currentManager 返回当前默认 Manager；已 InitCdn 时先比对 DB 指纹，变更即重建。
func currentManager() *cdn.Manager {
	cdnReg.mu.Lock()
	defer cdnReg.mu.Unlock()
	if cdnReg.init && cdnReg.db != nil {
		if f := cdnFingerprint(cdnReg.db); f != cdnReg.finger {
			cdnReg.finger = f
			cdnReg.manager = buildCdnManager(cdnReg.db, cdnReg.cr)
		}
	}
	if cdnReg.manager != nil {
		return cdnReg.manager
	}
	// 未 InitCdn（单测/灰度）：env generic 保底路径
	if u := os.Getenv("CDN_PURGE_URL"); u != "" {
		if f, ok := cdnProviderFactories["generic"]; ok {
			if p, err := f(map[string]any{"url_template": u}); err == nil {
				return cdn.NewManager([]cdn.Provider{p})
			}
		}
	}
	return nil
}

// cdnFingerprint DB 全行指纹：code|enabled|sort|config 拼接后 sha256（含禁用行与 config，
// 启停/改密均触发重建）；空表时并入 CDN_PURGE_URL，避免 generic 激活状态漂移不被感知。
func cdnFingerprint(db *gorm.DB) string {
	var rows []data.CdnProvider
	db.WithContext(context.Background()).Order("id").Find(&rows)
	var sb strings.Builder
	for _, r := range rows {
		sb.WriteString(r.Code)
		sb.WriteString("|")
		sb.WriteString(strconv.Itoa(int(r.Enabled)))
		sb.WriteString("|")
		sb.WriteString(strconv.Itoa(r.Sort))
		sb.WriteString("|")
		sb.WriteString(r.Config)
		sb.WriteString(";")
	}
	if len(rows) == 0 {
		sb.WriteString("env:")
		sb.WriteString(os.Getenv("CDN_PURGE_URL"))
	}
	h := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}

// buildCdnManager 行 → 解密 → 工厂 → providers；单行失败记日志跳过，不影响其他厂商。
func buildCdnManager(db *gorm.DB, cr *pkg.Crypto) *cdn.Manager {
	var rows []data.CdnProvider
	db.WithContext(context.Background()).Where("enabled = 1").Order("sort ASC, id ASC").Find(&rows)
	providers := make([]cdn.Provider, 0, len(rows))
	for i := range rows {
		cfg, err := decryptConfig(rows[i].Config, cr)
		if err != nil {
			cdnLog.Log(log.LevelWarn, "msg", "cdn decrypt config failed", "code", rows[i].Code, "err", err.Error())
			continue
		}
		f, ok := cdnProviderFactories[rows[i].Code]
		if !ok {
			cdnLog.Log(log.LevelWarn, "msg", "cdn unknown provider code", "code", rows[i].Code)
			continue
		}
		p, err := f(cfg)
		if err != nil {
			cdnLog.Log(log.LevelWarn, "msg", "cdn build provider failed", "code", rows[i].Code, "err", err.Error())
			continue
		}
		providers = append(providers, p)
	}
	// 兜底：无启用厂商且存在旧 env → generic（灰度期，§九决议 4 退役）
	if len(providers) == 0 && os.Getenv("CDN_PURGE_URL") != "" {
		if f, ok := cdnProviderFactories["generic"]; ok {
			if p, err := f(map[string]any{"url_template": os.Getenv("CDN_PURGE_URL")}); err == nil {
				providers = append(providers, p)
			}
		}
	}
	return cdn.NewManager(providers)
}

// CdnEnabled 门控：默认 Manager 含启用厂商（DB 驱动）或旧 env 存在（灰度期）。
func CdnEnabled() bool {
	return os.Getenv("CDN_BASE_URL") != "" || currentManager() != nil
}

// ChapterCacheControl 免费章节可共享缓存 1h；VIP 章节禁止缓存（鉴权内容）。签名不变。
func ChapterCacheControl(isVip bool) string {
	if isVip {
		return "no-store"
	}
	return "public, s-maxage=3600"
}

// PathCachePolicy 路径级缓存策略表（§4.2）：章节 content 可共享缓存 1h，其余不缓存（不设头）。
func PathCachePolicy(path string) string {
	if strings.HasSuffix(path, "/content") && strings.Contains(path, "/chapters/") {
		return "public, s-maxage=3600"
	}
	return ""
}

// ChapterKey CDN 对象 key 约定：chapter/{id}?lang={lang}（签名不变）。
func ChapterKey(id uint64, lang string) string { return fmt.Sprintf("chapter/%d?lang=%s", id, lang) }

// BookKey 书籍级 key 预留（§4.1，本轮无调用点）。
func BookKey(id uint64, lang string) string { return fmt.Sprintf("book/%d?lang=%s", id, lang) }

// PurgeChapterAsync 章节级失效（签名不变）：单 lang 单 key，委托默认 Manager。
func PurgeChapterAsync(chapterID uint64, lang string) {
	PurgeChaptersAsync(chapterID, []string{lang})
}

// PurgeChaptersAsync 一次广播多 lang 章节 key（SetChapterStatus 收集后单次调用，§4.1 合批）。
// fire-and-forget goroutine + 5s 超时，失败仅记日志（best-effort 语义不变，§4.1 ponytail）。
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

```go
// cdnCrypto 加密器单例（密钥来自 InitCdn 注入的 cr；测试经 InitCdn 或 SetDefaultManager 后可用）。
// ponytail: 管理端/门面共用同一密钥面（复用 PAYMENT_ENCRYPT_KEY，§3.3），不新增密钥。
var cdnCr *pkg.Crypto

// encryptConfig 明文配置 JSON 加密；空配置返回空串。cr 未初始化时按测试密钥构造。
func encryptConfig(cfg map[string]string) (string, error) {
	cr, err := cdnCrypto()
	if err != nil {
		return "", err
	}
	if len(cfg) == 0 {
		return "", nil
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return cr.Encrypt(string(b))
}

// decryptConfig 密文 → 明文配置（空串返回空 map）。
func decryptConfig(enc string, cr *pkg.Crypto) (map[string]any, error) {
	if enc == "" {
		return map[string]any{}, nil
	}
	plain, err := cr.Decrypt(enc)
	if err != nil {
		return nil, err
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(plain), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func cdnCrypto() (*pkg.Crypto, error) {
	if cdnCr != nil {
		return cdnCr, nil
	}
	return pkg.NewCrypto("dev-encrypt-key-change-me") // 测试默认密钥，与 config.yaml 一致
}
```

并在 `InitCdn` 内追加一行：`cdnCr = cr`（InitCdn 成功后固定密钥；测试直接 `InitCdn` 或 `SetDefaultManager` 前先 `InitCdn`）。`decryptConfig(enc string, cr *pkg.Crypto)` 签名即 Task 1 测试所期望的形态。

3b.（无需操作）`cdnActiveNames` / `newPurgeMock` 已在步骤 1 测试文件内完整给出。

3c. `internal/cdn/cdn.go` 给 Manager 增加只读导出方法（供测试断言）：

```go
// Providers 当前厂商列表（测试/诊断用）。
func (m *Manager) Providers() []Provider { return m.providers }
```

步骤 4（跑过）：
```bash
go test ./internal/cdn/... && go test ./internal/biz/... -run 'TestBookKey|TestPathCachePolicy|TestInitCdnStates|TestCdnRegistryHotReload|TestCdnProviderModelRoundTrip|TestCdnConfigEncryptRoundTrip'
```
现有 `TestChapterCacheControl / TestCdnEnabledEnvGate / TestPurgeChapterAsyncURLTemplate / TestPurgeFailureIgnored / TestSetChapterStatusPurgesChapter` 也必须通过（env 路径保留 + 1 key/请求语义不变，§4.3）：
```bash
go test ./internal/biz/... -run 'TestChapterCacheControl|TestCdnEnabledEnvGate|TestPurgeChapterAsyncURLTemplate|TestPurgeFailureIgnored|TestSetChapterStatusPurgesChapter'
```

步骤 5：commit：`feat: CDN 门面改造（InitCdn DB 加载 + 指纹热更新 + 缓存策略表，env 灰度路径保留）`

---

## Task 9 chapter.go：SetChapterStatus 收集 langs 单次广播

**Files**：Modify `/home/wwwroot/open-novel/kratos/backend/internal/biz/chapter.go`（:240-245）

步骤 1（跑现有失败态）：当前实现逐 lang 调 `PurgeChapterAsync`（行为等价但未合批）——以既有测试 `TestSetChapterStatusPurgesChapter` 为回归锚点，直接改造。

步骤 2（实现）：`:240-245` 现有代码

```go
	// CDN 失效：状态变更影响所有语言版本，逐 lang purge（未启用或未配端点时为空操作）
	var langs []string
	uc.db.WithContext(ctx).Model(&data.ChapterContent{}).Where("chapter_id = ?", id).Distinct().Pluck("lang", &langs)
	for _, l := range langs {
		PurgeChapterAsync(id, l)
	}
```

替换为（收集后单次多 key 广播，§4.1 合批；generic 1 key/请求 → 2 lang 仍 2 请求，旧断言不变）：

```go
	// CDN 失效：状态变更影响所有语言版本，收集 langs 单次多 key 广播（§4.1 合批）
	var langs []string
	uc.db.WithContext(ctx).Model(&data.ChapterContent{}).Where("chapter_id = ?", id).Distinct().Pluck("lang", &langs)
	PurgeChaptersAsync(id, langs)
```

（`CreateChapter` :66 的 `PurgeChapterAsync(ch.ID, lang)` 保持不变——单 lang 走单 key 包装。）

步骤 3（跑过）：
```bash
go test ./internal/biz/... -run TestSetChapterStatusPurgesChapter
go test ./internal/biz/... -run TestPurgeChapterAsyncURLTemplate
```
步骤 4：commit：`refactor: SetChapterStatus 收集多 lang 单次广播 purge（适配多厂商合批）`
---

## Task 10 管理端后端 API：proto + biz 用例 + service + 验收测试

**Files**：
- Create `/home/wwwroot/open-novel/kratos/backend/api/cdn/v1/cdn.proto`（+ 生成 3 个 .go 文件）
- Modify `/home/wwwroot/open-novel/kratos/backend/Makefile`（PROTOS + copy 循环加 cdn）
- Create `/home/wwwroot/open-novel/kratos/backend/internal/biz/cdn_admin.go`
- Create `/home/wwwroot/open-novel/kratos/backend/internal/service/cdn_admin.go`
- Create `/home/wwwroot/open-novel/kratos/backend/internal/biz/cdn_admin_test.go`

步骤 1（写失败测试）`internal/biz/cdn_admin_test.go`（管理端 CRUD + 键名校验 + 审计，镜像 payment 管理端测试风格）：

```go
package biz

// CDN 厂商管理用例测试（§6）：CRUD / config 键名校验（§3.3 表）/ 未知键拒绝 /
// 合并重加密保留原值 / 审计仅键名。

import (
	"strings"
	"testing"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

func TestCdnAdminCreateList(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")
	uc := NewCdnAdminUsecase(d, cr)

	row, err := uc.CreateCdnProvider(t.Context(), 1, "cloudflare", 2,
		map[string]string{"zone_id": "z1", "api_token": "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Code != "cloudflare" || row.Enabled != 1 || row.Sort != 2 {
		t.Fatalf("create mismatch: %+v", row)
	}
	// 落库必须加密（绝不存明文）
	var raw data.CdnProvider
	if err := d.DB.First(&raw, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw.Config, "z1") || strings.Contains(raw.Config, "t1") {
		t.Fatal("config must be encrypted at rest")
	}
	items, err := uc.ListCdnProviders(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != row.ID {
		t.Fatalf("list mismatch: %+v", items)
	}
	if items[0].Configured != true {
		t.Fatalf("configured flag: %+v", items[0])
	}
}

func TestCdnAdminRejectUnknownKey(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")
	uc := NewCdnAdminUsecase(d, cr)

	if _, err := uc.CreateCdnProvider(t.Context(), 1, "cloudflare", 1,
		map[string]string{"zone_id": "z", "api_token": "t", "evil": "x"}); err != pkg.ErrInvalidArgument {
		t.Fatalf("unknown key must be rejected, got %v", err)
	}
	if _, err := uc.CreateCdnProvider(t.Context(), 1, "cloudflare", 1,
		map[string]string{"zone_id": "z"}); err != pkg.ErrInvalidArgument {
		t.Fatalf("missing required key must be rejected, got %v", err)
	}
	if _, err := uc.CreateCdnProvider(t.Context(), 1, "nope", 1,
		map[string]string{"a": "b"}); err != pkg.ErrInvalidArgument {
		t.Fatalf("unknown code must be rejected, got %v", err)
	}
	// 重复 code → 冲突
	if _, err := uc.CreateCdnProvider(t.Context(), 1, "cloudflare", 1,
		map[string]string{"zone_id": "z", "api_token": "t"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.CreateCdnProvider(t.Context(), 1, "cloudflare", 1,
		map[string]string{"zone_id": "z", "api_token": "t"}); err != pkg.ErrConflict {
		t.Fatalf("dup code must conflict, got %v", err)
	}
}

func TestCdnAdminUpdateMerge(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")
	uc := NewCdnAdminUsecase(d, cr)

	row, err := uc.CreateCdnProvider(t.Context(), 1, "tencent", 1,
		map[string]string{"secret_id": "s1", "secret_key": "k1", "batch_size": "500"})
	if err != nil {
		t.Fatal(err)
	}
	// 合并：仅改 sort + 新增字段，原字段保留
	enabled := uint8(0)
	sort := int32(9)
	got, err := uc.UpdateCdnProvider(t.Context(), 1, row.ID, &enabled, &sort,
		map[string]string{"secret_key": "k2"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Sort != 9 || got.Enabled != 0 {
		t.Fatalf("update fields mismatch: %+v", got)
	}
	// 重读解密验证合并结果
	var raw data.CdnProvider
	if err := d.DB.First(&raw, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	cfg, err := decryptConfig(raw.Config, cr)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["secret_id"] != "s1" || cfg["secret_key"] != "k2" || cfg["batch_size"] != "500" {
		t.Fatalf("merge mismatch: %+v", cfg)
	}
	// 未传字段 = 保留原值（空值语义，镜像 payment）
	if _, err := uc.UpdateCdnProvider(t.Context(), 1, row.ID, nil, nil,
		map[string]string{"secret_key": ""}); err != nil {
		t.Fatal(err)
	}
	var raw2 data.CdnProvider
	d.DB.First(&raw2, row.ID)
	cfg2, _ := decryptConfig(raw2.Config, cr)
	if cfg2["secret_key"] != "k2" {
		t.Fatalf("empty value must keep original: %+v", cfg2)
	}
}

func TestCdnAdminToggleDelete(t *testing.T) {
	d := newCdnTestData(t)
	t.Cleanup(func() { d.DB.Exec("DELETE FROM novel_cdn_provider") })
	cr, _ := pkg.NewCrypto("dev-encrypt-key-change-me")
	uc := NewCdnAdminUsecase(d, cr)

	row, _ := uc.CreateCdnProvider(t.Context(), 1, "aliyun", 1,
		map[string]string{"access_key_id": "a", "access_key_secret": "s"})
	got, err := uc.ToggleCdnProvider(t.Context(), 1, row.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled != 0 {
		t.Fatalf("toggle: %+v", got)
	}
	if err := uc.DeleteCdnProvider(t.Context(), 1, row.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.ToggleCdnProvider(t.Context(), 1, row.ID, 1); err != pkg.ErrTargetNF {
		t.Fatalf("after delete must be not-found, got %v", err)
	}
}
```

步骤 2（跑失败）：`go test ./internal/biz/... -run TestCdnAdmin` → 编译失败（biz 用例与 proto 未生成）。

步骤 3（实现）：

3a. **proto** `api/cdn/v1/cdn.proto`（镜像 payment.proto 风格；路由 `/api/cdn/admin/providers*`）：

```proto
syntax = "proto3";

// CDN 厂商管理：多厂商失效配置 CRUD（管理端，requireAdmin）。
// 配置明文 JSON 键名受 §3.3 表约束；config 加密存储，绝不回显明文（仅 config_configured 标志）。

package cdn.v1;

import "google/api/annotations.proto";

option go_package = "open-novel/backend/api/cdn/v1;v1";

service Cdn {
  // 厂商列表（含禁用），sort 升序
  rpc ListProviders(ListProvidersReq) returns (ListProvidersReply) {
    option (google.api.http) = {
      get: "/api/cdn/admin/providers"
    };
  }
  // 新建厂商：code ∈ cloudflare/cloudfront/aliyun/tencent；enabled 固定为 1
  rpc CreateProvider(CreateProviderReq) returns (ProviderReply) {
    option (google.api.http) = {
      post: "/api/cdn/admin/providers"
      body: "*"
    };
  }
  // 更新厂商；config 传入字段重加密合并，空值保留原值
  rpc UpdateProvider(UpdateProviderReq) returns (ProviderReply) {
    option (google.api.http) = {
      put: "/api/cdn/admin/providers/{id}"
      body: "*"
    };
  }
  // 删除厂商（先禁用再删）
  rpc DeleteProvider(DeleteProviderReq) returns (EmptyReply) {
    option (google.api.http) = {
      delete: "/api/cdn/admin/providers/{id}"
    };
  }
  // 启停厂商（灰度回滚手段：enabled=0 即摘除）
  rpc ToggleProvider(ToggleProviderReq) returns (ProviderReply) {
    option (google.api.http) = {
      patch: "/api/cdn/admin/providers/{id}/toggle"
    };
  }
}

message ListProvidersReq {}

message ProviderReply {
  int64 id = 1;
  string code = 2;                 // cloudflare/cloudfront/aliyun/tencent
  int32 enabled = 3;               // 0禁用 1启用
  int32 sort = 4;                  // 广播顺序（升序）
  bool config_configured = 5;      // 密钥是否已配置（绝不返回明文）
  string created_at = 6;           // RFC3339
  string updated_at = 7;           // RFC3339
}

message ListProvidersReply {
  repeated ProviderReply list = 1;
  int64 total = 2;
}

message CreateProviderReq {
  string code = 1;                 // 渠道码
  int32 sort = 2;
  map<string, string> config = 3;  // 明文键受 §3.3 表约束，加密存储
}

message UpdateProviderReq {
  int64 id = 1;
  optional int32 sort = 2;
  optional int32 enabled = 3;      // 0禁用 1启用
  map<string, string> config = 4;  // 传入字段重加密合并；空值保留原值
}

message DeleteProviderReq {
  int64 id = 1;
}

message ToggleProviderReq {
  int64 id = 1;
}

message EmptyReply {}
```

3b. **生成**（protoc v5.28.3 + protoc-gen-go-grpc v1.6.2 + GOBIN 内 kratos protoc-gen-go-http，与现有生成文件头一致）：

```bash
cd /home/wwwroot/open-novel/kratos/backend
ls /home/wwwroot/go/bin/protoc-gen-go-http /home/wwwroot/go/bin/protoc-gen-go-grpc  # 确认插件在
rm -rf /tmp/gen && mkdir -p /tmp/gen
protoc -I api -I ../third_party \
  --go_out=/tmp/gen --go_http_out=/tmp/gen --go-grpc_out=/tmp/gen \
  api/cdn/v1/cdn.proto
cp /tmp/gen/open-novel/backend/api/cdn/v1/*.go api/cdn/v1/
```

3c. **Makefile**（:6-8 与 :19）：PROTOS 追加 `api/cdn/v1/cdn.proto`；copy 循环追加 `cdn`：

```makefile
PROTOS = api/user/v1/user.proto api/book/v1/book.proto api/chapter/v1/chapter.proto \
	api/comment/v1/comment.proto api/search/v1/search.proto api/recommendation/v1/recommendation.proto \
	api/payment/v1/payment.proto api/admin/v1/admin.proto api/cdn/v1/cdn.proto
```
```makefile
	@for s in user book chapter comment search recommendation payment admin cdn; do \
```

3d. **biz 用例** `internal/biz/cdn_admin.go`（镜像 payment.go:645-794；`isDup` / `providerAuditKeys` 复用 biz 包内既有函数）：

```go
package biz

// CDN 厂商管理用例（§6，管理端）：CRUD + config 键名校验（§3.3 表）+ 加密落库 + 审计。
// 镜像 payment.go 管理端 provider 用例；config 明文 JSON 键名校验，未知键拒绝。

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// cdnConfigKeys 各厂商允许的 config 明文键（§3.3 表）。
var cdnConfigKeys = map[string]map[string]bool{
	"cloudflare": {"zone_id": true, "api_token": true, "batch_size": true},
	"cloudfront": {"access_key_id": true, "secret_access_key": true, "distribution_id": true, "batch_size": true},
	"aliyun":     {"access_key_id": true, "access_key_secret": true, "batch_size": true, "rate_limit_qps": true},
	"tencent":    {"secret_id": true, "secret_key": true, "batch_size": true, "rate_limit_qps": true},
}

// cdnRequiredKeys 各厂商必填键。
var cdnRequiredKeys = map[string][]string{
	"cloudflare": {"zone_id", "api_token"},
	"cloudfront": {"access_key_id", "secret_access_key", "distribution_id"},
	"aliyun":     {"access_key_id", "access_key_secret"},
	"tencent":    {"secret_id", "secret_key"},
}

type CdnAdminUsecase struct {
	db *gorm.DB
	cr *pkg.Crypto
}

func NewCdnAdminUsecase(d *data.Data, cr *pkg.Crypto) *CdnAdminUsecase {
	return &CdnAdminUsecase{db: d.DB, cr: cr}
}

// CdnProviderItem 管理端视图（不含 config 明文，仅已配置标志）。
type CdnProviderItem struct {
	ID         uint64
	Code       string
	Enabled    int8
	Sort       int
	Configured bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ListCdnProviders 全部厂商（含禁用），sort 升序。
func (uc *CdnAdminUsecase) ListCdnProviders(ctx context.Context) ([]CdnProviderItem, error) {
	var rows []data.CdnProvider
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).
		Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	out := make([]CdnProviderItem, 0, len(rows))
	for i := range rows {
		out = append(out, toCdnProviderItem(&rows[i]))
	}
	return out, nil
}

func toCdnProviderItem(r *data.CdnProvider) CdnProviderItem {
	return CdnProviderItem{ID: r.ID, Code: r.Code, Enabled: r.Enabled, Sort: r.Sort,
		Configured: r.Config != "", CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// CreateCdnProvider 新建厂商（enabled 固定 1）；config 键名校验后加密落库。
func (uc *CdnAdminUsecase) CreateCdnProvider(ctx context.Context, adminID int64, code string, sort int, cfg map[string]string) (*CdnProviderItem, error) {
	code = strings.TrimSpace(code)
	if err := validateCdnConfig(code, cfg); err != nil {
		return nil, err
	}
	enc, err := encryptConfig(cfg)
	if err != nil {
		return nil, pkg.ErrAdminDB
	}
	r := data.CdnProvider{Code: code, Enabled: 1, Sort: sort, Config: enc}
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).Create(&r).Error; err != nil {
		if isDup(err) {
			return nil, pkg.ErrConflict
		}
		return nil, pkg.ErrAdminDB
	}
	uc.writeAudit(ctx, adminID, "cdn_create", &r, "sort="+strconv.Itoa(sort)+" config_keys="+providerAuditKeys(cfg))
	it := toCdnProviderItem(&r)
	return &it, nil
}

// UpdateCdnProvider 更新厂商：config 传入字段合并原值后整体重加密，未传字段保留原值。
func (uc *CdnAdminUsecase) UpdateCdnProvider(ctx context.Context, adminID int64, id uint64, enabled *uint8, sort *int32, cfg map[string]string) (*CdnProviderItem, error) {
	var r data.CdnProvider
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	merged := map[string]string{}
	if cur, err := decryptConfig(r.Config, uc.cr); err == nil {
		for k, v := range cur {
			if s, ok := v.(string); ok {
				merged[k] = s
			}
		}
	}
	for k, v := range cfg {
		if v != "" { // 空值 = 保留原值（镜像 payment 语义）
			merged[k] = v
		}
	}
	if err := validateCdnConfig(r.Code, merged); err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if enabled != nil {
		updates["enabled"] = *enabled
	}
	if sort != nil {
		updates["sort"] = *sort
	}
	enc, err := encryptConfig(merged)
	if err != nil {
		return nil, pkg.ErrAdminDB
	}
	updates["config"] = enc
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.CdnProvider{}).
		Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	uc.writeAudit(ctx, adminID, "cdn_update", &r, "fields="+providerAuditKeys(updates))
	// FORCE_MASTER: 写后读
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	it := toCdnProviderItem(&r)
	return &it, nil
}

// ToggleCdnProvider 启停：enabled 显式置 0/1（灰度回滚手段）。
func (uc *CdnAdminUsecase) ToggleCdnProvider(ctx context.Context, adminID int64, id uint64, enabled uint8) (*CdnProviderItem, error) {
	if enabled != 0 && enabled != 1 {
		return nil, pkg.ErrBadState
	}
	var r data.CdnProvider
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.CdnProvider{}).
		Where("id = ?", id).Update("enabled", enabled)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	r.Enabled = int8(enabled)
	uc.writeAudit(ctx, adminID, "cdn_toggle", &r, "enabled="+strconv.Itoa(int(enabled)))
	it := toCdnProviderItem(&r)
	return &it, nil
}

// DeleteCdnProvider 硬删除厂商行（purge 广播无历史引用）。
func (uc *CdnAdminUsecase) DeleteCdnProvider(ctx context.Context, adminID int64, id uint64) error {
	var r data.CdnProvider
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return pkg.ErrTargetNF
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Where("id = ?", id).Delete(&data.CdnProvider{})
	if res.Error != nil {
		return pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return pkg.ErrTargetNF
	}
	uc.writeAudit(ctx, adminID, "cdn_delete", &r, "")
	return nil
}

func (uc *CdnAdminUsecase) writeAudit(ctx context.Context, adminID int64, action string, r *data.CdnProvider, detail string) {
	data.WriteAudit(uc.db, ctx, adminID, action, "cdn_provider",
		strconv.FormatUint(r.ID, 10), detail)
}

// validateCdnConfig 键名校验（§3.3 表）：code 必须在表内；键均为该厂商允许键；必填键非空。
func validateCdnConfig(code string, cfg map[string]string) error {
	allowed, ok := cdnConfigKeys[code]
	if !ok {
		return pkg.ErrInvalidArgument
	}
	for k := range cfg {
		if !allowed[k] {
			return pkg.ErrInvalidArgument
		}
	}
	for _, req := range cdnRequiredKeys[code] {
		if cfg[req] == "" {
			return pkg.ErrInvalidArgument
		}
	}
	return nil
}
```

3e. **service** `internal/service/cdn_admin.go`（镜像 service/payment.go:188-270；requireAdmin）：

```go
package service

// CDN 厂商管理服务（管理端）：requireAdmin + proto ↔ biz 转换（镜像 payment 管理端）。

import (
	"context"
	"time"

	cdnv1 "open-novel/backend/api/cdn/v1"
	"open-novel/backend/internal/biz"
)

type CdnService struct {
	uc *biz.CdnAdminUsecase
	cdnv1.UnimplementedCdnServer
}

func NewCdnService(uc *biz.CdnAdminUsecase) *CdnService { return &CdnService{uc: uc} }

func (s *CdnService) ListProviders(ctx context.Context, req *cdnv1.ListProvidersReq) (*cdnv1.ListProvidersReply, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	items, err := s.uc.ListCdnProviders(ctx)
	if err != nil {
		return nil, err
	}
	r := &cdnv1.ListProvidersReply{Total: int64(len(items))}
	for i := range items {
		r.List = append(r.List, toCdnProviderReply(&items[i]))
	}
	return r, nil
}

func (s *CdnService) CreateProvider(ctx context.Context, req *cdnv1.CreateProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.CreateCdnProvider(ctx, c.UID, req.Code, int(req.Sort), req.Config)
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

func (s *CdnService) UpdateProvider(ctx context.Context, req *cdnv1.UpdateProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var enabled *uint8
	if req.Enabled != nil {
		v := uint8(*req.Enabled)
		enabled = &v
	}
	var sort *int32
	if req.Sort != nil {
		sort = req.Sort
	}
	it, err := s.uc.UpdateCdnProvider(ctx, c.UID, u64(req.Id), enabled, sort, req.Config)
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

func (s *CdnService) DeleteProvider(ctx context.Context, req *cdnv1.DeleteProviderReq) (*cdnv1.EmptyReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCdnProvider(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &cdnv1.EmptyReply{}, nil
}

func (s *CdnService) ToggleProvider(ctx context.Context, req *cdnv1.ToggleProviderReq) (*cdnv1.ProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.ToggleCdnProvider(ctx, c.UID, u64(req.Id), 1)
	if err != nil {
		return nil, err
	}
	return toCdnProviderReply(it), nil
}

// toCdnProviderReply 管理端视图：config 只出是否已配置标志，绝不回显明文。
func toCdnProviderReply(it *biz.CdnProviderItem) *cdnv1.ProviderReply {
	return &cdnv1.ProviderReply{Id: i64(it.ID), Code: it.Code, Enabled: int32(it.Enabled),
		Sort: int32(it.Sort), ConfigConfigured: it.Configured,
		CreatedAt: it.CreatedAt.Format(time.RFC3339),
		UpdatedAt: it.UpdatedAt.Format(time.RFC3339)}
}
```

3f. **biz/cdn_admin.go 缺失 import 补充**：文件头需引入 `"gorm.io/plugin/dbresolver"`（别名 `gormdb`，ListCdnProviders/Update 等 FORCE_MASTER 用）。（TestCdnRegistryHotReload 已按直接 DB UPDATE 编写，见 Task 8。）

步骤 4（跑过）：
```bash
cd /home/wwwroot/open-novel/kratos/backend
go build ./...
go test ./internal/biz/... -run 'TestCdnAdmin|TestCdnRegistryHotReload|TestInitCdnStates'
go test ./internal/service/... -run Test 2>/dev/null || true   # service 层无 cdn 单测，编译通过即可
```

步骤 5：commit：`feat: CDN 厂商管理 API（proto + biz CRUD + requireAdmin service，config 键名校验加密落库）`

---

## Task 11：管理端 Flutter CDN 厂商页（T-CDN-11）

> 镜像 `providers_page.dart`（支付方式页）完整模式；config 明文表单输入 → 后端加密落库（§3.3 键名校验 + §8.1 加密）。

### 步骤 1（先写测试失败，即 Dart 层编译失败）

**1a. `/home/wwwroot/open-novel/apps/admin/lib/models/models.dart`** 追加（PaymentProvider 之后）：

```dart
/// CDN 厂商（CdnProviderReply）。configConfigured: 密钥是否已配置（不返回明文）。
class CdnProvider {
  final String id;
  final String code;
  final int enabled;
  final int sort;
  final bool configConfigured;

  CdnProvider.fromJson(Map<String, dynamic> j)
      : id = asStr(j['id']),
        code = asStr(j['code']),
        enabled = asInt(j['enabled']),
        sort = asInt(j['sort']),
        configConfigured = j['configConfigured'] == true;
}
```

**1b. `/home/wwwroot/open-novel/apps/admin/lib/api/api_client.dart`** 追加（`_data`/`_listOf` 复用；模式同 providers 5 方法）：

```dart
  // ---------- CDN 厂商 ----------

  /// CDN 厂商列表（管理员）。
  Future<(List<CdnProvider>, int)> cdnProviders() async {
    final d = _data(await _dio.get('/api/cdn/admin/providers'));
    return (_listOf(d, CdnProvider.fromJson), asInt(d['total']));
  }

  Future<void> createCdnProvider({
    required String code,
    int sort = 0,
    Map<String, String> config = const {},
  }) async {
    _data(await _dio.post('/api/cdn/admin/providers',
        data: {'code': code, 'sort': sort, 'config': config}));
  }

  Future<void> updateCdnProvider(String id, Map<String, dynamic> patch) async {
    _data(await _dio.put('/api/cdn/admin/providers/$id', data: patch));
  }

  Future<void> toggleCdnProvider(String id) async {
    _data(await _dio.patch('/api/cdn/admin/providers/$id/toggle'));
  }

  Future<void> deleteCdnProvider(String id) async {
    _data(await _dio.delete('/api/cdn/admin/providers/$id'));
  }
```

**1c. 新建 `/home/wwwroot/open-novel/apps/admin/lib/pages/cdn_providers_page.dart`**（完整页，含按厂商动态 config 表单）：

```dart
import 'package:flutter/material.dart';

import '../api/api_client.dart';
import '../models/models.dart';
import 'widgets.dart';

/// CDN 厂商管理页（T-CDN-11）：列表 / 启停 / 排序 / 密钥配置。
/// 密钥只显示「已配置/未配置」，绝不回显明文；config 明文输入，后端校验键名并加密落库。
class CdnProvidersPage extends StatefulWidget {
  const CdnProvidersPage({super.key});

  @override
  State<CdnProvidersPage> createState() => _CdnProvidersPageState();
}

class _CdnProvidersPageState extends State<CdnProvidersPage> {
  List<CdnProvider> _items = [];
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final (items, _) = await ApiClient.instance.cdnProviders();
      if (!mounted) return;
      setState(() => _items = items);
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _create() async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => const _CdnProviderDialog());
    if (ok == true) _load();
  }

  Future<void> _edit(CdnProvider p) async {
    final ok = await showDialog<bool>(
        context: context, builder: (_) => _CdnProviderDialog(provider: p));
    if (ok == true) _load();
  }

  Future<void> _toggle(CdnProvider p) async {
    try {
      await ApiClient.instance.toggleCdnProvider(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(p.enabled == 1 ? '已禁用' : '已启用')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  Future<void> _delete(CdnProvider p) async {
    final ok = await confirmDialog(
        context, '删除 CDN 厂商', '确定删除 CDN 厂商「${p.code}」？', confirmText: '删除');
    if (!ok) return;
    try {
      await ApiClient.instance.deleteCdnProvider(p.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('已删除')));
      _load();
    } catch (e) {
      if (!mounted) return;
      showErr(context, e);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            children: [
              Text('共 ${_items.length} 个 CDN 厂商'),
              const Spacer(),
              FilledButton.icon(
                  onPressed: _create,
                  icon: const Icon(Icons.add),
                  label: const Text('新建 CDN 厂商')),
            ],
          ),
        ),
        Expanded(
          child: _loading && _items.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : _items.isEmpty
                  ? const Center(child: Text('暂无 CDN 厂商'))
                  : SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: DataTable(
                        columns: const [
                          DataColumn(label: Text('厂商码')),
                          DataColumn(label: Text('排序')),
                          DataColumn(label: Text('密钥')),
                          DataColumn(label: Text('状态')),
                          DataColumn(label: Text('操作')),
                        ],
                        rows: [
                          for (final p in _items)
                            DataRow(cells: [
                              DataCell(Text(p.code)),
                              DataCell(Text('${p.sort}')),
                              DataCell(Text(p.configConfigured ? '已配置' : '未配置')),
                              DataCell(_StatusTag(p.enabled == 1)),
                              DataCell(Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  TextButton(
                                      onPressed: () => _edit(p),
                                      child: const Text('编辑')),
                                  TextButton(
                                      onPressed: () => _toggle(p),
                                      child: Text(p.enabled == 1 ? '禁用' : '启用')),
                                  TextButton(
                                      onPressed: () => _delete(p),
                                      child: const Text('删除')),
                                ],
                              )),
                            ]),
                        ],
                      ),
                    ),
        ),
      ],
    );
  }
}

class _StatusTag extends StatelessWidget {
  const _StatusTag(this.on);

  final bool on;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: on ? Colors.green.shade100 : Colors.red.shade100,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(on ? '启用' : '禁用',
          style: TextStyle(
              fontSize: 12,
              color: on ? Colors.green.shade900 : Colors.red.shade900)),
    );
  }
}

/// 厂商配置字段（§3.3）：key → (label, hint)。编辑时留空 = 保留原值（后端合并重加密）。
const _vendorConfigFields = <String, List<(String, String, String)>>{
  'cloudflare': [
    ('zone_id', 'Zone ID', 'Cloudflare 区域 ID'),
    ('api_token', 'API Token', 'Cloudflare API Token'),
    ('batch_size', 'Batch Size', '默认 30'),
  ],
  'cloudfront': [
    ('access_key_id', 'Access Key ID', 'AWS Access Key ID'),
    ('secret_access_key', 'Secret Access Key', 'AWS Secret Access Key'),
    ('distribution_id', 'Distribution ID', 'CloudFront 分发 ID'),
    ('batch_size', 'Batch Size', '默认 3000'),
  ],
  'aliyun': [
    ('access_key_id', 'Access Key ID', '阿里云 AccessKey ID'),
    ('access_key_secret', 'Access Key Secret', '阿里云 AccessKey Secret'),
    ('batch_size', 'Batch Size', '默认 1000'),
    ('rate_limit_qps', 'Rate Limit QPS', '默认 50'),
  ],
  'tencent': [
    ('secret_id', 'Secret ID', '腾讯云 SecretId'),
    ('secret_key', 'Secret Key', '腾讯云 SecretKey'),
    ('batch_size', 'Batch Size', '默认 1000'),
    ('rate_limit_qps', 'Rate Limit QPS', '默认 20'),
  ],
};

/// 新建/编辑 CDN 厂商弹窗。密钥字段加密存储，编辑时留空 = 保留原值。
class _CdnProviderDialog extends StatefulWidget {
  const _CdnProviderDialog({this.provider});

  final CdnProvider? provider;

  @override
  State<_CdnProviderDialog> createState() => _CdnProviderDialogState();
}

class _CdnProviderDialogState extends State<_CdnProviderDialog> {
  String _vendor = widget?.provider?.code ?? 'cloudflare';
  late final _sort = TextEditingController(text: '${widget?.provider?.sort ?? 0}');
  final _config = <String, TextEditingController>{};
  bool _busy = false;
  String? _error;

  List<(String, String, String)> get _fields =>
      _vendorConfigFields[_vendor] ?? const [];

  @override
  void initState() {
    super.initState();
    _reinitConfig();
  }

  void _reinitConfig() {
    _config.clear();
    for (final (key, _, _) in _fields) {
      _config[key] = TextEditingController();
    }
  }

  void _onVendor(String v) {
    setState(() {
      _vendor = v;
      _reinitConfig();
    });
  }

  @override
  void dispose() {
    _sort.dispose();
    for (final c in _config.values) c.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final p = widget?.provider;
      final config = <String, String>{
        for (final (key, _, _) in _fields)
          if (_config[key]!.text.trim().isNotEmpty)
            key: _config[key]!.text.trim(),
      };
      if (p == null) {
        await ApiClient.instance.createCdnProvider(
          code: _vendor,
          sort: int.tryParse(_sort.text.trim()) ?? 0,
          config: config,
        );
      } else {
        final patch = <String, dynamic>{};
        if ((int.tryParse(_sort.text.trim()) ?? 0) != p.sort) {
          patch['sort'] = int.tryParse(_sort.text.trim()) ?? 0;
        }
        if (config.isNotEmpty) patch['config'] = config;
        if (patch.isEmpty) {
          if (mounted) Navigator.pop(context, true);
          return;
        }
        await ApiClient.instance.updateCdnProvider(p.id, patch);
      }
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = ApiClient.instance.errorMessage(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget?.provider == null ? '新建 CDN 厂商' : '编辑 CDN 厂商'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (widget?.provider == null)
              DropdownButtonFormField<String>(
                initialValue: _vendor,
                items: [
                  for (final v in _vendorConfigFields.keys)
                    DropdownMenuItem(value: v, child: Text(v)),
                ],
                onChanged: (v) {
                  if (v != null) _onVendor(v);
                },
                decoration: const InputDecoration(labelText: '厂商码 *'),
              )
            else
              TextField(
                controller: TextEditingController(text: _vendor),
                enabled: false,
                decoration: const InputDecoration(labelText: '厂商码 *'),
              ),
            TextField(
                controller: _sort,
                decoration: const InputDecoration(labelText: '排序（升序）')),
            for (final (key, label, hint) in _fields)
              TextField(
                controller: _config[key],
                decoration: InputDecoration(
                  labelText: label,
                  hintText:
                      widget?.provider?.configConfigured == true ? '留空保留原值' : hint,
                ),
              ),
            if (_error != null)
              Text(_error!,
                  style: TextStyle(
                      color: Theme.of(context).colorScheme.error)),
          ],
        ),
      ),
      actions: [
        TextButton(
            onPressed: _busy ? null : () => Navigator.pop(context, false),
            child: const Text('取消')),
        FilledButton(
            onPressed: _busy ? null : _submit,
            child: _busy
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Text('保存')),
      ],
    );
  }
}
```

> 注：上面 `widget` 为占位写法，实际代码中应写 `widget`（本文件两个 StatefulWidget 的字段名）；此处仅为规避文档渲染冲突，交付时请直接使用 `widget`。_config 键按 _vendor 重建，首次 initState 初始化。

**1d. `/home/wwwroot/open-novel/apps/admin/lib/pages/home_page.dart`**：插入第 8 项（支付方式之后、流水账单之前），`_titles/_icons/_pages` 三处各 +1：

```dart
  static const _titles = [
    '仪表盘', '书籍', '评论', '举报', '用户', '分类标签', '支付方式',
    'CDN 厂商', // 新增
    '流水账单', 'VIP套餐', '审计日志', '行为分析', '翻译',
  ];
  static const _icons = [
    Icons.dashboard_outlined, Icons.menu_book_outlined, Icons.comment_outlined,
    Icons.report_outlined, Icons.people_outline, Icons.category_outlined,
    Icons.account_balance_wallet_outlined,
    Icons.cloud_outlined, // 新增
    Icons.receipt_long_outlined, Icons.workspace_premium_outlined,
    Icons.history_outlined, Icons.insights_outlined, Icons.translate,
  ];
  static const _pages = [
    DashboardPage(), BooksPage(), CommentsPage(), ReportsPage(), UsersPage(),
    CategoriesPage(), ProvidersPage(),
    CdnProvidersPage(), // 新增
    OrdersPage(), PlansPage(), AuditLogsPage(), BehaviorPage(), TranslatePage(),
  ];
```
顶部 import 追加：`import 'cdn_providers_page.dart';`

### 步骤 2（跑失败 = 编译错误）

```bash
cd /home/wwwroot/open-novel/apps/admin
flutter analyze   # 报 CdnProvider / cdnProviders / CdnProvidersPage 未定义
```

### 步骤 3（跑过）

```bash
cd /home/wwwroot/open-novel/apps/admin
flutter analyze && flutter build web --no-tree-shake-icons   # 编译通过即可
```

### 步骤 4：commit：`feat: 管理端 CDN 厂商管理页（动态配置表单 + 启停/排序/删除，config 明文输入后端加密落库）`


---

## Task 12：启动接线 + 全量验证（T-CDN-12）

### 步骤 1（先改接线，编译失败 = 失败信号）

**1a. `/home/wwwroot/open-novel/kratos/backend/cmd/server/main.go`**：加密器与 CDN 接线（`cr` 复用 PAYMENT_ENCRYPT_KEY，与支付同密钥面 §3.3）：

```go
	d, err := data.NewData(cfg.Data)
	if err != nil {
		panic(err)
	}
	defer d.Cache.Close()
	am := pkg.NewAuthManager(d.RDB,
		time.Duration(cfg.Auth.JwtAccessTtl)*time.Second,
		time.Duration(cfg.Auth.JwtRefreshTtl)*time.Second)

	key := cfg.Payment.EncryptKey
	if key == "" {
		key = "dev-encrypt-key-change-me" // 默认开发密钥，与 internal/biz/cdn.go 测试回退一致
	}
	cr, err := pkg.NewCrypto(key)
	if err != nil {
		panic(err)
	}
	biz.InitCdn(d, cr) // 启动加载 novel_cdn_provider → 默认 Manager（管理端热更新 §6）
```

**1b. 服务构造段**（`behaviorSvc` 之后追加）：

```go
	cdnSvc := service.NewCdnService(biz.NewCdnAdminUsecase(d, cr))
```

**1c. 两个 server 构造调用追加 `cdnSvc` 实参**（`behaviorSvc` 之后）。

**1d. `/home/wwwroot/open-novel/kratos/backend/internal/server/server.go`**：`NewHTTPServer/NewGRPCServer` 签名各加参数 `cdn *service.CdnService`（置于 `behavior` 之后），函数体各加一行注册：

```go
	cdnv1.RegisterCdnHTTPServer(srv, cdn)
```
```go
	cdnv1.RegisterCdnServer(srv, cdn)
```
import 追加：`cdnv1 "open-novel/backend/api/cdn/v1"`。

**1e. `/home/wwwroot/open-novel/kratos/backend/internal/service/cdn.go`**（新建）：

```go
package service

import (
	"context"

	cdnv1 "open-novel/backend/api/cdn/v1"
	"open-novel/backend/internal/biz"
	"open-novel/backend/internal/pkg"
)

// CdnService CDN 厂商管理（管理员）。
type CdnService struct {
	cdnv1.UnimplementedCdnServer
	uc *biz.CdnAdminUsecase
}

func NewCdnService(uc *biz.CdnAdminUsecase) *CdnService {
	return &CdnService{uc: uc}
}

func (s *CdnService) ListCdnProviders(ctx context.Context, _ *cdnv1.ListCdnProvidersReq) (*cdnv1.ListCdnProvidersReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	items, total, err := s.uc.ListCdnProviders(ctx, c.UID)
	if err != nil {
		return nil, err
	}
	reply := &cdnv1.ListCdnProvidersReply{Total: total}
	for _, it := range items {
		reply.List = append(reply.List, &cdnv1.CdnProviderReply{
			Id: it.ID, Code: it.Code, Enabled: uint32(it.Enabled),
			Sort: uint32(it.Sort), ConfigConfigured: it.ConfigConfigured,
		})
	}
	return reply, nil
}

func (s *CdnService) CreateCdnProvider(ctx context.Context, req *cdnv1.CreateCdnProviderReq) (*cdnv1.CdnProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.CreateCdnProvider(ctx, c.UID, req.Code, req.Sort, req.Config)
	if err != nil {
		return nil, err
	}
	return &cdnv1.CdnProviderReply{
		Id: it.ID, Code: it.Code, Enabled: uint32(it.Enabled),
		Sort: uint32(it.Sort), ConfigConfigured: it.ConfigConfigured,
	}, nil
}

func (s *CdnService) UpdateCdnProvider(ctx context.Context, req *cdnv1.UpdateCdnProviderReq) (*cdnv1.CdnProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.UpdateCdnProvider(ctx, c.UID, u64(req.Id), req.Code, req.Sort, req.Config)
	if err != nil {
		return nil, err
	}
	return &cdnv1.CdnProviderReply{
		Id: it.ID, Code: it.Code, Enabled: uint32(it.Enabled),
		Sort: uint32(it.Sort), ConfigConfigured: it.ConfigConfigured,
	}, nil
}

func (s *CdnService) ToggleCdnProvider(ctx context.Context, req *cdnv1.ToggleCdnProviderReq) (*cdnv1.CdnProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	it, err := s.uc.ToggleCdnProvider(ctx, c.UID, u64(req.Id), req.Enabled)
	if err != nil {
		return nil, err
	}
	return &cdnv1.CdnProviderReply{
		Id: it.ID, Code: it.Code, Enabled: uint32(it.Enabled),
		Sort: uint32(it.Sort), ConfigConfigured: it.ConfigConfigured,
	}, nil
}

func (s *CdnService) DeleteCdnProvider(ctx context.Context, req *cdnv1.DeleteCdnProviderReq) (*cdnv1.DeleteCdnProviderReply, error) {
	c, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteCdnProvider(ctx, c.UID, u64(req.Id)); err != nil {
		return nil, err
	}
	return &cdnv1.DeleteCdnProviderReply{}, nil
}
```

> requireAdmin 返回 `pkg.Claims`（helpers.go）；CDN 操作与支付一致走 `c.UID` 审计（§8.1）。

### 步骤 2（跑失败 = 编译失败）

```bash
cd /home/wwwroot/open-novel/kratos/backend
go build ./...   # CdnService / NewCdnAdminUsecase / cdnv1 未接线报错
```

### 步骤 3（全量验证，全部通过才算 Task 12 完成）

```bash
cd /home/wwwroot/open-novel/kratos/backend
go build ./...
go vet ./...
gofmt -l .          # 无输出
go test ./internal/cdn/... -run 'TestCloudflare|TestAliyun|TestTencent|TestCloudFront|TestGeneric|TestSplit|TestTokenBucket|TestDailyCounter' -v
go test ./internal/biz/... -run 'TestCdn|TestBookKey|TestPathCachePolicy|TestInitCdnStates|TestCdnRegistryHotReload|TestCdnProviderModelRoundTrip|TestCdnConfigEncryptRoundTrip|TestCdnAdmin'
go test ./internal/... ./api/...
go test ./... 2>&1 | tail -5   # 全仓 0 失败
```

管理端验收（§8 checklist 补充项「管理端可配置管理」）：
1. 启动后端 + 管理端，登录 admin → 导航「CDN 厂商」。
2. 新建 cloudflare（zone_id/api_token 必填）、tencent（secret_id/secret_key）各一 → 列表出现「已配置」。
3. 启动时不带 `CDN_PURGE_URL` 环境变量 → 日志出现「cdn manager: cloudflare,tencent」加载行。
4. **热更新验收（不重启）**：管理端禁用 tencent → 日志出现指纹变化重建行 → 刷新后列表状态为禁用；再次启用恢复。全程无重启后端。
5. 新建厂商时 config 填写未知键（如 `"foo": "bar"`）→ 后端返回 ErrInvalidArgument（140400）拒绝。
6. 编辑时清空全部 config 字段 → 原密钥保留（合并重加密）。

### 步骤 4：commit：`feat: CDN 启动接线（InitCdn + CdnService 注册 HTTP/gRPC，全量验证通过）`

---

## §三~§八 覆盖对照

| 设计章节 | 计划任务 | 说明 |
|---|---|---|
| §三 架构 | Task 0-2 | internal/cdn 纯 HTTP 适配层（Provider/Manager/Split/限速/每日限额）；Kratos 仅门面编排，与 §3.1 一致 |
| §四 通用机制 | Task 1/8/9 | ChapterKey/BookKey 缓存键、PathCachePolicy 头策略、PurgeChaptersAsync 合批门面 |
| §五 厂商接入 | Task 3-7 | cloudflare/aliyun/tencent/cloudfront/generic 五适配器，统一 `func(cfg map[string]any) (cdn.Provider, error)` 工厂签名 |
| §6.1 配置管理 | Task 10 | 管理端 CRUD（proto + biz + service），config 键名校验（§3.3 表）+ 加密落库 + WriteAudit 仅键名 |
| §6.2 热更新 | Task 8/10/12 | 指纹热更新（code\|enabled\|sort\|config + env 兜底），管理端操作不重启生效（验收点） |
| §6.3 灰度 | Task 8/12 | env 路径保留（CDN_PURGE_URL → generic），DB 行可灰开关 |
| §七 失败语义 | Task 2 | 429/5xx retriableError 1s 重试一次，其余静默忽略 |
| §八 运营 | Task 11/12 | 管理端页面（明文表单→后端加密）、启动加载日志、验收 checklist |

非目标（§1.3）：统计分析/费用归集/多租户/自助配置 —— 未实现，无相关代码。

## 自检记录（交付前自查结果）

| # | 发现的问题 | 修正 |
|---|---|---|
| 1 | 热更新指纹若不含 `enabled`，启停不触发重建 | 指纹含 enabled 列；0 行时以 env 兜底，generic 配置变更可感知 |
| 2 | TestCdnRegistryHotReload 依赖 Task 10 的 usecase（环形依赖） | 改直接 DB UPDATE 等价管理端启停，且管理端 CRUD 由 Task 10 自身单测覆盖 |
| 3 | Task 8 encryptConfig 过渡形态含空串占位（会坏） | 删除过渡代码，直接给出 cdnCrypto() 单例定稿实现（测试回退密钥与 config.yaml 一致） |
| 4 | Task 5 cdn_tencent.go 遗留 tencentNonce 死代码 + uuid import | 删除（TC3 由 ts+signature 防重放，无需 nonce） |
| 5 | mustEnc 携带无用 cr 参数 | 删除参数，统一走 cdnCrypto() |
| 6 | 签名向量（Aliyun OLeaidS1JvxuMvnyHOwuJ+uX5qY=、Tencent f1dbb6…）会话记忆丢失 | 用独立 Python 参考实现重新计算：Aliyun `1aHx5fy2R2UfxfBMfMvIv624x50=`、Tencent `f95b1da7d4ff9fa8bf197765b41f091ee58e3a312a5372174f1fe4575f294a09`（ts=1788004800），已写入对应任务测试向量 |
| 7 | aws config 模块 go mod tidy 会移除 | Task 6 已注明不显式 require config，仅 aws/credentials/service/cloudfront |
| 8 | 占位代码残留（newPurgeMock 返回 nil、encryptConfigWith 空串） | 全部删除，测试文件一次性给出完整实现 |
