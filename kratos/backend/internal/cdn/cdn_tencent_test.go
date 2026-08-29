package cdn

// 腾讯云 adapter（§五 5.4）：JSON POST + TC3-HMAC-SHA256；≤1000/批；20qps；每日 10000 预警。
// 签名向量由独立参考实现（Python）计算，防回归。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// TestTencentSignVector 固定输入 → 固定签名（独立参考实现计算）：
// SecretId=AKIDEXAMPLE, SecretKey=Gu5t9xGARNpq86cd98joQYCN3Cozk1qA,
// ts=1788004800（=2026-08-29 UTC）, host=cdn.tencentcloudapi.com,
// body={"Urls":["chapter/123?lang=zh-CN"]}。
func TestTencentSignVector(t *testing.T) {
	body := []byte(`{"Urls":["chapter/123?lang=zh-CN"]}`)
	got := tencentSign("AKIDEXAMPLE", "Gu5t9xGARNpq86cd98joQYCN3Cozk1qA",
		"2026-08-29", 1788004800, body, "cdn.tencentcloudapi.com")
	want := "TC3-HMAC-SHA256 Credential=AKIDEXAMPLE/2026-08-29/cdn/tc3_request, " +
		"SignedHeaders=content-type;host;x-tc-action, " +
		"Signature=f95b1da7d4ff9fa8bf197765b41f091ee58e3a312a5372174f1fe4575f294a09"
	if got != want {
		t.Fatalf("TC3 vector mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestNewTencentValidate(t *testing.T) {
	if _, err := NewTencent(map[string]any{"secret_id": "a"}); err == nil {
		t.Fatal("missing secret_key must error")
	}
}

// TestTencentPurgeHeaders mock 断言：TC3 头 + JSON Urls。
func TestTencentPurgeHeaders(t *testing.T) {
	var got struct {
		Action, Version, Timestamp, Auth, CT string
		Urls                                 []string `json:"Urls"`
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Action = r.Header.Get("X-TC-Action")
		got.Version = r.Header.Get("X-TC-Version")
		got.Timestamp = r.Header.Get("X-TC-Timestamp")
		got.Auth = r.Header.Get("Authorization")
		got.CT = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	p, err := NewTencent(map[string]any{"secret_id": "sid", "secret_key": "skey",
		"endpoint": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "tencent" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Action != "PurgeUrlsCache" || got.Version != "2018-06-06" {
		t.Fatalf("headers mismatch: %+v", got)
	}
	if got.CT != "application/json; charset=utf-8" {
		t.Fatalf("content-type must match SignedHeaders: %q", got.CT)
	}
	if !strings.HasPrefix(got.Auth, "TC3-HMAC-SHA256 Credential=sid/") ||
		!strings.Contains(got.Auth, "SignedHeaders=content-type;host;x-tc-action") {
		t.Fatalf("auth header: %q", got.Auth)
	}
	if _, err := strconv.ParseInt(got.Timestamp, 10, 64); err != nil {
		t.Fatalf("bad timestamp: %q", got.Timestamp)
	}
	if len(got.Urls) != 2 || got.Urls[0] != "chapter/1?lang=zh-CN" {
		t.Fatalf("urls: %v", got.Urls)
	}
}

// TestTencentRetriable 429 → 可重试。
func TestTencentRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()
	p, _ := NewTencent(map[string]any{"secret_id": "s", "secret_key": "k", "endpoint": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}
