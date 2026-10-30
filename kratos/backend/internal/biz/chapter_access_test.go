package biz

import (
	"testing"

	"open-novel/backend/internal/pkg"
)

// chapterAccess 纯逻辑：下架/禁用/VIP 门槛/正常访问/管理端放行。
func TestChapterAccess(t *testing.T) {
	cases := []struct {
		name      string
		bookStat  int8
		chStatus  uint8
		chVip     uint8
		uid       int64
		vipActive bool
		isStaff   bool
		want      error
	}{
		{"normal free", 1, 1, 0, 0, false, false, nil},
		{"book offline", 0, 1, 0, 1, true, false, pkg.ErrChapterNF},
		{"chapter disabled", 1, 0, 0, 1, true, false, pkg.ErrChapterDisabled},
		{"vip anonymous", 1, 1, 1, 0, false, false, pkg.ErrVIPRequired},
		{"vip inactive", 1, 1, 1, 7, false, false, pkg.ErrVIPRequired},
		{"vip active", 1, 1, 1, 7, true, false, nil},
		{"vip book offline wins", 0, 1, 1, 7, true, false, pkg.ErrChapterNF},
		{"staff book offline", 0, 1, 0, 9, false, true, nil},
		{"staff chapter disabled", 1, 0, 0, 9, false, true, nil},
		{"staff vip anonymous", 1, 1, 1, 0, false, true, nil},
		{"staff all blocked", 0, 0, 1, 0, false, true, nil},
	}
	for _, c := range cases {
		if got := chapterAccess(c.bookStat, c.chStatus, c.chVip, c.uid, c.vipActive, c.isStaff); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
