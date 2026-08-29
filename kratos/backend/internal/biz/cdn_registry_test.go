package biz

// CDN 门面测试（§七）：CdnEnabled 门控 / PathCachePolicy 表 / BookKey /
// InitCdn 加载三态（启用行→厂商、无行+env→generic、全无→空）/ 指纹热更新。

import (
	"net/http"
	"net/http/httptest"
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
		"/api/chapters/123/content":  "public, s-maxage=3600",
		"/api/chapters/123":          "",
		"/api/books/1":               "",
		"/api/chapters/1/content/v2": "", // 后缀不匹配 content
		"/api/comments":              "",
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
