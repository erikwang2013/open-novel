package cdn

// 阿里云 adapter（§五 5.3）：RPC form + HMAC-SHA1 签名；ObjectPath 换行分隔；≤1000/批；50qps；
// 每日 10000 URL 预警（8000 起 warn）。签名向量由独立参考实现（Python）计算，防回归。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestAliyunSignVector 固定输入 → 固定签名（独立参考实现计算）：
// AccessKeyId=testid, Action=DescribeCdnService, Format=JSON,
// SignatureMethod=HMAC-SHA1, SignatureNonce=4c33d64d-ee48-4b13-bb81-0a3b0a4f5b7b,
// SignatureVersion=1.0, Timestamp=2016-02-23T12:46:24Z, Version=2014-11-11,
// SecretKey=testsecret。
func TestAliyunSignVector(t *testing.T) {
	params := map[string]string{
		"AccessKeyId":      "testid",
		"Action":           "DescribeCdnService",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   "4c33d64d-ee48-4b13-bb81-0a3b0a4f5b7b",
		"SignatureVersion": "1.0",
		"Timestamp":        "2016-02-23T12:46:24Z",
		"Version":          "2014-11-11",
	}
	got := aliyunSign("testsecret", params)
	want := "1aHx5fy2R2UfxfBMfMvIv624x50="
	if got != want {
		t.Fatalf("signature vector mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestNewAliyunValidate(t *testing.T) {
	if _, err := NewAliyun(map[string]any{"access_key_id": "a"}); err == nil {
		t.Fatal("missing access_key_secret must error")
	}
}

// TestAliyunPurgeForm mock 断言：POST 表单含 ObjectPath（\n 分隔）与签名参数。
func TestAliyunPurgeForm(t *testing.T) {
	var got url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got = r.PostForm
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	p, err := NewAliyun(map[string]any{"access_key_id": "ak", "access_key_secret": "sk",
		"endpoint": ts.URL})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "aliyun" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/1?lang=zh-CN", "chapter/2?lang=en"}); err != nil {
		t.Fatal(err)
	}
	if got.Get("Action") != "RefreshObjectCaches" || got.Get("ObjectType") != "File" {
		t.Fatalf("action mismatch: %v", got)
	}
	if got.Get("ObjectPath") != "chapter/1?lang=zh-CN\nchapter/2?lang=en" {
		t.Fatalf("object path: %q", got.Get("ObjectPath"))
	}
	if got.Get("Signature") == "" || got.Get("Timestamp") == "" || got.Get("SignatureNonce") == "" {
		t.Fatalf("missing signature params: %v", got)
	}
}

// TestAliyunBatch1000 超 1000 个 key 切多批。
func TestAliyunBatch1000(t *testing.T) {
	var batches int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches++
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	p, _ := NewAliyun(map[string]any{"access_key_id": "a", "access_key_secret": "s",
		"endpoint": ts.URL, "batch_size": float64(100)})
	keys := make([]string, 250)
	for i := range keys {
		keys[i] = "k" + strings.Repeat("y", i)
	}
	if err := p.Purge(t.Context(), keys); err != nil {
		t.Fatal(err)
	}
	if batches != 3 {
		t.Fatalf("want 3 batches, got %d", batches)
	}
}

// TestAliyunRetriable 5xx → 可重试错误。
func TestAliyunRetriable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()
	p, _ := NewAliyun(map[string]any{"access_key_id": "a", "access_key_secret": "s", "endpoint": ts.URL})
	err := p.Purge(t.Context(), []string{"k"})
	if !httpRetriable(err) {
		t.Fatalf("want retriable, got %v", err)
	}
}

// TestAliyunDailyCounter 日累计计数到达阈值记 warn（当天仅一次）。
func TestAliyunDailyCounter(t *testing.T) {
	var warns int
	c := newDailyCounter(8000, func(string) { warns++ })
	p := &aliyunProvider{counter: c}
	p.counter.Add(7990)
	p.counter.Add(10)
	if warns != 1 {
		t.Fatalf("want 1 warn, got %d", warns)
	}
	p.counter.Add(5)
	if warns != 1 {
		t.Fatalf("want still 1 warn, got %d", warns)
	}
	_ = context.Background()
}
