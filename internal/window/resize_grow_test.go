package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// growScreen keeps the previous display until Redraw builds the new one, so
// a test can catch a cursor mapped through the stale artifact.
type growScreen struct {
	*fakeScreen
	display *layout.Display
	hoverX  float64
}

func (g *growScreen) Display() *layout.Display { return g.display }

func (g *growScreen) Hover(_ context.Context, x, y float64) error {
	g.hoverX = x

	return nil
}

func (g *growScreen) Redraw(context.Context) error {
	g.redraws++
	g.display = &layout.Display{Width: g.width, Height: g.height}

	return nil
}

func TestResizeGrowMapsThroughTheCommittedFrame(t *testing.T) {
	t.Parallel()

	app := &growScreen{
		fakeScreen: &fakeScreen{width: 300, height: 500, minW: 1, minH: 1},
		display:    &layout.Display{Width: 300, Height: 500},
	}
	s := newShell(context.Background(), app)
	s.cursorX, s.cursorY = 550, 40

	s.Layout(600, 500)
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if app.hoverX != 550 {
		t.Fatalf("hover x = %v, want 550 through the committed frame", app.hoverX)
	}
}
