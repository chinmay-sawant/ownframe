package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

// The overlay layout tests mount the real components and measure their boxes.
// They cover the engine rules the stylesheet relies on: a stretched side sheet
// needs a zero body basis, and a body that must hug short content cannot be a
// flex container with max-height.

var overlayLayoutSizes = []struct {
	name string
	w, h int
}{
	{"desktop", 1440, 960},
	{"minimum", 1100, 720},
}

func overlayLayoutMount(t *testing.T, w, h int, root *ui.Element) *ui.View {
	t.Helper()
	view, err := ui.Mount(ownframe.Config{
		Width:  w,
		Height: h,
		Theme:  chadcn.Light + chadcn.Stylesheet,
	}, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return view
}

func overlayLayoutBox(t *testing.T, view *ui.View, id string) ownframe.Box {
	t.Helper()
	box, ok := pagetest.BoxByID(view.Page, id)
	if !ok {
		t.Fatalf("missing box %q", id)
	}
	return box
}

func overlayLayoutSheet(tall bool) *ui.Element {
	height := 40
	if tall {
		height = 1200
	}
	return chadcn.Sheet(chadcn.SheetOptions{
		ID:          "layout-sheet",
		Open:        true,
		Side:        chadcn.SheetRight,
		Title:       "Project settings",
		Description: "Notifications, routing, and verification.",
		Footer: []*ui.Element{
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonText}).ID("layout-sheet-cancel"),
			chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("layout-sheet-save"),
		},
		OnOpenChange: func(context.Context, bool) error { return nil },
	},
		ui.Div().ID("layout-sheet-top").Text("top"),
		ui.Div().ID("layout-sheet-filler").Height(ui.Px(float64(height))).Text("filler"),
	)
}

// TestOverlayLayoutSheetKeepsFooterInside checks the right sheet is 400px wide,
// reaches the viewport edge, and keeps its Save and Cancel row inside the
// panel under a body that is taller than the space left for it.
func TestOverlayLayoutSheetKeepsFooterInside(t *testing.T) {
	t.Parallel()
	for _, size := range overlayLayoutSizes {
		t.Run(size.name, func(t *testing.T) {
			view := overlayLayoutMount(t, size.w, size.h, overlayLayoutSheet(true))
			panel := overlayLayoutBox(t, view, "layout-sheet-panel")
			if panel.W != 400 {
				t.Fatalf("panel width=%.0f, want 400", panel.W)
			}
			if panel.X+panel.W != float64(size.w) {
				t.Fatalf("panel right edge=%.0f, want %d", panel.X+panel.W, size.w)
			}
			if panel.Y != 0 || panel.Y+panel.H > float64(size.h) {
				t.Fatalf("panel y=%.0f h=%.0f leaves the %dpx viewport", panel.Y, panel.H, size.h)
			}
			cancel := overlayLayoutBox(t, view, "layout-sheet-cancel")
			save := overlayLayoutBox(t, view, "layout-sheet-save")
			if save.Y+save.H >= panel.Y+panel.H {
				t.Fatalf("save bottom=%.0f, panel bottom=%.0f: the footer is clipped", save.Y+save.H, panel.Y+panel.H)
			}
			if cancel.Y <= panel.Y {
				t.Fatalf("cancel y=%.0f sits above the panel", cancel.Y)
			}
			filler := overlayLayoutBox(t, view, "layout-sheet-filler")
			if filler.Y+filler.H >= cancel.Y {
				t.Fatalf("body filler bottom=%.0f reaches the footer at %.0f", filler.Y+filler.H, cancel.Y)
			}
			if filler.H >= 1200 {
				t.Fatalf("body filler kept its full %.0fpx height instead of scrolling", filler.H)
			}
			backdrop := overlayLayoutBox(t, view, "layout-sheet-backdrop")
			if backdrop.X != 0 || backdrop.Y != 0 || backdrop.W != float64(size.w) || backdrop.H != float64(size.h) {
				t.Fatalf("backdrop = %.0f,%.0f %.0fx%.0f, want the whole viewport", backdrop.X, backdrop.Y, backdrop.W, backdrop.H)
			}
		})
	}
}

func overlayLayoutDialog(tall bool) *ui.Element {
	height := 40
	if tall {
		height = 1200
	}
	return chadcn.Dialog(chadcn.DialogOptions{
		ID:           "layout-dialog",
		Open:         true,
		Title:        "Create alert rule",
		Description:  "Notify the channel that owns this service.",
		Width:        ui.Px(620),
		Footer:       []*ui.Element{chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("layout-dialog-save")},
		OnOpenChange: func(context.Context, bool) error { return nil },
	},
		ui.Div().ID("layout-dialog-top").Text("top"),
		ui.Div().ID("layout-dialog-filler").Height(ui.Px(float64(height))).Text("filler"),
	)
}

// TestOverlayLayoutDialogScrollsAboveFooter checks a tall dialog body clips
// above a footer that stays inside the panel at both window sizes.
func TestOverlayLayoutDialogScrollsAboveFooter(t *testing.T) {
	t.Parallel()
	for _, size := range overlayLayoutSizes {
		t.Run(size.name, func(t *testing.T) {
			view := overlayLayoutMount(t, size.w, size.h, overlayLayoutDialog(true))
			panel := overlayLayoutBox(t, view, "layout-dialog-panel")
			if panel.Y < 0 || panel.Y+panel.H > float64(size.h) {
				t.Fatalf("panel y=%.0f h=%.0f leaves the %dpx viewport", panel.Y, panel.H, size.h)
			}
			save := overlayLayoutBox(t, view, "layout-dialog-save")
			if save.Y+save.H >= panel.Y+panel.H {
				t.Fatalf("save bottom=%.0f, panel bottom=%.0f: the footer is clipped", save.Y+save.H, panel.Y+panel.H)
			}
			if panel.Y+panel.H-save.Y > 80 {
				t.Fatalf("save y=%.0f leaves %.0fpx of empty panel below it", save.Y, panel.Y+panel.H-save.Y)
			}
			filler := overlayLayoutBox(t, view, "layout-dialog-filler")
			if filler.Y <= panel.Y {
				t.Fatalf("body filler y=%.0f starts above the panel", filler.Y)
			}
			if filler.Y+filler.H <= save.Y {
				t.Fatalf("body filler ends at %.0f before the footer at %.0f: the body is not clipped", filler.Y+filler.H, save.Y)
			}
			if panel.H > float64(size.h)-24 {
				t.Fatalf("panel h=%.0f at %dpx tall, want the body cap to keep it in the viewport", panel.H, size.h)
			}
		})
	}
}

// TestOverlayLayoutDialogHugsShortBody checks a dialog with a little content
// does not paint at the old 85vh cap with empty panel under the actions.
func TestOverlayLayoutDialogHugsShortBody(t *testing.T) {
	t.Parallel()
	view := overlayLayoutMount(t, 1440, 960, overlayLayoutDialog(false))
	panel := overlayLayoutBox(t, view, "layout-dialog-panel")
	if panel.H > 320 {
		t.Fatalf("panel h=%.0f for a short body, want it to hug the content", panel.H)
	}
	save := overlayLayoutBox(t, view, "layout-dialog-save")
	if panel.Y+panel.H-save.Y > 80 {
		t.Fatalf("save y=%.0f leaves %.0fpx of empty panel below it", save.Y, panel.Y+panel.H-save.Y)
	}
}

func overlayLayoutDrawer(tall bool) *ui.Element {
	height := 40
	if tall {
		height = 1200
	}
	return chadcn.Drawer(chadcn.DrawerOptions{
		ID:          "layout-drawer",
		Open:        true,
		Title:       "Log detail",
		Description: "api-gateway · 12:00 · status 200",
		Footer: []*ui.Element{
			chadcn.Button("Close", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).ID("layout-drawer-close"),
		},
		OnOpenChange: func(context.Context, bool) error { return nil },
	},
		ui.Div().ID("layout-drawer-top").Text("top"),
		ui.Div().ID("layout-drawer-filler").Height(ui.Px(float64(height))).Text("filler"),
	)
}

// TestOverlayLayoutDrawerSitsOnBottomEdge checks the drawer is full width and
// bottom anchored, its footer reaches the bottom edge, and a tall body clips
// above it.
func TestOverlayLayoutDrawerSitsOnBottomEdge(t *testing.T) {
	t.Parallel()
	for _, size := range overlayLayoutSizes {
		for _, tall := range []bool{true, false} {
			name := size.name + "/short"
			if tall {
				name = size.name + "/tall"
			}
			t.Run(name, func(t *testing.T) {
				view := overlayLayoutMount(t, size.w, size.h, overlayLayoutDrawer(tall))
				panel := overlayLayoutBox(t, view, "layout-drawer-panel")
				if panel.X != 0 || panel.W != float64(size.w) {
					t.Fatalf("panel spans %.0f..%.0f, want the full %dpx width", panel.X, panel.X+panel.W, size.w)
				}
				if panel.Y+panel.H != float64(size.h) {
					t.Fatalf("panel bottom=%.0f, want the %dpx viewport bottom", panel.Y+panel.H, size.h)
				}
				close := overlayLayoutBox(t, view, "layout-drawer-close")
				if close.Y+close.H >= panel.Y+panel.H {
					t.Fatalf("close bottom=%.0f, panel bottom=%.0f: the footer is clipped", close.Y+close.H, panel.Y+panel.H)
				}
				if panel.Y+panel.H-close.Y > 80 {
					t.Fatalf("close y=%.0f leaves %.0fpx of empty panel below it", close.Y, panel.Y+panel.H-close.Y)
				}
				filler := overlayLayoutBox(t, view, "layout-drawer-filler")
				if tall && panel.H > float64(size.h)-24 {
					t.Fatalf("panel h=%.0f at %dpx tall, want the body cap to keep it in the viewport", panel.H, size.h)
				}
				if tall && filler.Y+filler.H <= close.Y {
					t.Fatalf("body filler ends at %.0f before the footer at %.0f: the body is not clipped", filler.Y+filler.H, close.Y)
				}
				if !tall && panel.H > 320 {
					t.Fatalf("panel h=%.0f for a short body, want it to hug the content", panel.H)
				}
				backdrop := overlayLayoutBox(t, view, "layout-drawer-backdrop")
				if backdrop.H != float64(size.h) || backdrop.W != float64(size.w) {
					t.Fatalf("backdrop = %.0fx%.0f, want the whole viewport", backdrop.W, backdrop.H)
				}
			})
		}
	}
}

// TestOverlayLayoutModalBandMatches checks the sheet joins the dialog and
// drawer on the shared modal layer, and that no overlay carries an inline
// z-index that would beat it.
func TestOverlayLayoutModalBandMatches(t *testing.T) {
	t.Parallel()
	const band = "z-index: var(--of-modal-z, 55);"
	for _, selector := range []string{".of-sheet-backdrop", ".of-dialog-backdrop", ".of-drawer-backdrop"} {
		rule := overlayLayoutRule(t, chadcn.Stylesheet, selector)
		if !strings.Contains(rule, band) {
			t.Fatalf("%s is outside the modal band %q:\n%s", selector, band, rule)
		}
	}
	for name, root := range map[string]*ui.Element{
		"sheet":  overlayLayoutSheet(true),
		"dialog": overlayLayoutDialog(true),
		"drawer": overlayLayoutDrawer(true),
	} {
		markup, err := ui.Render(root)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(markup, "z-index") {
			t.Fatalf("%s carries an inline z-index that beats the modal layer:\n%s", name, markup)
		}
	}
}

// TestOverlayLayoutBodyUsesZeroBasis checks the sheet body takes the leftover
// height instead of sizing the panel from its content.
func TestOverlayLayoutBodyUsesZeroBasis(t *testing.T) {
	t.Parallel()
	body := overlayLayoutRule(t, chadcn.Stylesheet, ".of-sheet-body")
	for _, want := range []string{"flex-basis: 0px;", "min-height: 0;", "overflow-y: auto;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("sheet body is missing %q:\n%s", want, body)
		}
	}
}

// TestOverlayLayoutFooterStaysClickable checks a clipped body never covers the
// footer's hit map: a click on the Save control still reaches its action.
func TestOverlayLayoutFooterStaysClickable(t *testing.T) {
	t.Parallel()
	build := func(t *testing.T) (*ui.View, *bool) {
		t.Helper()
		hit := false
		root := chadcn.Dialog(chadcn.DialogOptions{
			ID:           "layout-hit",
			Open:         true,
			Title:        "Create alert rule",
			Description:  "Notify the channel that owns this service.",
			Width:        ui.Px(620),
			Footer:       []*ui.Element{chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("layout-hit-save").OnClick(func(context.Context) error { hit = true; return nil })},
			OnOpenChange: func(context.Context, bool) error { return nil },
		},
			ui.Div().ID("layout-hit-filler").Height(ui.Px(1200)).Text("filler"),
		)
		return overlayLayoutMount(t, 1100, 720, root), &hit
	}
	view, hit := build(t)
	if err := pagetest.Click(context.Background(), view.Page, "layout-hit-save"); err != nil {
		t.Fatal(err)
	}
	if !*hit {
		t.Fatal("a click on Save was swallowed by the overflowing body")
	}
}

// TestOverlayLayoutEdgeSheetsHugContent checks the top and bottom sheets keep
// their footer on the panel's edge for a short and a tall body, instead of
// painting an empty strip under it.
func TestOverlayLayoutEdgeSheetsHugContent(t *testing.T) {
	t.Parallel()
	for _, side := range []chadcn.SheetSide{chadcn.SheetTop, chadcn.SheetBottom} {
		for _, tall := range []bool{false, true} {
			name := "top/short"
			if side == chadcn.SheetBottom {
				name = "bottom/short"
			}
			if tall {
				name = strings.Replace(name, "short", "tall", 1)
			}
			t.Run(name, func(t *testing.T) {
				height := 40
				if tall {
					height = 1200
				}
				root := chadcn.Sheet(chadcn.SheetOptions{
					ID:           "layout-edge",
					Open:         true,
					Side:         side,
					Title:        "Filters",
					Description:  "Narrow the stream.",
					Footer:       []*ui.Element{chadcn.Button("Apply", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("layout-edge-apply")},
					OnOpenChange: func(context.Context, bool) error { return nil },
				}, ui.Div().ID("layout-edge-top").Text("top"), ui.Div().ID("layout-edge-filler").Height(ui.Px(float64(height))).Text("filler"))
				view := overlayLayoutMount(t, 1440, 960, root)
				panel := overlayLayoutBox(t, view, "layout-edge-panel")
				apply := overlayLayoutBox(t, view, "layout-edge-apply")
				if apply.Y+apply.H >= panel.Y+panel.H {
					t.Fatalf("apply bottom=%.0f, panel bottom=%.0f: the footer is clipped", apply.Y+apply.H, panel.Y+panel.H)
				}
				if panel.Y+panel.H-apply.Y > 80 {
					t.Fatalf("apply y=%.0f leaves %.0fpx of empty panel below it", apply.Y, panel.Y+panel.H-apply.Y)
				}
				if tall && panel.H > 960-24 {
					t.Fatalf("panel h=%.0f, want the body cap to keep it in the viewport", panel.H)
				}
				if !tall && panel.H > 320 {
					t.Fatalf("panel h=%.0f for a short body, want it to hug the content", panel.H)
				}
			})
		}
	}
}

// overlayLayoutRule returns one rule body from the stylesheet, braces kept.
// The marker must start a selector line, so a scoped rule such as
// ".of-sheet--top .of-sheet-body" does not answer for ".of-sheet-body".
func overlayLayoutRule(t *testing.T, sheet, selector string) string {
	t.Helper()
	start := strings.Index(sheet, "\n"+selector+" {")
	if start < 0 {
		t.Fatalf("stylesheet is missing %s", selector)
	}
	start++
	end := strings.Index(sheet[start:], "}")
	if end < 0 {
		t.Fatalf("stylesheet rule %s is unterminated", selector)
	}
	return sheet[start : start+end+1]
}
