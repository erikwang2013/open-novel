package cdn

// CloudFront adapter（§五 5.2）：aws-sdk-go-v2 CreateInvalidation。
// invalidation path 必须去 query + 补前导 /（官方要求，带 query 直接 400）。
// CallerReference = 批次内 key 排序后 sha256 前 16 hex（唯一、幂等）。
// 测试经 BaseEndpoint 指向 mock server，静态假凭据（mock 不验签）。

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCfPath(t *testing.T) {
	cases := map[string]string{
		"chapter/123?lang=zh-CN": "/chapter/123",
		"chapter/123":            "/chapter/123",
		"/book/9":                "/book/9",
	}
	for in, want := range cases {
		if got := cfPath(in); got != want {
			t.Fatalf("cfPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCfCallerRefStable(t *testing.T) {
	a := cfCallerRef([]string{"/b", "/a"})
	b := cfCallerRef([]string{"/a", "/b"})
	if a != b {
		t.Fatalf("caller ref must be order-independent: %s vs %s", a, b)
	}
	if len(a) != 16 {
		t.Fatalf("caller ref must be 16 hex chars, got %d: %s", len(a), a)
	}
}

func TestNewCloudFrontValidate(t *testing.T) {
	if _, err := NewCloudFront(map[string]any{"access_key_id": "a", "secret_access_key": "s"}); err == nil {
		t.Fatal("missing distribution_id must error")
	}
}

// TestCloudFrontPurgeXML mock 断言：POST /distribution/{id}/invalidation，XML body 无 query。
func TestCloudFrontPurgeXML(t *testing.T) {
	got := struct {
		Method, Path string
		Body         string
	}{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Method = r.Method
		got.Path = r.URL.Path
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		got.Body = string(buf[:n])
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<Invalidation xmlns="http://cloudfront.amazonaws.com/doc/2020-05-31/">
  <Id>I1</Id><Status>InProgress</Status>
  <CreateTime>2026-08-29T00:00:00.000Z</CreateTime>
  <InvalidationBatch>
    <Paths><Quantity>1</Quantity><Items>/chapter/123</Items></Paths>
    <CallerReference>abcdef1234567890</CallerReference>
  </InvalidationBatch>
</Invalidation>`))
	}))
	defer ts.Close()

	p, err := NewCloudFront(map[string]any{
		"access_key_id": "AK", "secret_access_key": "SK", "distribution_id": "D1",
		"base_endpoint": ts.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "cloudfront" {
		t.Fatalf("name: %s", p.Name())
	}
	if err := p.Purge(t.Context(), []string{"chapter/123?lang=zh-CN"}); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost || !strings.Contains(got.Path, "/distribution/D1/invalidation") {
		t.Fatalf("request mismatch: %+v", got)
	}
	if strings.Contains(got.Body, "?lang") {
		t.Fatalf("invalidation path must strip query: %s", got.Body)
	}
	if !strings.Contains(got.Body, "/chapter/123") {
		t.Fatalf("body missing path: %s", got.Body)
	}
}
