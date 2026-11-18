package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestWebInputDrivesTheLiveSession plays what the -web routes deliver (a box
// click, a key, typed text) through webHandlers into the same navigation
// session the window uses, with the page scrolled so the offset is tested too.
func TestWebInputDrivesTheLiveSession(t *testing.T) {
	r := newRig(t, "nav")
	in := &webInput{}
	r.runner.src = in.source()
	h := webHandlers(r.page, in)
	ctx := context.Background()

	box := func(id string) ownframe.Box {
		it := r.box(id)

		return ownframe.Box{X: it.X, Y: it.Y, W: it.W, H: it.H}
	}

	_, height := r.page.Size()
	r.page.Scroll().SetScrollOffset(0, int(max(r.box("n-a").Y-float64(height)/3, 0)))

	must(t, h.Click(ctx, box("n-a")))
	must(t, r.step(rigStep))

	if got := r.nav.st.focus["focus"]; got != "n-a" {
		t.Fatalf("a web click on n-a focused %q", got)
	}

	// A click and a key in one frame would run the key first, so give the
	// click its own frame, as a person does.
	must(t, h.KeyDown(ctx, "arrowdown"))
	must(t, r.step(rigStep))

	if got := r.nav.st.focus["focus"]; got != "n-f" {
		t.Fatalf("a web ArrowDown from n-a moved focus to %q, want n-f (nav-down)", got)
	}

	if got := r.textOf("chip-focus"); got != "moved n-a -> n-f via nav-down" {
		t.Errorf("chip = %q", got)
	}

	// Typed text and Backspace reach a field the same way.
	must(t, h.Click(ctx, box("fs-content")))
	must(t, r.step(rigStep))
	before := r.box("fs-content").W

	must(t, h.Type(ctx, "xyz"))
	must(t, r.step(rigStep))

	if grown := r.box("fs-content").W - before; grown <= 0 {
		t.Errorf("typed text did not grow the content field (%g)", grown)
	}

	grown := r.box("fs-content").W

	must(t, h.Backspace(ctx))
	must(t, r.step(rigStep))

	if after := r.box("fs-content").W; after >= grown {
		t.Errorf("Backspace did not shrink the content field (%g then %g)", grown, after)
	}
}
