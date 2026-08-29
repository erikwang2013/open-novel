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
	cdnCr = cr // 加密密钥面固定（与支付同 KEY，§3.3）
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
