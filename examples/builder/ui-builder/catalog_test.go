package main

import (
	"bytes"
	"context"
	"errors"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestCustomComponentsKeepIndependentState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:custom"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, a.page, "custom-project-0-pin"); err != nil {
		t.Fatal(err)
	}
	if !a.custom.pinned[0] || a.custom.pinned[1] {
		t.Fatal("first instance changed the second instance")
	}
	if err := pagetest.Click(ctx, a.page, "custom-project-1-open"); err != nil {
		t.Fatal(err)
	}
	if a.custom.opened != "Release kit" {
		t.Fatalf("opened=%q", a.custom.opened)
	}
	if err := pagetest.Click(ctx, a.page, "custom-project-1-pin"); err != nil {
		t.Fatal(err)
	}
	if !a.custom.pinned[0] || !a.custom.pinned[1] {
		t.Fatal("second instance did not keep its state")
	}
	if err := pagetest.ClickAction(ctx, a.page, "theme:mode:light"); err != nil {
		t.Fatal(err)
	}
	if !a.custom.pinned[0] || !a.custom.pinned[1] {
		t.Fatal("theme change reset component state")
	}
	if err := a.page.Back(ctx); !errors.Is(err, ownframe.ErrNoHistory) {
		t.Fatalf("updates navigated: %v", err)
	}
}

func TestGlobalThemesAndLocalCustomProperties(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.tab = "custom"
	for _, appearance := range []galleryAppearance{galleryLight, galleryDark, galleryStudio} {
		a.theme.Appearance = appearance
		a.themeDirty = true
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(a.page.Testing().PNGWithContext(ctx)))
		if err != nil {
			t.Fatal(err)
		}
		box, ok := pagetest.BoxByID(a.page, "custom-project-1-open")
		if !ok {
			t.Fatal("missing local action")
		}
		got := color.RGBAModel.Convert(img.At(int(box.X)+5, int(box.Y+box.H/2)))
		if got != (color.RGBA{R: 120, G: 82, B: 29, A: 255}) {
			t.Fatalf("%s erased local color: %v", appearance, got)
		}
	}
}

func TestCustomAndThemePanelsFitSupportedViewports(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{900, 1120} {
		a.page.SetSize(width, 610)
		for _, tab := range []string{"widgets", "theme", "custom", "compare", "catalog"} {
			a.tab = tab
			if err := a.refresh(ctx); err != nil {
				t.Fatal(err)
			}
			for _, box := range a.page.Boxes() {
				if box.X+box.W > float64(width)+1 {
					t.Fatalf("%s overflow at %d: id=%s right=%.0f", tab, width, box.ID, box.X+box.W)
				}
			}
		}
	}
}

func TestButtonVariantAndSizeControls(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"widgets:variant:outlined", "widgets:size:large"} {
		if err := pagetest.ClickAction(ctx, a.page, action); err != nil {
			t.Fatal(err)
		}
	}
	if a.widgets.variant != "outlined" || a.widgets.size != chadcn.ButtonLarge {
		t.Fatal("catalog controls did not update options")
	}
}

func TestCatalogRendersEveryEntry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, box := range a.page.Boxes() {
		if box.Text == "ChadCN catalog" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("catalog heading missing")
	}
	if got := len(catalogEntries()); got != 78 {
		t.Fatalf("catalog entries=%d want 78", got)
	}
}

func TestCatalogOpensFromSidebarClick(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:catalog"); err != nil {
		t.Fatal(err)
	}
	if a.tab != "catalog" {
		t.Fatalf("tab=%q, want catalog", a.tab)
	}
	found := map[string]bool{}
	for _, box := range a.page.Boxes() {
		found[box.Text] = true
	}
	for _, want := range []string{"Area Chart", "Filter Bar", "Stat Card", "Status Segments", "Hex Grid"} {
		if !found[want] {
			t.Fatalf("catalog heading %q missing after sidebar click", want)
		}
	}
}

func TestCatalogFilterPillClickUpdatesNotice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "catalog:filter:weekly"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(a.notice), "weekly") {
		t.Fatalf("notice=%q, want it to mention weekly", a.notice)
	}
	if a.tab != "catalog" {
		t.Fatalf("tab=%q, want catalog", a.tab)
	}
}

func TestCatalogNavRowClickUpdatesNotice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "catalog:nav:work-orders"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(a.notice), "work-orders") {
		t.Fatalf("notice=%q, want it to mention work-orders", a.notice)
	}
}

func TestCatalogWheelDefersRebuild(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.SetSize(900, 610)
	a.tab = "catalog"
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	if a.center.Metrics().MaxOffset() == 0 {
		t.Fatal("compact catalog should overflow")
	}
	viewport := a.centerViewport()
	x, y := viewport.X+viewport.W/2, viewport.Y+20
	gen := a.page.Probe().Generation()
	before := a.center.Offset()
	thumbBefore, ok := pagetest.BoxByID(a.page, builderScrollThumbID)
	if !ok {
		t.Fatal("missing reusable thumb")
	}
	for range 3 {
		if _, err := a.handleWheel(ctx, x, y, 0, -1); err != nil {
			t.Fatal(err)
		}
	}
	if got := a.page.Probe().Generation(); got != gen {
		t.Fatalf("handleWheel rebuilt synchronously: generation %d -> %d", gen, got)
	}
	if a.center.Offset() <= before {
		t.Fatal("wheel did not advance catalog")
	}
	if want := before + 3*int(builderScrollStep); a.center.Metrics().MaxOffset() >= want {
		if got := a.center.Offset(); got != want {
			t.Fatalf("offset=%d want %d after 3 wheel notches", got, want)
		}
	}
	if err := a.refreshOnResize(ctx); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Probe().Generation(); got <= gen {
		t.Fatalf("refreshOnResize did not rebuild: generation %d -> %d", gen, got)
	}
	thumbAfter, ok := pagetest.BoxByID(a.page, builderScrollThumbID)
	if !ok {
		t.Fatal("missing reusable thumb after refresh")
	}
	if thumbAfter.Y == thumbBefore.Y {
		t.Fatalf("thumb did not move after refresh: before=%+v after=%+v offset=%d", thumbBefore, thumbAfter, a.center.Offset())
	}
}
