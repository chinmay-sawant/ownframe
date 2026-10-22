package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestResizeReResolvesHoverAndPress(t *testing.T) {
	t.Parallel()

	app := &moveScreen{fakeScreen: fakeScreen{
		width: 600, height: 500,
		minW: 1, minH: 1,
		maxW: 2560, maxH: 2560,
	}}
	app.boxes = []layout.Box{{ID: "close", X: 520, Y: 20, W: 60, H: 40}}
	app.hover = "close"
	app.pressed = "close"

	shell := newShell(context.Background(), app)
	shell.cursorX, shell.cursorY = 550, 40
	shell.mouseDown = true

	shell.Layout(300, 500)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.hover != "" || app.pressed != "" {
		t.Fatalf("hover = %q, pressed = %q after the control moved away", app.hover, app.pressed)
	}

	// Widen back: the control returns under the cursor.
	shell.Layout(600, 500)
	if err := shell.resize(); err != nil { // moving, inside the throttle
		t.Fatal(err)
	}

	if err := shell.resize(); err != nil { // settled
		t.Fatal(err)
	}

	if app.hover != "close" || app.pressed != "close" {
		t.Fatalf("hover = %q, pressed = %q after the control returned", app.hover, app.pressed)
	}
}
