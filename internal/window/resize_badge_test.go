package window

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// replayScreen hands out one display list per generation.
type replayScreen struct {
	fakeScreen
}

func (r *replayScreen) Display() *layout.Display {
	return &layout.Display{Width: r.width, Height: r.height}
}

func TestMotionKeepsTheDrawnArtifact(t *testing.T) {
	t.Parallel()

	app := &replayScreen{fakeScreen: fakeScreen{width: 480, height: 640, minW: 1, minH: 1}}
	shell := newShell(context.Background(), app)

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	drawn := shell.display
	shell.lastRelayout = time.Now()

	shell.Layout(600, 700)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if shell.display != drawn || shell.fallback {
		t.Fatal("a throttled motion frame rebuilt the drawn artifact")
	}
}
