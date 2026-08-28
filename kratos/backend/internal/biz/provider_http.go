package biz

// 渠道共用轻量 REST 直调与配置读取（照 provider_np.go 的 do 模式）：纯 net/http，不引入第三方库。

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
)

// cfgStr 取解密后的渠道配置字符串值（admin 配置经 JSON 序列化后均为字符串）。
func cfgStr(cfg map[string]any, key string) string {
	s, _ := cfg[key].(string)
	return s
}

// basicAuthHeader Basic Auth 凭证（Razorpay 下单、PayPal OAuth 用）。
func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

// doJSON 发送 JSON 请求并解析响应；HTTP ≥400 返回截断的错误体；body 为 nil 时无请求体。
func doJSON(ctx context.Context, c *http.Client, method, url string, headers map[string]string, body, out any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		s := string(b)
		if len(s) > 200 {
			s = s[:200]
		}
		return fmt.Errorf("%s %s: %s", method, url, s)
	}
	if out != nil && len(b) > 0 {
		return json.Unmarshal(b, out)
	}
	return nil
}

// moneyToCents 解析渠道十进制金额字符串为整数分（PayPal value 等）。
func moneyToCents(s string) int64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(math.Round(f * 100))
}
