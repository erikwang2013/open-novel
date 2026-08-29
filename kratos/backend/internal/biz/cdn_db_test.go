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
