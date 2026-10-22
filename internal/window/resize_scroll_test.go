package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// scrollScreen is tall above 500 px wide, so a shrink makes the content
// shorter than the current offset.
type scrollScreen struct {
	fakeScreen
}

func (s *scrollScreen) Redraw(context.Context) error {
	s.redraws++
	s.boxes = nil

	if s.width > 500 {
		s.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: float64(s.width), H: 1200}}
	}

	return nil
}

func (s *scrollScreen) Display() *layout.Display {
	return &layout.Display{Width: s.width, Height: s.height}
}

func TestScrollClampsAfterAShrink(t *testing.T) {
	t.Parallel()

	app := &scrollScreen{fakeScreen: fakeScreen{width: 600, height: 300, minW: 1, minH: 1}}
	app.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: 600, H: 1200}}

	shell := newShell(context.Background(), app)

	if _, contentH := shell.contentSize(); contentH != 1200 {
		t.Fatalf("content height = %d, want 1200", contentH)
	}

	shell.scrollY = 900 // the bottom of a 600 x 300 window

	shell.Layout(400, 300)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	if shell.scrollY != 0 {
		t.Fatalf("scrollY = %d after the shrink, want 0", shell.scrollY)
	}
}
