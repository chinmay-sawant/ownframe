package chadcn_test

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func TestThemesReachRenderedCard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	card := chadcn.Card(ui.Text("Themed surface")).ID("card").Width(ui.Px(200)).Height(ui.Px(120))
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320}, card)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, sheet string
		want        color.RGBA
	}{
		{"fallback", chadcn.Stylesheet, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"light", theme.Light + chadcn.Stylesheet, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"dark", theme.Dark + chadcn.Stylesheet, color.RGBA{R: 28, G: 39, B: 56, A: 255}},
		{"custom", theme.Dark + `:root{--of-surface:#123456}` + chadcn.Stylesheet, color.RGBA{R: 18, G: 52, B: 86, A: 255}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := view.Page.SetTheme(tt.sheet); err != nil {
				t.Fatal(err)
			}
			if err := view.Page.Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(view.Page.Testing().PNGWithContext(ctx)))
			if err != nil {
				t.Fatal(err)
			}
			got := color.RGBAModel.Convert(img.At(30, 90))
			if got != tt.want {
				t.Fatalf("rendered surface=%v want %v", got, tt.want)
			}
		})
	}
}

func TestDialogFocusAndDisabledButton(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	hits := 0
	root := func(open bool) *ui.Element {
		tree := ui.Div().Children(chadcn.Button("Open", chadcn.ButtonOptions{}).ID("open"),
			chadcn.Button("Disabled", chadcn.ButtonOptions{Disabled: true}).ID("disabled").OnClick(func(context.Context) error { hits++; return nil }))
		if open {
			tree.Children(chadcn.Dialog(chadcn.DialogOptions{
				BackdropID: "backdrop", PanelID: "modal", Label: "Confirm", CloseAction: "close", Width: ui.Px(240),
			}, chadcn.Button("Confirm", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("confirm"), chadcn.Button("Cancel", chadcn.ButtonOptions{}).ID("cancel")))
		}
		return tree
	}
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320}, root(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "disabled"); err != nil {
		t.Fatal(err)
	}
	if hits != 0 {
		t.Fatal("disabled callback fired")
	}
	if err := view.Page.Form().Focus(ctx, "open"); err != nil {
		t.Fatal(err)
	}
	var focus chadcn.DialogFocus
	focus.Activate(view.Page, "confirm", "cancel")
	// Exercise the same deferred draw behavior as native pointer input.
	view.Page.Probe().BeginRedrawBatch()
	if err := view.Update(ctx, root(true)); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Probe().EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FocusID(); got != "confirm" {
		t.Fatalf("initial focus=%q", got)
	}
	for _, want := range []string{"cancel", "confirm"} {
		if err := view.Page.Form().FocusNext(ctx); err != nil {
			t.Fatal(err)
		}
		if got := view.Page.Form().FocusID(); got != want {
			t.Fatalf("Tab focus=%q want %q", got, want)
		}
	}
	if err := view.Page.Form().FocusPrev(ctx); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FocusID(); got != "cancel" {
		t.Fatalf("Shift+Tab focus=%q", got)
	}
	focus.Deactivate(view.Page)
	if err := view.Update(ctx, root(false)); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FocusID(); got != "open" {
		t.Fatalf("restored focus=%q", got)
	}
}

func TestButtonHasUsablePadding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320, Theme: chadcn.Stylesheet}, chadcn.Button("Save", chadcn.ButtonOptions{}).ID("save"))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	box, ok := pagetest.BoxByID(p, "save")
	if !ok {
		t.Fatal("missing button")
	}
	if box.H < 30 || box.W < 45 {
		t.Fatalf("button padding was not laid out: %.0fx%.0f", box.W, box.H)
	}
}

func TestOverridePriorityAcrossGlobalThemes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	appOverrides := theme.Overrides{"--of-primary": "#123456"}
	appSheet, err := appOverrides.Stylesheet()
	if err != nil {
		t.Fatal(err)
	}
	scoped := theme.Overrides{"--of-primary": "#345678"}
	local := theme.Overrides{"--of-primary": "#56789a"}
	root := ui.Div().Children(
		chadcn.Button("Global", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("global"),
		scoped.Apply(chadcn.Panel(
			chadcn.Button("Scoped", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("scoped"),
			chadcn.Button("Local", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled, Overrides: local}).ID("local"),
		)),
	)
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320}, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, base := range []string{theme.Light, theme.Dark} {
		if err := p.SetTheme(base + appSheet + chadcn.Stylesheet); err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(p.Testing().PNGWithContext(ctx)))
		if err != nil {
			t.Fatal(err)
		}
		for id, want := range map[string]color.RGBA{
			"global": {R: 18, G: 52, B: 86, A: 255}, "scoped": {R: 52, G: 86, B: 120, A: 255}, "local": {R: 86, G: 120, B: 154, A: 255},
		} {
			box, ok := pagetest.BoxByID(p, id)
			if !ok {
				t.Fatalf("missing %s", id)
			}
			got := color.RGBAModel.Convert(img.At(int(box.X)+5, int(box.Y+box.H/2)))
			if got != want {
				t.Fatalf("%s background=%v want %v", id, got, want)
			}
		}
	}
}

func TestIconClickReachesAction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320, Theme: chadcn.Stylesheet}, chadcn.Button("Settings", chadcn.ButtonOptions{
		Icon: ui.Text("⚙"),
	}).Attr("data-action", "settings"))
	if err != nil {
		t.Fatal(err)
	}
	clicks := 0
	p.Handle(ownframe.Handlers{Click: func(_ context.Context, box ownframe.Box) error {
		if box.Action == "settings" {
			clicks++
		}
		return nil
	}})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, box := range p.Boxes() {
		if box.Tag == "span" && box.Text == "⚙" {
			found = true
			if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if !found {
		t.Fatal("missing icon box")
	}
	if clicks != 1 {
		t.Fatalf("icon click reached %d actions", clicks)
	}
}
