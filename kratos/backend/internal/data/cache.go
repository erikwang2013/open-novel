package data

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

// 缓存策略（规划文档 §五，从 svc1 移植）：TTL 5min±30s 抖动、空值缓存 60s、
// SETNX 单飞防击穿。

const EmptyMarker = "\x00empty"

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache { return &Cache{rdb: rdb} }

func ttlJitter() time.Duration {
	n, _ := rand.Int(rand.Reader, big.NewInt(61)) // 0..60
	return 5*time.Minute - 30*time.Second + time.Duration(n.Int64())*time.Second
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	v, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return v, true
}

func (c *Cache) Set(ctx context.Context, key, val string, ttl time.Duration) {
	c.rdb.Set(ctx, key, val, ttl)
}

func (c *Cache) Del(ctx context.Context, keys ...string) {
	c.rdb.Del(ctx, keys...)
}

// DelPattern SCAN 匹配后批量删除（非阻塞；超大 key 空间可改游标分批）。
func (c *Cache) DelPattern(ctx context.Context, pattern string) {
	var keys []string
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		c.rdb.Del(ctx, keys...)
	}
}

// GetOrLoad：缓存 →（SETNX 单飞）→ loader → 回填；空结果缓存 60s 防穿透。
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
