package pkg

import "testing"

func TestClientIP(t *testing.T) {
	// 取最后一跳：伪造首个 XFF 不影响结果
	if got := ClientIP("6.6.6.6, 2.2.2.2", "10.0.0.1:1234"); got != "2.2.2.2" {
		t.Fatalf("want last XFF hop, got %q", got)
	}
	if a, b := ClientIP("1.1.1.1, 2.2.2.2", "10.0.0.1:1"), ClientIP("6.6.6.6, 2.2.2.2", "10.0.0.1:1"); a != b {
		t.Fatalf("changing first XFF must not change key: %q vs %q", a, b)
	}
	// 空 XFF → remote 去端口
	if got := ClientIP("", "10.0.0.1:1234"); got != "10.0.0.1" {
		t.Fatalf("want remote host, got %q", got)
	}
	if got := ClientIP("", "10.0.0.1"); got != "10.0.0.1" {
		t.Fatalf("want bare remote host, got %q", got)
	}
	// 都为空 → 空
	if got := ClientIP("", ""); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}
