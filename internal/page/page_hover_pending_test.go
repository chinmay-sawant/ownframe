package page

import (
	"context"
	"testing"
)

// TestHoverPendingFullRecoversChangeSet pins the full-mark recovery: a full
// frame was committed after the host painted, but the host has not taken that
// damage when the hover patch lands. The page diffs against the painted
// display, so the report and the damage describe the hover instead of the
// stale full frame.
func TestHoverPendingFullRecoversChangeSet(t *testing.T) {
	p := newHoverPatchPage(t, true)
	p.TakeDirtyRects()

	p.markFull()
	p.generation++ // a display commit landed after the paint, before this hover

	hoverAt(t, p, "row-b")

	patch := p.LastPatch()
	if patch.Full || patch.Changed > 4 {
		t.Fatalf("patch = %+v, want the hover change only", patch)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("hover did not dirty")
	}

	if rect.Dx()*rect.Dy() >= p.width*p.height {
		t.Fatalf("damage %v covers the frame", rect)
	}
}

// TestHoverPendingFullWithoutPaintStaysFull pins the other side: before the
// first paint there is no baseline to recover from, so the frame stays full.
func TestHoverPendingFullWithoutPaintStaysFull(t *testing.T) {
	p := newHoverPatchPage(t, true)
	p.markFull()

	hoverAt(t, p, "row-b")

	if patch := p.LastPatch(); !patch.Full {
		t.Fatalf("patch = %+v, want the frame kept full before the first paint", patch)
	}
}

// TestHoverFullDamageKeepsChangedCount pins the perf counters: a full frame
// still reports every operation as dirty, but the changed count is the diff's
// own result rather than the whole display.
func TestHoverFullDamageKeepsChangedCount(t *testing.T) {
	p, err := New(Config{
		HTML:        hoverPatchFixture,
		Width:       320,
		Height:      200,
		RetainState: true,
		Perf:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	hoverAt(t, p, "row-a")

	stats := p.Stats()
	if stats.ChangedOps != p.LastPatch().Changed {
		t.Fatalf("changed = %d, patch = %d", stats.ChangedOps, p.LastPatch().Changed)
	}

	if stats.ChangedOps >= stats.DirtyOps {
		t.Fatalf("changed %d >= dirty %d, want the diff count", stats.ChangedOps, stats.DirtyOps)
	}
}
