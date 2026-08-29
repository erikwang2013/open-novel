package biz

// CDN 厂商管理用例（§6，管理端）：CRUD + config 键名校验（§3.3 表）+ 加密落库 + 审计。
// 镜像 payment.go 管理端 provider 用例；config 明文 JSON 键名校验，未知键拒绝。

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

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
	auditKeys := make(map[string]any, len(cfg))
	for k := range cfg {
		auditKeys[k] = ""
	}
	uc.writeAudit(ctx, adminID, "cdn_create", &r, "sort="+strconv.Itoa(sort)+" config_keys="+providerAuditKeys(auditKeys))
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

// ToggleCdnProvider 启停：enabled 翻转（镜像 payment.ToggleProvider，§6 灰度回滚）。
func (uc *CdnAdminUsecase) ToggleCdnProvider(ctx context.Context, adminID int64, id uint64) (*CdnProviderItem, error) {
	var r data.CdnProvider
	if err := uc.db.Clauses(gormdb.Write).WithContext(ctx).First(&r, id).Error; err != nil {
		return nil, pkg.ErrTargetNF
	}
	next := int8(1)
	if r.Enabled == 1 {
		next = 0
	}
	res := uc.db.Clauses(gormdb.Write).WithContext(ctx).Model(&data.CdnProvider{}).
		Where("id = ?", id).Update("enabled", next)
	if res.Error != nil {
		return nil, pkg.ErrAdminDB
	}
	if res.RowsAffected == 0 {
		return nil, pkg.ErrTargetNF
	}
	r.Enabled = next
	uc.writeAudit(ctx, adminID, "cdn_toggle", &r, "enabled="+strconv.Itoa(int(next)))
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
