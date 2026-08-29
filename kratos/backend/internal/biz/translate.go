package biz

// DeepL 翻译客户端 + 管理端翻译用例（标题/简介、批量章节）。
// key 取 env TRANSLATE_API_KEY，为空 → 180405；base 默认 api-free，TRANSLATE_BASE_URL 可覆盖。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	gormdb "gorm.io/plugin/dbresolver"

	"open-novel/backend/internal/data"
	"open-novel/backend/internal/pkg"
)

// translateLangMap 项目语言 → DeepL target_lang。
// ponytail: DeepL 新增语种时在此补充；bn（孟加拉语）DeepL 暂不支持，直接报 180404。
var translateLangMap = map[string]string{
	"zh-CN": "ZH-HANS", "en": "EN", "ja": "JA", "ko": "KO",
	"fr": "FR", "de": "DE", "es": "ES", "ru": "RU",
	"pt": "PT", "hi": "HI", "ar": "AR", "id": "ID",
}

type Translator struct {
	baseURL string
	key     string
	client  *http.Client
}

// NewTranslator baseURL 为空时用 DeepL free 端点；key 为空则翻译请求返回未配置错误。
func NewTranslator(baseURL, key string) *Translator {
	if baseURL == "" {
		baseURL = "https://api-free.deepl.com"
	}
	return &Translator{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		key:     key,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type deeplReq struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
}

type deeplResp struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// translate 一次调用翻译多条文本，返回与输入顺序一致的译文。
func (t *Translator) translate(ctx context.Context, lang string, texts []string) ([]string, error) {
	if t.key == "" {
		return nil, pkg.ErrTranslateNotCfg
	}
	target, ok := translateLangMap[lang]
	if !ok {
		return nil, pkg.ErrTranslateFailed
	}
	body, err := json.Marshal(deeplReq{Text: texts, TargetLang: target})
	if err != nil {
		return nil, pkg.ErrTranslateFailed
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/v2/translate", bytes.NewReader(body))
	if err != nil {
		return nil, pkg.ErrTranslateFailed
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+t.key)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, pkg.ErrTranslateFailed
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, pkg.ErrTranslateFailed
	}
	var out deeplResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, pkg.ErrTranslateFailed
	}
	if len(out.Translations) != len(texts) {
		return nil, pkg.ErrTranslateFailed
	}
	res := make([]string, len(out.Translations))
	for i, tr := range out.Translations {
		res[i] = tr.Text
	}
	return res, nil
}

// TranslateBook 翻译书籍标题+简介到目标语言（源取书籍原始语言 novel_book.lang，缺省 zh-CN），
// upsert 到 novel_book_translation 并失效详情缓存。
func (uc *AdminUsecase) TranslateBook(ctx context.Context, bookID uint64, lang string) (*data.BookTranslation, error) {
	var b data.Book
	if err := uc.db.WithContext(ctx).First(&b, bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrTargetNF
		}
		return nil, pkg.ErrAdminDB
	}
	src := b.Lang
	if src == "" {
		src = "zh-CN"
	}
	out, err := uc.tr.translate(ctx, lang, []string{b.Title, b.Summary})
	if err != nil {
		return nil, err
	}
	if err := upsertTranslation(uc.db, bookID, lang, out[0], out[1], ""); err != nil {
		return nil, pkg.ErrAdminDB
	}
	uc.cache.Del(ctx, fmt.Sprintf("book:%d:%s", bookID, lang))
	return &data.BookTranslation{BookID: bookID, Lang: lang, Title: out[0], Summary: out[1]}, nil
}

// ChaptersTranslateResult 批量章节翻译结果。
type ChaptersTranslateResult struct {
	Total          int
	Succeeded      int
	Failed         int
	FailedChapters []uint32 // 失败章节号（chapter_no）
}

// TranslateBookChapters 串行翻译书籍全部章节正文到目标语言，单章失败收集编号不中断。
// 源正文优先书籍原始语言，缺失回落 zh-CN；章节标题随请求发送（novel_chapter_content 无标题列，仅落正文）。
func (uc *AdminUsecase) TranslateBookChapters(ctx context.Context, bookID uint64, lang string) (*ChaptersTranslateResult, error) {
	var b data.Book
	if err := uc.db.WithContext(ctx).First(&b, bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkg.ErrTargetNF
		}
		return nil, pkg.ErrAdminDB
	}
	src := b.Lang
	if src == "" {
		src = "zh-CN"
	}
	var chapters []data.Chapter
	if err := uc.db.WithContext(ctx).Where("book_id = ? AND status = 1", bookID).
		Order("chapter_no").Find(&chapters).Error; err != nil {
		return nil, pkg.ErrAdminDB
	}
	// 批量取源正文：优先 srcLang，其次 zh-CN
	srcContent := map[uint64]string{}
	if len(chapters) > 0 {
		ids := make([]uint64, len(chapters))
		for i, ch := range chapters {
			ids[i] = ch.ID
		}
		var rows []data.ChapterContent
		if err := uc.db.WithContext(ctx).
			Where("chapter_id IN ? AND lang IN ?", ids, []string{src, "zh-CN"}).Find(&rows).Error; err != nil {
			return nil, pkg.ErrAdminDB
		}
		zhFallback := map[uint64]string{}
		for _, r := range rows {
			if r.Lang == src {
				srcContent[r.ChapterID] = r.Content
			} else {
				zhFallback[r.ChapterID] = r.Content
			}
		}
		for id, c := range zhFallback {
			if _, ok := srcContent[id]; !ok {
				srcContent[id] = c
			}
		}
	}
	res := &ChaptersTranslateResult{Total: len(chapters)}
	for _, ch := range chapters {
		if _, ok := srcContent[ch.ID]; !ok {
			res.Failed++
			res.FailedChapters = append(res.FailedChapters, ch.ChapterNo)
			continue
		}
		out, err := uc.tr.translate(ctx, lang, []string{ch.Title, srcContent[ch.ID]})
		if err != nil {
			res.Failed++
			res.FailedChapters = append(res.FailedChapters, ch.ChapterNo)
			continue
		}
		if err := uc.db.Clauses(gormdb.Write).Model(&data.ChapterContent{}).
			Where("chapter_id = ? AND lang = ?", ch.ID, lang).
			Assign(data.ChapterContent{Content: out[1]}).
			FirstOrCreate(&data.ChapterContent{ChapterID: ch.ID, Lang: lang, Content: out[1]}).Error; err != nil {
			res.Failed++
			res.FailedChapters = append(res.FailedChapters, ch.ChapterNo)
			continue
		}
		res.Succeeded++
	}
	// 失效章节正文缓存，避免 C 端读到旧值
	uc.cache.DelPattern(ctx, fmt.Sprintf("chapter:content:%d:*:%s", bookID, lang))
	return res, nil
}
