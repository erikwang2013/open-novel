package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterPruneSweep(t *testing.T) {
	l := &rateLimiter{seen: map[string][]time.Time{}, limit: 2}
	ip := "1.2.3.4"
	old := time.Now().Add(-2 * time.Minute)
	l.seen[ip] = []time.Time{old, old}

	l.prune(ip, time.Now())
	if _, ok := l.seen[ip]; ok {
		t.Fatal("stale entry should be deleted")
	}

	if !l.allow(ip) || !l.allow(ip) || l.allow(ip) {
		t.Fatal("limit 2 should allow 2 then block")
	}
	if len(l.seen[ip]) != 2 {
		t.Fatalf("want 2 hits, got %d", len(l.seen[ip]))
	}

	l.seen["dead1"] = []time.Time{old}
	l.seen["dead2"] = []time.Time{old}
	l.sweep(time.Now())
	if len(l.seen) != 1 {
		t.Fatalf("sweep should keep only the live IP, got %d entries", len(l.seen))
	}
}
