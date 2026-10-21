package page

import (
	"context"
	"testing"
)

// afterRedrawWideHTML is a box that follows the viewport, so a stale box
// shows a stale width.
const afterRedrawWideHTML = `<style>body { margin: 0; } div { width: 100%; height: 20px; background: #ff0000; }</style><div id="a"></div>`

func TestAfterRedrawSeesTheNewBoxes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawWideHTML)

	var width int
	var box Box
	p.SetAfterRedraw(func() {
		if d := p.Display(); d != nil {
			width = d.Width
		}
		if b, ok := boxWithID(p.Boxes(), "a"); ok {
			box = b
		}
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if width != 200 {
		t.Fatalf("display width = %d, want 200", width)
	}
	if box.W != 200 {
		t.Fatalf("box = %+v, want the 200px box", box)
	}

	p.SetSize(320, 200)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if width != 320 {
		t.Fatalf("display width = %d, want 320", width)
	}
	if box.W != 320 {
		t.Fatalf("box = %+v, want the 320px box", box)
	}
}

func TestAfterRedrawRefusesAReentrantRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	p.SetAfterRedraw(func() {
		calls++
		if err := p.Redraw(ctx); err != nil {
			t.Errorf("Redraw in the callback: %v", err)
		}
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// The callback asks for a draw, the draw runs after the callback
	// returned, and the callback runs again in it. The second request is
	// dropped, so the callback cannot loop.
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestAfterRedrawNilRemovesTheCallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := afterRedrawPage(t, afterRedrawHTML)

	calls := 0
	p.SetAfterRedraw(func() { calls++ })
	p.SetAfterRedraw(nil)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("calls = %d, want 0", calls)
	}
}
