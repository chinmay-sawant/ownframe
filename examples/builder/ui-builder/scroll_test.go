package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestCatalogScrollRangeSurvivesWheelAndNativeDrag(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.SetSize(900, 610)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	initial := a.center.Metrics()
	if initial.MaxOffset() == 0 {
		t.Fatal("compact catalog should overflow")
	}
	viewport := a.centerViewport()
	for range 3 {
		if _, err := a.handleWheel(ctx, viewport.X+viewport.W/2, viewport.Y+20, 0, -1); err != nil {
			t.Fatal(err)
		}
		if err := a.refreshOnResize(ctx); err != nil {
			t.Fatal(err)
		}
		if a.center.Metrics().Content != initial.Content {
			t.Fatalf("scroll changed intrinsic content height: before=%+v after=%+v", initial, a.center.Metrics())
		}
	}
	before := a.center.Offset()
	thumb, ok := pagetest.BoxByID(a.page, builderScrollThumbID)
	if !ok {
		t.Fatal("missing reusable thumb")
	}
	x, y := thumb.X+thumb.W/2, thumb.Y+thumb.H/2
	claimed, err := a.page.Testing().BeginDrag(ctx, x, y)
	if err != nil || !claimed {
		t.Fatalf("begin drag: claimed=%v err=%v", claimed, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x, y+40); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.center.Offset() <= before {
		t.Fatal("native drag did not advance catalog")
	}
	a.tab = "dialog"
	a.dialog.open = true
	before = a.center.Offset()
	if _, err := a.handleWheel(ctx, x, y, 0, -1); err != nil {
		t.Fatal(err)
	}
	if a.center.Offset() != before {
		t.Fatal("modal allowed background scrolling")
	}
}
