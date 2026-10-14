package board

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestCascadeStaysBanded mirrors the dump's rotate scenario and frame
// classification with the rotation driven by the tick: after the initial
// draw no frame may mark the whole viewport, and the rotated tiles must keep
// flipping while the generation holds.
func TestCascadeStaysBanded(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().TakeDirtyRects()

	for i := 0; i < 3; i++ {
		runTick(t, ctx, app)
	}

	gen := app.Page().Generation()
	forceRotation(app)

	full, banded, empty := 0, 0, 0

	for i := 0; i < 150; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		if !app.Page().FrameDirty() {
			t.Fatalf("frame %d did not opt into the frame dirty path", i)
		}

		rects, ok := app.Page().TakeDirtyRects()
		if !ok {
			empty++

			continue
		}

		if coversFrame(app.Page().Display(), rects) {
			full++

			continue
		}

		banded++
	}

	if full != 0 {
		t.Fatalf("%d frames marked the whole viewport, want bands", full)
	}

	if banded == 0 {
		t.Fatal("no frame marked a band")
	}

	if empty == 0 {
		t.Fatal("no frame blitted")
	}

	if got := app.Page().Generation(); got != gen {
		t.Fatalf("cascade redrew the page: generation %d -> %d", gen, got)
	}
}

// coversFrame reports whether any rect covers every pixel of the display.
func coversFrame(d *ownframe.Display, rects []image.Rectangle) bool {
	if d == nil {
		return false
	}

	frame := image.Rect(0, 0, d.Width, d.Height)
	for _, r := range rects {
		if frame.In(r) {
			return true
		}
	}

	return false
}
