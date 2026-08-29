package middleware

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"open-novel/backend/internal/pkg"
)

func TestCheckAttackBlockHigh(t *testing.T) {
	r := httptestReq("http://x/api/search?q=../../etc/passwd", "10.0.0.1")
	if err := checkAttack(r); err != pkg.ErrPermission {
		t.Fatalf("want ErrPermission, got %v", err)
	}
}

func TestCheckAttackAllowBenign(t *testing.T) {
	r := httptestReq("http://x/api/search?q=hello", "10.0.0.2")
	if err := checkAttack(r); err != nil {
		t.Fatalf("benign request should pass, got %v", err)
	}
}

// 5 次高危攻击后 IP 被拉黑，后续普通请求也被拒绝。
func TestCheckAttackBansIP(t *testing.T) {
	ip := fmt.Sprintf("10.1.%d.%d", time.Now().UnixNano()%250, time.Now().UnixNano()%250)
	attack := httptestReq("http://x/api/search?q=../../etc/passwd", ip)
	for i := 0; i < 5; i++ {
		if err := checkAttack(attack); err != pkg.ErrPermission {
			t.Fatalf("attack %d should be blocked, got %v", i, err)
		}
	}
	benign := httptestReq("http://x/api/search?q=hello", ip)
	if err := checkAttack(benign); err != pkg.ErrPermission {
		t.Fatalf("banned IP should be blocked on benign request, got %v", err)
	}
}

func httptestReq(url, ip string) *http.Request {
	r, _ := http.NewRequest("GET", url, nil)
	r.Header.Set("X-Forwarded-For", ip)
	return r
}
