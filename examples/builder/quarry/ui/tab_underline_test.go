package ui

import (
	"image/color"
	"testing"
)

// TestActiveTabHasAccentUnderline: the active tab is marked by a 2 px accent
// line directly above the strip's hairline, and inactive tabs are not.
func TestActiveTabHasAccentUnderline(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("tab-new")
	act := r.box("tab-" + itoa(r.a.activeTabPtr().id))
	other := r.box("tab-" + itoa(r.a.tabs[0].id))
	accent := color.RGBA{0x3e, 0xcf, 0x8e, 0xff}
	for _, y := range []int{int(act.Y+act.H) - 2, int(act.Y+act.H) - 1} {
		if got := pixelAt(r, int(act.X+act.W/2), y); got != accent {
			t.Fatalf("active tab underline at y=%d is %v, want %v", y, got, accent)
		}
		if got := pixelAt(r, int(other.X+other.W/2), y); got == accent {
			t.Fatalf("inactive tab shows the accent at y=%d", y)
		}
	}
}
