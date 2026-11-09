package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestSliderDragUpdatesCatalogAndWidgets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	progress := a.widgets.progress
	dragSlider(t, ctx, a, widgetsSliderID, 0, 30)
	if a.widgets.progress != progress {
		t.Fatalf("vertical drag changed progress from %d to %d", progress, a.widgets.progress)
	}
	dragSlider(t, ctx, a, widgetsSliderID, 80, 0)
	if a.widgets.progress <= progress {
		t.Fatalf("widgets slider stayed at %d", a.widgets.progress)
	}

	claimed, err := a.claimSliderDrag(ctx, ownframe.Box{ID: "not-a-slider"})
	if err != nil || claimed {
		t.Fatalf("non-slider claim = %v, %v", claimed, err)
	}

	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	before := a.catalogSlider
	dragSlider(t, ctx, a, catalogSliderID, 120, 0)
	if a.catalogSlider <= before {
		t.Fatalf("catalog slider stayed at %v", a.catalogSlider)
	}
	held := a.catalogSlider

	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	thumb, ok := pagetest.BoxByID(a.page, builderScrollThumbID)
	if !ok {
		t.Fatal("missing scroll thumb")
	}
	offset := a.center.Offset()
	x, y := thumb.X+thumb.W/2, thumb.Y+thumb.H/2
	claimed, err = a.page.Testing().BeginDrag(ctx, x, y)
	if err != nil || !claimed {
		t.Fatalf("thumb drag claimed=%v err=%v", claimed, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x, y+40); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.center.Offset() <= offset {
		t.Fatal("thumb drag did not scroll")
	}
	if a.catalogSlider != held {
		t.Fatalf("thumb drag changed catalog slider from %v to %v", held, a.catalogSlider)
	}
}

func dragSlider(t *testing.T, ctx context.Context, a *builderApp, id string, dx, dy float64) {
	t.Helper()
	box, ok := pagetest.BoxByID(a.page, id)
	if !ok {
		t.Fatalf("missing slider %s", id)
	}
	x, y := box.X+box.W/2, box.Y+box.H/2
	claimed, err := a.page.Testing().BeginDrag(ctx, x, y)
	if err != nil || !claimed {
		t.Fatalf("drag %s claimed=%v err=%v box=%+v", id, claimed, err, box)
	}
	if err := a.page.Testing().MoveDrag(ctx, x+dx, y+dy); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
}
