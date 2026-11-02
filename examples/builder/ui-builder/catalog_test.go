package main

import (
	"bytes"
	"context"
	"errors"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui/components"
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
		for _, tab := range []string{"widgets", "theme", "custom", "compare"} {
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
	if a.widgets.variant != "outlined" || a.widgets.size != components.ButtonLarge {
		t.Fatal("catalog controls did not update options")
	}
}
