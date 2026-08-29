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
	got, err := uc.ToggleCdnProvider(t.Context(), 1, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled != 0 {
		t.Fatalf("toggle must disable, got %+v", got)
	}
	got, err = uc.ToggleCdnProvider(t.Context(), 1, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled != 1 {
		t.Fatalf("toggle must re-enable, got %+v", got)
	}
	if err := uc.DeleteCdnProvider(t.Context(), 1, row.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.ToggleCdnProvider(t.Context(), 1, row.ID); err != pkg.ErrTargetNF {
		t.Fatalf("after delete must be not-found, got %v", err)
	}
}
