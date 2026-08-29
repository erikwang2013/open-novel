package data

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// 测试需本地 Redis（与 internal/service 现有测试同款基础设施：127.0.0.1:6380）。
func newTestCache(t *testing.T) *Cache {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6380"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis 不可用，跳过: %v", err)
	}
	c, err := NewCache(rdb)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

// TestCacheL1: L1 命中不回源、Set/Del/DelPattern 双删、Get 未命中回源回填。
func TestCacheL1(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	key := "l1:test:" + time.Now().Format("150405.000")

	c.Set(ctx, key, "v1", time.Minute)
	if v, ok := c.Get(ctx, key); !ok || v != "v1" {
		t.Fatalf("Set 后应命中 v1，got %q ok=%v", v, ok)
	}
	// L1 命中不回源：删掉 Redis 后仍应中 L1
	c.rdb.Del(ctx, key)
	if v, ok := c.Get(ctx, key); !ok || v != "v1" {
		t.Fatalf("L1 应命中 v1，got %q ok=%v", v, ok)
	}
	// Del 双删
	c.Del(ctx, key)
	if _, ok := c.Get(ctx, key); ok {
		t.Fatal("Del 后不应命中")
	}
	// Get 未命中回源并回填 L1
	c.rdb.Set(ctx, key, "v2", time.Minute)
	if v, ok := c.Get(ctx, key); !ok || v != "v2" {
		t.Fatalf("回源应得 v2，got %q ok=%v", v, ok)
	}
	c.rdb.Del(ctx, key)
	if v, ok := c.Get(ctx, key); !ok || v != "v2" {
		t.Fatalf("回填后 L1 应命中 v2，got %q ok=%v", v, ok)
	}
	// DelPattern 双删
	c.Set(ctx, key, "v3", time.Minute)
	c.DelPattern(ctx, "l1:test:*")
	if _, ok := c.Get(ctx, key); ok {
		t.Fatal("DelPattern 后不应命中")
	}
}

// TestCacheL1Expiry: 业务 TTL 短于 l1TTL 时取短者，过期视为未命中。
func TestCacheL1Expiry(t *testing.T) {
	c := newTestCache(t)
	ctx := context.Background()
	key := "l1:exp:" + time.Now().Format("150405.000")

	c.Set(ctx, key, "x", 20*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if _, ok := c.Get(ctx, key); ok {
		t.Fatal("过期条目不应命中")
	}
}
