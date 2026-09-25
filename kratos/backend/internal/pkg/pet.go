package pkg

// 项目宠物 Novi：GET /api/pet.svg 返回的活体状态图。
// 源图 docs/pet.svg（与本目录 pet.svg 同一份），表情与状态牌按依赖探测结果
// 注入 svg 的 <!--MOOD_FACE--> / <!--MOOD_BADGE--> 两处标记位。

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

//go:embed pet.svg
var petSVGTemplate string

// Mood 宠物心情，由 MySQL / Redis / OpenSearch 三项探测结果决定。
type Mood string

const (
	MoodHappy    Mood = "happy"    // 三项全通
	MoodDegraded Mood = "degraded" // 部分不可用（降级运行）
	MoodDown     Mood = "down"     // 三项全不可用
)

const (
	faceMarker  = "<!--MOOD_FACE_START--><!--MOOD_FACE_END-->"
	badgeMarker = "<!--MOOD_BADGE_START--><!--MOOD_BADGE_END-->"
	probeBudget = 2 * time.Second
)

// moodColor 心情主色（状态牌与圆点）。
var moodColor = map[Mood]string{
	MoodHappy:    "#2e7d32",
	MoodDegraded: "#ef6c00",
	MoodDown:     "#c62828",
}

// classifyMood 三项探测结果 → 心情：全通 happy，全断 down，其余 degraded。
func classifyMood(db, cache, search bool) Mood {
	switch {
	case db && cache && search:
		return MoodHappy
	case !db && !cache && !search:
		return MoodDown
	default:
		return MoodDegraded
	}
}

// PetStatus 一次探测的结果快照。
type PetStatus struct {
	Mood   Mood
	Down   []string // 不可用依赖名（mysql / redis / opensearch）
	Uptime time.Duration
}

// Pet 持有依赖句柄与进程启动时间，渲染活体宠物图。
type Pet struct {
	db    *gorm.DB
	rdb   *redis.Client
	esURL string
	start time.Time
}

// NewPet esAddr 为 OpenSearch 地址（可带或不带 http:// 前缀）。
func NewPet(db *gorm.DB, rdb *redis.Client, esAddr string) *Pet {
	if esAddr != "" && !strings.HasPrefix(esAddr, "http") {
		esAddr = "http://" + esAddr
	}
	return &Pet{db: db, rdb: rdb, esURL: strings.TrimRight(esAddr, "/"), start: time.Now()}
}

// Status 探测三项依赖（整体 2s 预算）并分类心情。
func (p *Pet) Status(ctx context.Context) PetStatus {
	ctx, cancel := context.WithTimeout(ctx, probeBudget)
	defer cancel()

	dbOK, cacheOK, searchOK := true, true, true
	var down []string

	var one int
	if p.db == nil || p.db.WithContext(ctx).Raw("SELECT 1").Scan(&one).Error != nil {
		dbOK, down = false, append(down, "mysql")
	}
	if p.rdb == nil || p.rdb.Ping(ctx).Err() != nil {
		cacheOK, down = false, append(down, "redis")
	}
	if !p.searchHealthy(ctx) {
		searchOK, down = false, append(down, "opensearch")
	}
	return PetStatus{Mood: classifyMood(dbOK, cacheOK, searchOK), Down: down, Uptime: time.Since(p.start)}
}

// searchHealthy 探测 OpenSearch 集群健康（200 即视为可用）。
func (p *Pet) searchHealthy(ctx context.Context) bool {
	if p.esURL == "" {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.esURL+"/_cluster/health", nil)
	if err != nil {
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // 连接复用需要读完 body
	return resp.StatusCode == http.StatusOK
}

// SVG 探测 + 渲染，返回可直接写出的 image/svg+xml 内容。
func (p *Pet) SVG(ctx context.Context) string {
	return Render(p.Status(ctx))
}

// Render 把心情注入标记位；标记位缺失时原样返回（图仍可用）。
func Render(s PetStatus) string {
	svg := strings.Replace(petSVGTemplate, faceMarker, faceSnippet(s.Mood), 1)
	return strings.Replace(svg, badgeMarker, badgeSnippet(s), 1)
}

// faceSnippet 注入的表情叠加层（都画在墨身之上）：
// degraded 蒙一层灰纱 + 右眼半垂，down 灰纱加重 + 双眼闭合 + 飘 z。
//
// 眼睑直接用弧线画，不依赖 pet.svg 里的 clipPath：墨灵眨眼靠的是整只眼纵向缩放，
// 源图里没有眼睑裁剪路径可复用（不像上一版那只猫头鹰）。
func faceSnippet(m Mood) string {
	drop := `M256 96 Q344 186 344 224 A88 88 0 0 1 168 224 Q168 186 256 96 Z`
	lid := func(class, d string) string {
		return fmt.Sprintf(`<path class="%s" d="%s" fill="none" stroke="#1F2C3D" stroke-width="4" stroke-linecap="round"/>`, class, d)
	}
	switch m {
	case MoodDegraded:
		return fmt.Sprintf(`<path d="%s" fill="#90A4AE" opacity="0.28"/>`, drop) +
			lid("lid-r", "M269 208 Q288 218 307 208") // 右眼半垂
	case MoodDown:
		return fmt.Sprintf(`<path d="%s" fill="#90A4AE" opacity="0.46"/>`, drop) +
			lid("lid-l", "M205 216 Q224 234 243 216") +
			lid("lid-r", "M269 216 Q288 234 307 216") +
			`<text x="352" y="146" font-size="24" font-weight="bold" fill="#90a4ae" font-family="sans-serif">z</text>` +
			`<text x="378" y="118" font-size="17" font-weight="bold" fill="#b0bec5" font-family="sans-serif">z</text>`
	default:
		return ""
	}
}

// badgeSnippet 注入的状态牌（画布左下，避开头部与书）。
func badgeSnippet(s PetStatus) string {
	color := moodColor[s.Mood]
	if color == "" {
		color = moodColor[MoodDegraded]
	}
	detail := fmt.Sprintf("up %s", humanDur(s.Uptime))
	if len(s.Down) > 0 {
		detail = "down: " + strings.Join(s.Down, ", ")
	}
	return fmt.Sprintf(`<g font-family="sans-serif">`+
		`<rect x="26" y="252" width="128" height="58" rx="14" fill="#ffffff" fill-opacity="0.94" stroke="%s" stroke-width="1.5"/>`+
		`<circle cx="46" cy="274" r="6" fill="%s"/>`+
		`<text x="60" y="279" font-size="14" font-weight="700" fill="%s">%s</text>`+
		`<text x="38" y="298" font-size="10" fill="#78909c">%s</text>`+
		`</g>`, color, color, color, s.Mood, xmlEscape(detail))
}

// humanDur 运行时长（3d4h / 3h12m / 12m / 45s）。
func humanDur(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd%dh", int(d.Hours())/24, int(d.Hours())%24)
	case d >= time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	case d >= time.Minute:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
}

// xmlEscape 只转义依赖名与时长用到的字符（< & "），避免注入破坏 svg 结构。
func xmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
