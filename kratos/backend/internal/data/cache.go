package data

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"path"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
)

// 缓存策略（规划文档 §五，从 svc1 移植）：TTL 5min±30s 抖动、空值缓存 60s、
// SETNX 单飞防击穿。L1 进程内二级缓存（§7.3 ristretto）：128MB、30s 短 TTL
// 兜底一致性。
//
// ponytail: 本项按用户指示实施；原计划触发条件为压测证明 Redis 瓶颈
// （>10k QPS 热点集中），未达标时收益有限。

const EmptyMarker = "\x00empty"

// l1TTL L1 条目 TTL；业务 TTL 更短时取短者，避免 L1 活得比 Redis 久。
const l1TTL = 30 * time.Second

// l1Entry ristretto 无 TTL，value 自带过期时间。
type l1Entry struct {
	v    string
	dead time.Time
}

type Cache struct {
	rdb *redis.Client
	l1  *ristretto.Cache
	// expires L1 key 索引：仅用于 DelPattern 遍历（ristretto 无迭代 API）。
	// ponytail: map 双写 + 全表遍历，key 空间大时可换带 TTL 桶的索引结构。
	mu      sync.RWMutex
	expires map[string]time.Time
}

func NewCache(rdb *redis.Client) (*Cache, error) {
	l1, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,       // 约 4x 预估条目数
		MaxCost:     128 << 20, // 128MB
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("init ristretto: %w", err)
	}
	return &Cache{rdb: rdb, l1: l1, expires: make(map[string]time.Time)}, nil
}

// Close 释放 ristretto 后台 goroutine（进程退出前调用）。
func (c *Cache) Close() { c.l1.Close() }

func ttlJitter() time.Duration {
	n, _ := rand.Int(rand.Reader, big.NewInt(61)) // 0..60
	return 5*time.Minute - 30*time.Second + time.Duration(n.Int64())*time.Second
}

func (c *Cache) l1Set(key, val string, ttl time.Duration) {
	if ttl <= 0 || ttl > l1TTL {
		ttl = l1TTL
	}
	dead := time.Now().Add(ttl)
	c.mu.Lock()
	c.expires[key] = dead
	c.mu.Unlock()
	c.l1.Set(key, l1Entry{v: val, dead: dead}, int64(len(val)+32))
	// 新 key 的 Set 是异步生效的，Wait 保证立刻可见（微秒级，远小于随后的
	// Redis 网络往返），并保持 Del→Set 的先后顺序。
	c.l1.Wait()
}

func (c *Cache) l1Get(key string) (string, bool) {
	v, ok := c.l1.Get(key)
	if !ok {
		return "", false
	}
	e := v.(l1Entry)
	if time.Now().After(e.dead) {
		c.l1Del(key)
		return "", false
	}
	return e.v, true
}

func (c *Cache) l1Del(key string) {
	c.l1.Del(key)
	c.mu.Lock()
	delete(c.expires, key)
	c.mu.Unlock()
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	if v, ok := c.l1Get(key); ok {
		return v, true
	}
	v, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	c.l1Set(key, v, l1TTL) // 未命中回源并回填 L1
	return v, true
}

func (c *Cache) Set(ctx context.Context, key, val string, ttl time.Duration) {
	c.l1Set(key, val, ttl)
	c.rdb.Set(ctx, key, val, ttl)
}

func (c *Cache) Del(ctx context.Context, keys ...string) {
	for _, k := range keys {
		c.l1Del(k)
	}
	c.rdb.Del(ctx, keys...)
}

// DelPattern Redis SCAN 匹配后批量删除（非阻塞；超大 key 空间可改游标分批）；
// L1 侧遍历索引删除，并顺带清理已过期条目。
// ponytail: path.Match 近似 Redis glob（* ? []），转义/字符类等细微差异以
// Redis 为准，L1 残留在 30s TTL 内自然失效。
func (c *Cache) DelPattern(ctx context.Context, pattern string) {
	var keys []string
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		c.rdb.Del(ctx, keys...)
	}
	now := time.Now()
	c.mu.Lock()
	for k, dead := range c.expires {
		if now.After(dead) || matchPattern(pattern, k) {
			c.l1.Del(k)
			delete(c.expires, k)
		}
	}
	c.mu.Unlock()
}

func matchPattern(pattern, key string) bool {
	ok, _ := path.Match(pattern, key)
	return ok
}

// GetOrLoad：缓存 →（SETNX 单飞）→ loader → 回填；空结果缓存 60s 防穿透。
// 内部走 c.Get/c.Set，天然经过 L1，无需额外逻辑。
func (c *Cache) GetOrLoad(ctx context.Context, key string, loader func() (string, error)) (string, error) {
	if v, ok := c.Get(ctx, key); ok {
		return v, nil
	}
	lockKey := key + ":lock"
	got, err := c.rdb.SetNX(ctx, lockKey, "1", 5*time.Second).Result()
	if err != nil {
		return "", err
	}
	if got {
		defer c.rdb.Del(ctx, lockKey)
		v, err := loader()
		if err != nil {
			return "", err
		}
		if v == "" {
			c.Set(ctx, key, EmptyMarker, 60*time.Second)
			return EmptyMarker, nil
		}
		c.Set(ctx, key, v, ttlJitter())
		return v, nil
	}
	// 抢锁失败：等赢家回填后读缓存
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if v, ok := c.Get(ctx, key); ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("cache load timeout: %s", key)
}
