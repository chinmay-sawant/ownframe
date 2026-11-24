package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestStatusTablesHaveNoOverlap(t *testing.T) {
	t.Parallel()

	for name := range painterDeclines {
		if opOnly[name] {
			t.Errorf("%q is listed in both painterDeclines and opOnly", name)
		}
	}
}

func TestNativelyReplayedPropertiesNotDeclined(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	head := genHead()
	mask, err := shapesMask("mask")
	if err != nil {
		t.Fatal(err)
	}

	byName := map[string]genProp{}
	for _, r := range genProps() {
		byName[r.Name] = r
	}

	for _, name := range []string{
		"writing-mode", "text-autospace", "object-fit", "object-view-box",
		"border-image", "border-image-source", "border-image-slice", "border-image-width",
		"visibility", "gap", "scale",
		"border-top-color", "border-right-color", "border-bottom-color", "border-left-color", "outline-color",
		"grid-auto-columns", "grid-auto-flow", "grid-row-start", "grid-row-end",
		"cx", "cy", "r", "x", "y",
		"marker", "marker-start", "marker-mid", "marker-end",
		"text-overflow", "text-emphasis-style",
		"image-rendering", "tab-size",
		"mask-position", "-webkit-mask-position", "mask-size", "-webkit-mask-size",
		"background-blend-mode",
		"position-area", "position-try", "position-try-fallbacks", "position-try-order",
		"paint-order", "stop-color", "stop-opacity", "stroke-linejoin", "stroke-miterlimit",
		"color-interpolation", "color-interpolation-filters",
		"filter", "mask-clip", "mask-origin", "-webkit-mask-clip", "-webkit-mask-origin",
		"mask-border-mode", "mask-border-outset", "mask-border-repeat", "mask-border-slice", "mask-border-width",
		"-webkit-mask-box-image-outset", "-webkit-mask-box-image-repeat", "-webkit-mask-box-image-slice", "-webkit-mask-box-image-width",
		"fill-rule", "clip-rule", "font-width", "font-variant-emoji", "rule-overlap",
		"backdrop-filter", "color-scheme", "transform-box", "grid-template-areas", "rule-break", "position-visibility", "isolation",
		"font-feature-settings", "font-variant-ligatures", "font-language-override", "font-variant-alternates",
		"font-synthesis", "font-synthesis-weight", "font-variant-numeric", "font-variant-east-asian",
	} {
		if noPaint[name] || painterDeclines[name] || opOnly[name] {
			t.Errorf("%q should not be in noPaint/painterDeclines/opOnly", name)
		}

		r, ok := byName[name]
		if !ok {
			t.Fatalf("%q not found in genProps", name)
		}

		spec := specFor(r)
		page, err := ownframe.New(ownframe.Config{
			HTML:   head + spec.expand(r, "p", true),
			Width:  300,
			Height: 200,
		})
		if err != nil {
			t.Fatalf("%s: New: %v", name, err)
		}

		page.SetImage("mask", mask)

		if err = page.Redraw(ctx); err != nil {
			t.Fatalf("%s: Redraw: %v", name, err)
		}

		if page.Probe().Display() == nil {
			t.Errorf("%s fell back to bitmap instead of replaying natively", name)
		}
	}
}
