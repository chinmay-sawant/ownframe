package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestContentSizeIgnoresClippedOverflow pins that paint clipped by an
// overflow ancestor does not extend the scrollable document: a marquee
// track twice its lane must not summon a horizontal scrollbar.
func TestContentSizeIgnoresClippedOverflow(t *testing.T) {
	t.Parallel()

	s := newShell(context.Background(), &fakeScreen{width: 1600, height: 900})

	s.app.(*fakeScreen).boxes = []layout.Box{
		{ID: "page", Tag: "div", X: 0, Y: 0, W: 1600, H: 900},
		{
			ID: "tape-track", Tag: "div", X: 400, Y: 860, W: 2097, H: 28,
			OverflowClip:    [4]float64{400, 860, 800, 28},
			HasOverflowClip: true,
			OperationIDs:    []uint64{7},
			ParentIndex:     0,
		},
	}

	w, h := s.contentSize()
	if w != 1600 || h != 900 {
		t.Fatalf("content = %dx%d, want 1600x900 (clipped track ignored)", w, h)
	}

	unclipped := &fakeScreen{width: 1600, height: 900}
	s2 := newShell(context.Background(), unclipped)
	unclipped.boxes = []layout.Box{
		{ID: "page", Tag: "div", X: 0, Y: 0, W: 1600, H: 900},
		{ID: "wide", Tag: "div", X: 0, Y: 0, W: 2000, H: 900, OperationIDs: []uint64{9}},
	}

	w, _ = s2.contentSize()
	if w != 2000 {
		t.Fatalf("unclipped content width = %d, want 2000", w)
	}
}
