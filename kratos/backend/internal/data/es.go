package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"open-novel/backend/internal/pkg"
)

// OpenSearch 客户端（从 svc1 移植）：多语言 analyzer、refresh=wait_for 保证同步即一致。

const indexName = "novel_books"

// Analyzers: multi_lang_analyzer 处理 zh/en（standard tokenizer 按字符匹配 CJK），
// ja_analyzer 用 kuromoji 做日文分词。
const indexSettings = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "multi_lang_analyzer": {
          "tokenizer": "standard",
          "filter": ["lowercase", "icu_normalizer", "kuromoji_stemmer"]
        },
        "ja_analyzer": {
          "tokenizer": "kuromoji_tokenizer",
          "filter": ["lowercase", "icu_normalizer", "kuromoji_stemmer"]
        },
        "ko_analyzer": {
          "tokenizer": "nori_tokenizer",
          "filter": ["lowercase", "icu_normalizer"]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "book_id": {"type": "long"},
      "lang": {"type": "keyword"},
      "status": {"type": "integer"},
      "hot": {"type": "long"},
      "created_at": {"type": "date"},
      "title_zh": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "title_en": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "title_ja": {"type": "text", "analyzer": "ja_analyzer"},
      "title_ko": {"type": "text", "analyzer": "ko_analyzer"},
      "summary_zh": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "summary_en": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "summary_ja": {"type": "text", "analyzer": "ja_analyzer"},
      "summary_ko": {"type": "text", "analyzer": "ko_analyzer"},
      "author_zh": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "author_en": {"type": "text", "analyzer": "multi_lang_analyzer"},
      "author_ja": {"type": "text", "analyzer": "ja_analyzer"},
      "author_ko": {"type": "text", "analyzer": "ko_analyzer"}
    }
  }
}`

// SearchDoc 是搜索文档，每书一文档（book_id 为文档 id），四语字段由书籍服务同步。
type SearchDoc struct {
	BookID    uint64 `json:"book_id"`
	Lang      string `json:"lang"`
	Status    int    `json:"status"`
	Hot       int64  `json:"hot"`
	CreatedAt string `json:"created_at"`
	TitleZh   string `json:"title_zh"`
	TitleEn   string `json:"title_en"`
	TitleJa   string `json:"title_ja"`
	TitleKo   string `json:"title_ko"`
	SummaryZh string `json:"summary_zh"`
	SummaryEn string `json:"summary_en"`
	SummaryJa string `json:"summary_ja"`
	SummaryKo string `json:"summary_ko"`
	AuthorZh  string `json:"author_zh"`
	AuthorEn  string `json:"author_en"`
	AuthorJa  string `json:"author_ja"`
	AuthorKo  string `json:"author_ko"`
}

type ES struct {
	base string
	http *http.Client
}

func NewES(addr string) *ES {
	if addr == "" {
		addr = "http://127.0.0.1:9200"
	}
	return &ES{base: addr, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *ES) do(ctx context.Context, method, path string, body any, out any) error {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
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
		return fmt.Errorf("es %s %s: %s", method, path, s)
	}
	if out != nil && len(b) > 0 {
		return json.Unmarshal(b, out)
	}
	return nil
}

// EnsureIndex 幂等建索引（启动时调用）。
func (c *ES) EnsureIndex(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.base+"/"+indexName, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return c.do(ctx, http.MethodPut, "/"+indexName, json.RawMessage(indexSettings), nil)
}

// langFields 映射请求 lang 到检索字段 + analyzer；未知语言回落 zh 字段。
func langFields(lang string) (fields []string, analyzer string) {
	switch pkg.NormalizeLang(lang) {
	case "ja":
		return []string{"title_ja", "summary_ja", "author_ja"}, "ja_analyzer"
	case "ko":
		return []string{"title_ko", "summary_ko", "author_ko"}, "ko_analyzer"
	case "en":
		return []string{"title_en", "summary_en", "author_en"}, "multi_lang_analyzer"
	default:
		return []string{"title_zh", "summary_zh", "author_zh"}, "multi_lang_analyzer"
	}
}

type esHit struct {
	Source SearchDoc `json:"_source"`
}

type esSearchResp struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []esHit `json:"hits"`
	} `json:"hits"`
}

// Search 对 lang 对应字段跑 multi_match。
func (c *ES) Search(ctx context.Context, q, lang string, from, size int) ([]SearchDoc, int64, error) {
	fields, analyzer := langFields(lang)
	body := map[string]any{
		"from": from, "size": size,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []any{map[string]any{"multi_match": map[string]any{"query": q, "fields": fields, "analyzer": analyzer}}},
				// 排除显式下架（status=0）；未同步 status 的旧文档视为上架。
				// ponytail: 存量数据未回填 status，全量回填后可换 term(status:1)
				"must_not": []any{map[string]any{"term": map[string]any{"status": 0}}},
			},
		},
	}
	var resp esSearchResp
	if err := c.do(ctx, http.MethodPost, "/"+indexName+"/_search", body, &resp); err != nil {
		return nil, 0, err
	}
	docs := make([]SearchDoc, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		docs = append(docs, h.Source)
	}
	return docs, resp.Hits.Total.Value, nil
}

// Hot 按 hot 分倒序取榜。
func (c *ES) Hot(ctx context.Context, size int) ([]SearchDoc, error) {
	body := map[string]any{
		"size": size,
		"query": map[string]any{"match_all": map[string]any{}},
		"sort": []any{map[string]any{"hot": map[string]any{"order": "desc"}}},
	}
	var resp esSearchResp
	if err := c.do(ctx, http.MethodPost, "/"+indexName+"/_search", body, &resp); err != nil {
		return nil, err
	}
	docs := make([]SearchDoc, 0, len(resp.Hits.Hits))
	for _, h := range resp.Hits.Hits {
		docs = append(docs, h.Source)
	}
	return docs, nil
}

// Upsert 写入/替换文档；refresh=wait_for 保证同步→搜索一致。
func (c *ES) Upsert(ctx context.Context, doc SearchDoc) error {
	path := fmt.Sprintf("/%s/_doc/%d?refresh=wait_for", indexName, doc.BookID)
	return c.do(ctx, http.MethodPut, path, doc, nil)
}

func (c *ES) Delete(ctx context.Context, bookID uint64) error {
	path := fmt.Sprintf("/%s/_doc/%d?refresh=wait_for", indexName, bookID)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func clientIPFrom(xff, remoteAddr string) string {
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}
	return remoteAddr
}
