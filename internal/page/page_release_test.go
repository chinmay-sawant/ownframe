package page

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestDisplayReleaseOnCommit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<div id="box">first</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	// Host takes the dirty rects, setting p.painted and allowing maybeReleasePainted to clear it.
	p.TakeDirtyRects()

	// Update content to force a fresh layout.
	if err := p.Load(ctx, `<div id="box">second</div>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	second := p.Display()
	if second == nil {
		t.Fatal("no display after second redraw")
	}
	if second == first {
		t.Fatal("expected second display to differ from first")
	}

	// first should have been released during commit; calling Release again is safe and idempotent.
	first.Release()
}

func TestDisplayReleaseOnCommitWithRetainedState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box">first</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	// Host takes dirty rects (p.painted now holds first, and maybeReleasePainted returns early because RetainState=true).
	p.TakeDirtyRects()

	// Second commit: previous (first) must still be released even when previous == p.painted.
	if err := p.Load(ctx, `<div id="box">second</div>`); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display to differ from first")
	}

	// Verify first has been released: its result is nil, so RelayoutElement must return ErrNoRetainedState.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on commit: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}

func TestDisplayReleaseOnPatchElement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box" class="a">hello</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after redraw")
	}

	if err := p.PatchElement(ctx, "box", ElementEdit{Attribute: "class", Value: "b"}); err != nil {
		t.Fatal(err)
	}
	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display after patch")
	}

	// Verify first has been released during PatchElement.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on PatchElement: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}

func TestDisplayReleaseOnPatchElements(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{
		HTML:        `<div id="box1" class="a">one</div><div id="box2" class="a">two</div>`,
		Width:       320,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after redraw")
	}

	err = p.PatchElements(ctx, []ElementPatch{
		{ID: "box1", Edits: []ElementEdit{{Attribute: "class", Value: "b"}}},
		{ID: "box2", Edits: []ElementEdit{{Attribute: "class", Value: "b"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	second := p.Display()
	if second == nil || second == first {
		t.Fatal("expected second display after batch patch")
	}

	// Verify first has been released during PatchElements.
	_, relayoutErr := render.RelayoutElement(ctx, p.cache.Styled(), first, "box1", layout.Options{})
	if !errors.Is(relayoutErr, layout.ErrNoRetainedState) {
		t.Fatalf("first display not released on PatchElements: got %v, want %v", relayoutErr, layout.ErrNoRetainedState)
	}
}
