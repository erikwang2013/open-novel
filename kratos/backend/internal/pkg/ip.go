package pkg

import (
	"net"
	"strings"
)

// ClientIP 取客户端 IP：XFF 取最后一跳（可信代理追加的那项，客户端伪造的前置项不参与），
// XFF 为空时从 remote 去端口；都为空返回空串。
func ClientIP(xff, remote string) string {
	if xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(strings.TrimSpace(remote)); err == nil {
		return host
	}
	return strings.TrimSpace(remote)
}
