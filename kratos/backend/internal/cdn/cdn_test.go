package cdn

// 核心编排测试：去重 / 批量切分 / 并发广播 / 429 重试一次 / 非重试错误不重试 / 令牌桶 / 每日计数。

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type fakeProvider struct {
	name string
	hits *int32
	err  error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Purge(ctx context.Context, keys []string) error {
	atomic.AddInt32(f.hits, 1)
	return f.err
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"a", "b", "a", "c", "b"})
	if !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Fatalf("want [a b c], got %v", got)
	}
}

func TestSplit(t *testing.T) {
	got := Split([]string{"1", "2", "3", "4", "5"}, 2)
	if len(got) != 3 || len(got[0]) != 2 || len(got[2]) != 1 {
		t.Fatalf("split mismatch: %v", got)
	}
	if got := Split(nil, 0); got != nil {
		t.Fatalf("max<=0 must not panic, got %v", got)
	}
}

// TestManagerBroadcast 两个 provider 各收到同一批 key。
func TestManagerBroadcast(t *testing.T) {
	var h1, h2 int32
	m := NewManager([]Provider{&fakeProvider{name: "p1", hits: &h1}, &fakeProvider{name: "p2", hits: &h2}})
	m.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/1?lang=zh-CN", "chapter/1?lang=en"})
	if atomic.LoadInt32(&h1) != 1 || atomic.LoadInt32(&h2) != 1 {
		t.Fatalf("want 1 purge call per provider, got %d/%d", h1, h2)
	}
}

// TestManagerRetryOnce retriableError（429）重试一次（~1s 退避），重试后仍失败即放弃；
// 无错误调用不重试不 sleep。
func TestManagerRetryOnce(t *testing.T) {
	var calls int32
	p := &fakeProvider{name: "t", hits: &calls, err: &retriableError{status: 429}}
	m := NewManager([]Provider{p})
	start := time.Now()
	m.Purge(t.Context(), []string{"k"})
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("want 2 calls (1 retry), got %d", calls)
	}
	if time.Since(start) < 900*time.Millisecond {
		t.Fatal("retry backoff must sleep ~1s")
	}
	// 无错误调用：1 次、无重试、无 sleep
	ok := &fakeProvider{name: "ok", hits: new(int32)}
	m2 := NewManager([]Provider{ok})
	m2.Purge(t.Context(), []string{"k"})
	if atomic.LoadInt32(ok.hits) != 1 {
		t.Fatalf("want 1 hit, got %d", atomic.LoadInt32(ok.hits))
	}
}

// TestManagerNonRetriableNoRetry 非重试错误只记日志，不 sleep。
func TestManagerNonRetriableNoRetry(t *testing.T) {
	p := &fakeProvider{name: "x", err: &retriableError{status: 400}, hits: new(int32)}
	m := NewManager([]Provider{p})
	start := time.Now()
	m.Purge(t.Context(), []string{"k"})
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("non-retriable must not sleep")
	}
	if atomic.LoadInt32(p.hits) != 1 {
		t.Fatalf("want 1 call, got %d", atomic.LoadInt32(p.hits))
	}
}

func TestTokenBucketWait(t *testing.T) {
	b := newTokenBucket(100) // 100 qps
	start := time.Now()
	for i := 0; i < 10; i++ {
		if err := b.Wait(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("10 tokens at 100qps must be fast, took %v", time.Since(start))
	}
}

func TestDailyCounterWarnOnce(t *testing.T) {
	var warns int
	c := newDailyCounter(8000, func(string) { warns++ })
	c.Add(7999)
	if warns != 0 {
		t.Fatal("must not warn below threshold")
	}
	c.Add(1)
	if warns != 1 {
		t.Fatal("must warn at threshold")
	}
	c.Add(100)
	if warns != 1 {
		t.Fatal("must warn only once per day")
	}
}
