package pkg

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClassifyMood(t *testing.T) {
	cases := []struct {
		db, cache, search bool
		want              Mood
	}{
		{true, true, true, MoodHappy},
		{false, false, false, MoodDown},
		{false, true, true, MoodDegraded},
		{true, false, true, MoodDegraded},
		{true, true, false, MoodDegraded},
		{false, false, true, MoodDegraded},
	}
	for _, c := range cases {
		if got := classifyMood(c.db, c.cache, c.search); got != c.want {
			t.Errorf("classifyMood(%v,%v,%v) = %q, want %q", c.db, c.cache, c.search, got, c.want)
		}
	}
}

func TestRenderKeepsSVGWellFormed(t *testing.T) {
	cases := []PetStatus{
		{Mood: MoodHappy, Uptime: 3*time.Hour + 12*time.Minute},
		{Mood: MoodDegraded, Down: []string{"redis"}, Uptime: 90 * time.Second},
		{Mood: MoodDown, Down: []string{"mysql", "redis", "opensearch"}},
	}
	for _, s := range cases {
		out := Render(s)
		if err := xml.Unmarshal([]byte(out), new(struct {
			XMLName xml.Name `xml:"svg"`
		})); err != nil {
			t.Fatalf("mood %q: rendered svg is not well-formed: %v", s.Mood, err)
		}
		if strings.Contains(out, "MOOD_FACE_START") || strings.Contains(out, "MOOD_BADGE_START") {
			t.Errorf("mood %q: marker not replaced", s.Mood)
		}
		if !strings.Contains(out, ">"+string(s.Mood)+"<") {
			t.Errorf("mood %q: status label missing from badge", s.Mood)
		}
		if !strings.Contains(out, "https://erik.xyz") { // 出图必须带水印
			t.Errorf("mood %q: watermark missing", s.Mood)
		}
	}
	// 表情注入量：happy 不注入，degraded 只垂右眼，down 双眼闭合 + 飘 z
	if got := faceSnippet(MoodHappy); got != "" {
		t.Errorf("happy: expected no overlay, got %q", got)
	}
	if got := faceSnippet(MoodDegraded); !strings.Contains(got, `class="lid-r"`) || strings.Contains(got, `class="lid-l"`) {
		t.Errorf("degraded: expected right-eye-only lid, got %q", got)
	}
	if got := faceSnippet(MoodDown); !strings.Contains(got, `class="lid-l"`) || !strings.Contains(got, `class="lid-r"`) || !strings.Contains(got, ">z<") {
		t.Errorf("down: expected both lids + z, got %q", got)
	}
}

func TestRenderBadgeDetail(t *testing.T) {
	if got := Render(PetStatus{Mood: MoodHappy, Uptime: 25 * time.Hour}); !strings.Contains(got, "up 1d1h") {
		t.Errorf("uptime detail missing, got %q", got)
	}
	if got := Render(PetStatus{Mood: MoodDegraded, Down: []string{"redis"}}); !strings.Contains(got, "down: redis") {
		t.Errorf("down-list detail missing, got %q", got)
	}
}

func TestHumanDur(t *testing.T) {
	cases := map[time.Duration]string{
		45 * time.Second:              "45s",
		12 * time.Minute:              "12m",
		3*time.Hour + 12*time.Minute:  "3h12m",
		25*time.Hour + 30*time.Minute: "1d1h",
	}
	for d, want := range cases {
		if got := humanDur(d); got != want {
			t.Errorf("humanDur(%s) = %q, want %q", d, got, want)
		}
	}
}

// 两处 pet.svg 是同一份源图：docs/ 为可编辑源，本目录为 go:embed 副本。
// 独立拷贝本模块（无 docs/）时跳过。
func TestPetSVGInSyncWithDocs(t *testing.T) {
	const docsPath = "../../../../docs/pet.svg"
	b, err := os.ReadFile(docsPath)
	if os.IsNotExist(err) {
		t.Skip("docs/pet.svg not present")
	}
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != petSVGTemplate {
		t.Errorf("%s 与 internal/pkg/pet.svg 不一致，请同步后重跑", docsPath)
	}
}
