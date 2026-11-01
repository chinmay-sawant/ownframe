package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestDevThumbWinsOverDockEdge checks the overlapping strip between the page
// scrollbar thumb and the dock resize edge scrolls instead of resizing. The
// thumb occupies the last 8 px of the page view and the edge straddles the
// dock boundary by 4 px on each side, so the right half of the thumb is also
// the edge. pointerScrollbar must consume that press before devPointer can
// start a resize.
func TestDevThumbWinsOverDockEdge(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.boxes = []layout.Box{
		{ID: "known", Tag: "div", X: 10, Y: 10, W: 100, H: 40},
		{ID: "tall", Tag: "div", X: 0, Y: 0, W: 100, H: 2000},
	}
	s := newDevShell(d)
	s.screenW, s.screenH = 800, 600
	s.dev.on = true
	s.devLayout()

	viewW, viewH := s.viewSize()
	// Match the page to the docked viewport so scrollbars are enabled:
	// a stretched page disables them and the thumb never shows.
	d.SetSize(viewW, viewH)
	s.devLayout()
	viewW, viewH = s.viewSize()
	contentW, contentH := s.contentSize()
	if !scrollbarVisible(contentH, viewH) {
		t.Fatalf("content %dx%d view %dx%d shows no vertical thumb", contentW, contentH, viewW, viewH)
	}

	panelX := int(s.dev.panel.X)
	thumbX := viewW - 2
	if !devOnEdge(s.dev.panel, thumbX, 10) {
		t.Fatalf("x=%d not on dock edge %+v", thumbX, s.dev.panel)
	}
	if _, _, ok := s.scrollbarGrab(thumbX, 10); !ok {
		t.Fatalf("x=%d y=10 not on scrollbar thumb (view %dx%d)", thumbX, viewW, viewH)
	}

	if !s.pointerScrollbar(thumbX, 10, true, true) {
		t.Fatal("thumb press did not consume the overlapping edge press")
	}
	if s.dev.resizing {
		t.Fatal("thumb press started a dock resize")
	}
	if s.dragAxis == dragNone {
		t.Fatal("thumb press did not start a scrollbar drag")
	}
	_ = panelX
}

// TestDevPanelSlideOffCancelsTap checks a touch that lands on a dock button
// but slides past the slop never fires at its landing point.
func TestDevPanelSlideOffCancelsTap(t *testing.T) {
	t.Parallel()

	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	land := touchPos{id: 7, x: 1700, y: 100}
	if _, _, _ = devPanelStep([]touchPos{land}, panel, tr); len(tr.fingers) != 1 {
		t.Fatal("landing finger not tracked")
	}
	moved := touchPos{id: 7, x: 1700 + touchSlop + 2, y: 100}
	if _, _, _ = devPanelStep([]touchPos{moved}, panel, tr); !tr.fingers[0].drag {
		t.Fatal("horizontal slide past slop did not mark drag")
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 0 {
		t.Fatalf("slid finger tapped: %v", taps)
	}
}

// TestDevPanelTapFiresAtRelease checks a tap within the slop fires at its
// release point instead of its landing point.
func TestDevPanelTapFiresAtRelease(t *testing.T) {
	t.Parallel()

	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	if _, _, _ = devPanelStep([]touchPos{{id: 9, x: 1700, y: 100}}, panel, tr); len(tr.fingers) != 1 {
		t.Fatal("landing finger not tracked")
	}
	if _, _, _ = devPanelStep([]touchPos{{id: 9, x: 1702, y: 101}}, panel, tr); tr.fingers[0].drag {
		t.Fatal("move within slop marked drag")
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 1 {
		t.Fatalf("taps = %v, want one", taps)
	}
	if taps[0].x != 1702 || taps[0].y != 101 {
		t.Fatalf("tap = %v, want the release point 1702,101", taps[0])
	}
}
