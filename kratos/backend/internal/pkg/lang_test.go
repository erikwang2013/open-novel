package pkg

import "testing"

func TestNormalizeLang(t *testing.T) {
	cases := map[string]string{
		"": "", "zh": "zh-CN", "zh-CN": "zh-CN", "ZH-TW": "zh-CN",
		"en": "en", "en-US": "en", "ja": "ja", "ja-JP": "ja",
		"ko": "ko", "ko-KR": "ko", "fr": "fr",
	}
	for in, want := range cases {
		if got := NormalizeLang(in); got != want {
			t.Errorf("NormalizeLang(%q) = %q, want %q", in, got, want)
		}
	}
}
