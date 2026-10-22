package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestScrollClampsAfterARedraw(t *testing.T) {
	t.Parallel()

	app := &scrollScreen{fakeScreen: fakeScreen{width: 600, height: 300, minW: 1, minH: 1}}
	app.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: 600, H: 1200}}

	shell := newShell(context.Background(), app)

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	shell.scrollY = 900

	// A Load, Back, or Forward redraws at the same size with new content.
	app.redraws++
	app.boxes = nil

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	if shell.scrollY != 0 {
		t.Fatalf("scrollY = %d after the redraw, want 0", shell.scrollY)
	}
}
