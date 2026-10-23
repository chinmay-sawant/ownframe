package ui

import (
	"context"
	"testing"
)

func TestResizeKeepsAnchor(t *testing.T) {
	a := newHeadless(t)
	ctx := context.Background()

	a.page.Scroll().SetScrollOffset(0, 30*RowH)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	anchor := a.pager.AnchorID
	if anchor == 0 {
		t.Fatal("no reading anchor recorded")
	}

	a.page.SetSize(1100, 400)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	want := a.pager.AnchorOffset(HeaderH)
	if !a.havePending || a.pendingOffset != want {
		t.Fatalf("resize scroll = %d/%v, want %d", a.pendingOffset, a.havePending, want)
	}

	if a.pager.AnchorID != anchor {
		t.Fatalf("anchor moved from %d to %d", anchor, a.pager.AnchorID)
	}
}

func TestResizeAtBottomStaysBottom(t *testing.T) {
	a := newHeadless(t)
	ctx := context.Background()

	a.page.Scroll().SetScrollOffset(0, a.listHeight()-720)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	a.page.SetSize(1100, 400)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	want := a.listHeight() - 400
	if !a.havePending || a.pendingOffset != want {
		t.Fatalf("bottom resize scroll = %d/%v, want %d", a.pendingOffset, a.havePending, want)
	}
}
