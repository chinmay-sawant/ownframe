package chadcn_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Loud hover and rest colors, so a disabled control that still takes the hover
// face cannot hide in a near-gray blend.
const disabledPaintVars = `:root{--of-accent:#ff00ff;--of-primary:#0000ff;--of-primary-hover:#00ff00;` +
	`--of-control:#0000ff;--of-surface:#ffffff;--of-ring:#ff0000;--of-input:#0000ff;` +
	`--of-background:#ffffff;--of-popover:#ffffff;--of-on-primary:#ffffff}`

func TestDisabledHoverRulesExcludeDisabled(t *testing.T) {
	t.Parallel()
	sheet := chadcn.Stylesheet
	for _, selector := range []string{
		`.of-button--tonal:not([disabled]):hover`,
		`.of-button--filled:not([disabled]):hover`,
		`.of-button--outlined:not([disabled]):hover`,
		`.of-button--text:not([disabled]):hover`,
		`.of-button--danger:not([disabled]):hover`,
		`.of-toggle:not([disabled]):not([aria-disabled="true"]):hover`,
		`.of-checkbox:not([disabled]):not([aria-disabled="true"]):hover .of-checkbox-box`,
		`.of-checkbox--checked:not([disabled]):not([aria-disabled="true"]):hover .of-checkbox-box`,
		`.of-radio-group-item:not([disabled]):not([aria-disabled="true"]):hover .of-radio-group-indicator`,
		`.of-radio-group-item--checked:not([disabled]):not([aria-disabled="true"]):hover .of-radio-group-indicator`,
		`.of-select-item:not([disabled]):hover`,
	} {
		if !strings.Contains(sheet, selector) {
			t.Fatalf("stylesheet missing %s", selector)
		}
	}
	for _, selector := range []string{
		`.of-button--tonal:hover {`,
		`.of-button--filled:hover {`,
		`.of-button--outlined:hover {`,
		`.of-button--text:hover {`,
		`.of-button--danger:hover {`,
		`.of-toggle:hover {`,
		`.of-checkbox:hover `,
		`.of-checkbox--checked:hover `,
		`.of-radio-group-item:hover `,
		`.of-radio-group-item--checked:hover `,
		`.of-select-item:hover {`,
	} {
		if strings.Contains(sheet, selector) {
			t.Fatalf("stylesheet still applies %s", selector)
		}
	}
	busy := selectFixRule(t, sheet, `.of-button[aria-busy="true"]`)
	if strings.Contains(busy, "opacity: 1") || !strings.Contains(busy, "opacity: 0.5") {
		t.Fatalf("loading button rule = %s", busy)
	}
}

// TestDisabledPaintSkipsHoverFill hovers enabled and disabled controls. A
// disabled control keeps the color it had at rest. An enabled control still
// takes the hover face. switch.css has no hover fill, so both switch tracks stay.
func TestDisabledPaintSkipsHoverFill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p := disabledPaintPage(t)
	p.Window().SetAnimations(false)
	p.Window().SetTransitions(false)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	source := p.HTML()
	if !strings.Contains(source, `id="paint-select-option-1"`) || !strings.Contains(source, `disabled="disabled"`) {
		t.Fatalf("disabled select row lost its id or disabled attribute:\n%s", source)
	}

	cases := []disabledHoverCase{
		{name: "filled enabled", id: "btn-on", dx: 3, dy: 3, wantChange: true},
		{name: "filled disabled", id: "btn-off", dx: 3, dy: 3},
		{name: "filled loading", id: "btn-busy", dx: 3, dy: 3},
		{name: "toggle enabled", id: "toggle-on", dx: 4, dy: 4, wantChange: true},
		{name: "toggle disabled", id: "toggle-off", dx: 4, dy: 4},
		{name: "toggle aria-disabled", id: "toggle-aria", dx: 4, dy: 4},
		{name: "checkbox enabled", id: "check-on", childW: 16, childH: 16, dx: 8, dy: 0, wantChange: true},
		{name: "checkbox disabled", id: "check-off", childW: 16, childH: 16, dx: 8, dy: 0},
		{name: "checkbox aria-disabled", id: "check-aria", childW: 16, childH: 16, dx: 8, dy: 0},
		{name: "radio enabled", id: "paint-radio-item-0", childW: 16, childH: 16, dx: 8, dy: 0, wantChange: true},
		{name: "radio disabled", id: "paint-radio-item-1", childW: 16, childH: 16, dx: 8, dy: 0},
		{name: "radio aria-disabled", id: "radio-aria", childW: 16, childH: 16, dx: 8, dy: 0},
		{name: "switch enabled", id: "switch-on", childW: 36, childH: 20, dx: 4, dy: 10},
		{name: "switch disabled", id: "switch-off", childW: 36, childH: 20, dx: 4, dy: 10},
		{name: "select enabled", id: "paint-select-option-0", dx: 4, dy: 4, wantChange: true},
		{name: "select disabled", id: "paint-select-option-1", dx: 4, dy: 4},
	}
	rest := make(map[string]color.Color, len(cases))
	spots := make(map[string]image.Point, len(cases))
	targets := make(map[string]ownframe.Box, len(cases))
	before := disabledPaintImage(t, ctx, p)
	for _, tc := range cases {
		box := disabledHoverBox(t, p, tc)
		targets[tc.name] = box
		x, y := int(box.X+tc.dx), int(box.Y+tc.dy)
		spots[tc.name] = image.Pt(x, y)
		rest[tc.name] = disabledPaintPixel(t, before, tc.name, x, y)
	}
	if rest["filled disabled"] == rest["filled enabled"] {
		t.Fatalf("disabled button is not faded: %v", rest["filled disabled"])
	}
	if rest["filled loading"] != rest["filled disabled"] {
		t.Fatalf("loading button %v, disabled button %v", rest["filled loading"], rest["filled disabled"])
	}
	for _, tc := range cases {
		target := targets[tc.name]
		if err := p.Edit().Hover(ctx, target.X+target.W/2, target.Y+target.H/2); err != nil {
			t.Fatal(err)
		}
		got := disabledPaintPixel(t, disabledPaintImage(t, ctx, p), tc.name, spots[tc.name].X, spots[tc.name].Y)
		changed := got != rest[tc.name]
		if changed != tc.wantChange {
			t.Fatalf("%s hover color=%v rest=%v change=%v", tc.name, got, rest[tc.name], changed)
		}
	}
}

type disabledHoverCase struct {
	name       string
	id         string
	childW     float64
	childH     float64
	dx         float64
	dy         float64
	wantChange bool
}

func disabledPaintPage(t *testing.T) *ownframe.Page {
	t.Helper()
	root := ui.Div().Children(
		disabledPaintRow(chadcn.Button("Go", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("btn-on")),
		disabledPaintRow(chadcn.Button("Go", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled, Disabled: true}).ID("btn-off")),
		disabledPaintRow(chadcn.Button("Go", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled, Loading: true}).ID("btn-busy")),
		disabledPaintRow(chadcn.Toggle("Go", false).ID("toggle-on")),
		disabledPaintRow(chadcn.Toggle("Go", false, chadcn.ToggleOptions{Disabled: true}).ID("toggle-off")),
		disabledPaintRow(ui.Button("Go").Attr("type", "button").Class("of-toggle of-toggle--medium").Attr("aria-disabled", "true").ID("toggle-aria")),
		disabledPaintRow(chadcn.Checkbox("On", false).ID("check-on")),
		disabledPaintRow(chadcn.CheckboxWithOptions("Off", false, chadcn.CheckboxOptions{Disabled: true}).ID("check-off")),
		disabledPaintRow(ui.Button("").Attr("type", "button").Class("of-checkbox").Attr("aria-disabled", "true").ID("check-aria").
			Children(ui.Span().Class("of-checkbox-box"))),
		disabledPaintRow(chadcn.RadioGroup(chadcn.RadioGroupOptions{
			ID: "paint-radio",
			Items: []chadcn.RadioItem{
				{Value: "a", Label: "Alpha"},
				{Value: "b", Label: "Beta", Disabled: true},
			},
		})),
		disabledPaintRow(ui.Button("").Attr("type", "button").Class("of-radio-group-item").Attr("aria-disabled", "true").ID("radio-aria").
			Children(ui.Span().Class("of-radio-group-indicator"))),
		disabledPaintRow(chadcn.Switch("On", true).ID("switch-on")),
		disabledPaintRow(chadcn.SwitchWithOptions("Off", true, chadcn.SwitchOptions{Disabled: true}).ID("switch-off")),
		disabledPaintRow(chadcn.Select(chadcn.SelectOptions{
			ID:   "paint-select",
			Open: true,
			Items: []chadcn.SelectItem{
				{Value: "a", Label: "Alpha"},
				{Value: "b", Label: "Beta", Disabled: true},
			},
		})),
	)
	p, err := ui.NewPage(ownframe.Config{
		Width:  480,
		Height: 960,
		Theme:  chadcn.Light + disabledPaintVars + chadcn.Stylesheet,
	}, root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func disabledPaintRow(child *ui.Element) *ui.Element {
	return ui.Div().Display(ui.DisplayBlock).Children(child)
}

func disabledHoverBox(t *testing.T, p *ownframe.Page, tc disabledHoverCase) ownframe.Box {
	t.Helper()
	box, ok := pagetest.BoxByID(p, tc.id)
	if !ok {
		t.Fatalf("missing %s", tc.id)
	}
	if tc.childW == 0 {
		return box
	}
	boxes := p.Boxes()
	parent := -1
	for index, candidate := range boxes {
		if candidate.ID == tc.id {
			parent = index
		}
	}
	for _, candidate := range boxes {
		if candidate.ParentIndex != parent || !closeSize(candidate.W, tc.childW) || !closeSize(candidate.H, tc.childH) {
			continue
		}
		return candidate
	}
	t.Fatalf("%s has no %.0fx%.0f child", tc.id, tc.childW, tc.childH)
	return ownframe.Box{}
}

func disabledPaintImage(t *testing.T, ctx context.Context, p *ownframe.Page) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(p.Testing().PNGWithContext(ctx)))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func disabledPaintPixel(t *testing.T, img image.Image, name string, x, y int) color.Color {
	t.Helper()
	if !image.Pt(x, y).In(img.Bounds()) {
		t.Fatalf("%s sample %d,%d outside %v", name, x, y, img.Bounds())
	}
	return color.RGBAModel.Convert(img.At(x, y))
}

func closeSize(got, want float64) bool {
	delta := got - want
	if delta < 0 {
		delta = -delta
	}
	return delta < 1
}
