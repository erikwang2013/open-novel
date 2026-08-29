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

// Providers 当前厂商列表（测试/诊断用）。
func (m *Manager) Providers() []Provider { return m.providers }

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
// ponytail: 固定 1s 退避 + 单次重试；重试不感知 ctx 且整批重跑，best-effort 语义可容忍（旧内容最坏滞留 1h）；可靠失效需队列，暂不做。
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
