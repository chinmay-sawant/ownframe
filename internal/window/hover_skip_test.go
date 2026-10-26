package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// hoverScreen counts hover passes and moves the generation like a redraw.
type hoverScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	hovers  int
}

func (h *hoverScreen) Display() *layout.Display { return h.display }
func (h *hoverScreen) Generation() uint64       { return h.gen }
func (h *hoverScreen) Hover(context.Context, float64, float64) error {
	h.hovers++

	return nil
}

func newHoverShell(gen uint64) (*shell, *hoverScreen) {
	app := &hoverScreen{
		fakeScreen: &fakeScreen{width: 200, height: 200, minW: 1, minH: 1},
		display:    &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1},
		gen:        gen,
	}
	s := &shell{
		app: app, ctx: context.Background(),
		display: app.display, seq: gen,
		screenW: 200, screenH: 200,
	}

	return s, app
}

func TestHoverStillPointerSkipsUntilRedraw(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)
	s.frameCommit()

	for i := 0; i < 2; i++ {
		if err := s.hoverCursor(50, 50, 50, 50); err != nil {
			t.Fatal(err)
		}
	}
	if app.hovers != 1 {
		t.Fatalf("still pointer calls = %d, want 1", app.hovers)
	}

	// A redraw moves the generation and the shell rebuilds, so the next
	// frame resolves hover again.
	app.gen++
	app.display = &layout.Display{Width: 200, Height: 200, PointsPerPixel: 1}
	if err := s.syncImage(); err != nil {
		t.Fatal(err)
	}
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 2 {
		t.Fatalf("post-redraw calls = %d, want 2", app.hovers)
	}
}

func TestHoverMovedPointResolves(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)
	s.frameCommit()

	points := [][2]float64{{50, 50}, {60, 60}, {60, 108}}
	for _, p := range points {
		if err := s.hoverCursor(int(p[0]), int(p[1]), p[0], p[1]); err != nil {
			t.Fatal(err)
		}
	}
	if app.hovers != 3 {
		t.Fatalf("moving point calls = %d, want 3", app.hovers)
	}
}

// TestHoverDeferredUntilFirstFrame keeps the first update from pushing the
// cursor hover to the page, so startup paints without a hover redraw. The
// skip leaves the hover due, so the update after the first committed frame
// still resolves it.
func TestHoverDeferredUntilFirstFrame(t *testing.T) {
	t.Parallel()

	s, app := newHoverShell(1)

	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 0 {
		t.Fatalf("pre-frame calls = %d, want 0", app.hovers)
	}

	s.frameCommit()
	if err := s.hoverCursor(50, 50, 50, 50); err != nil {
		t.Fatal(err)
	}
	if app.hovers != 1 {
		t.Fatalf("post-frame calls = %d, want 1", app.hovers)
	}
}
