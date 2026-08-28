package pkg

import "strings"

// NormalizeLang 客户端语言别名归一化（docs/api.md：zh|en|ja；客户端还可能带 region）。
// zh*→zh-CN、en*→en、ja*→ja、ko*→ko；未知语言原样返回。
func NormalizeLang(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	switch {
	case l == "":
		return ""
	case strings.HasPrefix(l, "zh"):
		return "zh-CN"
	case strings.HasPrefix(l, "en"):
		return "en"
	case strings.HasPrefix(l, "ja"):
		return "ja"
	case strings.HasPrefix(l, "ko"):
		return "ko"
	}
	return lang
}
