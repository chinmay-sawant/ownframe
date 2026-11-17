package main

import (
	"strings"
	"testing"
)

// TestActionSheetStaysUnderThePinnedHeader checks the spatial-navigation-action
// sheet after it is scrolled right and down. The sheet must still move, and it
// must not carry a z-index: the pinned top block has z-index 10, and a
// scrolled sheet with a larger one painted its cells over that header.
func TestActionSheetStaysUnderThePinnedHeader(t *testing.T) {
	r := newRig(t, "nav")

	startX, startY := r.box("s-content").X, r.box("s-content").Y

	must(t, r.tap("s-a"))

	for range 3 {
		must(t, r.key("ArrowRight"))
		must(t, r.key("ArrowDown"))
	}

	if got := r.box("s-content").X; !near(got, startX-120) {
		t.Errorf("after 3 Right keys the sheet is at x=%g, want %g", got, startX-120)
	}

	if got := r.box("s-content").Y; !near(got, startY-120) {
		t.Errorf("after 3 Down keys the sheet is at y=%g, want %g", got, startY-120)
	}

	if style := r.style("sbox"); strings.Contains(style, "z-index") {
		t.Errorf("the scrolled sheet box sets a z-index and would paint over the pinned header: %q", style)
	}
}
