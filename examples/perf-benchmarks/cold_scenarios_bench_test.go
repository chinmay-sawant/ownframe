package perfbench_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	scenarioWinW   = 1440
	scenarioWinH   = 960
	gridBenchRows  = 30
	gridBenchCols  = 15
	gridBenchCells = gridBenchRows * gridBenchCols
	gridCellW      = 88
	gridCellH      = 24
	gridRowHdrW    = 40
	gridColHdrH    = 24
)

const minimalScenarioCSS = `
html,body{margin:0;padding:0;font-family:sans-serif;font-size:13px;color:#1f2328;background:#ffffff}
.sc-shell{position:relative;display:flex;flex-direction:column;height:100vh;overflow:hidden}
.sc-bar{display:flex;align-items:center;gap:8px;height:32px;padding:0 10px;background:#f3f2f1;border-bottom:1px solid #e1dfdd;flex-shrink:0}
.sc-grid{position:relative;flex:1;overflow:hidden;background:#ffffff}
.sc-canvas,.sc-overlay{position:absolute;left:40px;top:24px;right:0;bottom:0;overflow:hidden}
.sc-cell{position:absolute;box-sizing:border-box;overflow:hidden;white-space:nowrap;padding:0 4px;line-height:22px;border-right:1px solid #d4d4d4;border-bottom:1px solid #d4d4d4}
.sc-active{position:absolute;box-sizing:border-box;border:2px solid #107c41}
.sc-handle{position:absolute;width:6px;height:6px;background:#107c41;border:1px solid #ffffff}
.sc-editor{position:absolute;box-sizing:border-box;padding:0 3px;z-index:5;border:2px solid #107c41;background:#ffffff;color:#1f2328;font-size:13px}
.sc-colhdrs{position:absolute;left:40px;top:0;right:0;height:24px;overflow:hidden;background:#f3f2f1}
.sc-rowhdrs{position:absolute;left:0;top:24px;bottom:0;width:40px;overflow:hidden;background:#f3f2f1}
.sc-colh,.sc-rowh{position:absolute;box-sizing:border-box;text-align:center;font-size:12px;background:#f3f2f1;color:#323130;border-right:1px solid #d4d4d4;border-bottom:1px solid #d4d4d4}
.sc-hdr--on{background:#d2e8db;color:#107c41;font-weight:700}
.sc-row{display:flex;align-items:center;height:32px;padding:0 12px;gap:12px;border-bottom:1px solid #e1dfdd}
.sc-col-a{width:64px}.sc-col-b{width:200px;font-weight:600}.sc-col-c{flex:1;color:#605e5c}
`

const chadcnScenarioExtraCSS = `
.sc-shell{position:relative;display:flex;flex-direction:column;height:100vh;overflow:hidden;background:var(--of-bg);color:var(--of-fg);font-family:sans-serif;font-size:13px}
.sc-bar{display:flex;align-items:center;gap:8px;height:32px;padding:0 10px;background:var(--of-muted);border-bottom:1px solid var(--of-border);flex-shrink:0}
.sc-grid{position:relative;flex:1;overflow:hidden;background:var(--of-bg)}
.sc-canvas,.sc-overlay{position:absolute;left:40px;top:24px;right:0;bottom:0;overflow:hidden}
.sc-cell{position:absolute;box-sizing:border-box;overflow:hidden;white-space:nowrap;padding:0 4px;line-height:22px;color:var(--of-fg);border-right:1px solid var(--of-border);border-bottom:1px solid var(--of-border)}
.sc-active{position:absolute;box-sizing:border-box;border:2px solid var(--of-primary)}
.sc-handle{position:absolute;width:6px;height:6px;background:var(--of-primary);border:1px solid var(--of-bg)}
.sc-editor{position:absolute;box-sizing:border-box;padding:0 3px;z-index:5;border:2px solid var(--of-primary);background:var(--of-bg);color:var(--of-fg);font-size:13px}
.sc-colhdrs{position:absolute;left:40px;top:0;right:0;height:24px;overflow:hidden;background:var(--of-muted)}
.sc-rowhdrs{position:absolute;left:0;top:24px;bottom:0;width:40px;overflow:hidden;background:var(--of-muted)}
.sc-colh,.sc-rowh{position:absolute;box-sizing:border-box;text-align:center;font-size:12px;background:var(--of-muted);color:var(--of-muted-fg);border-right:1px solid var(--of-border);border-bottom:1px solid var(--of-border)}
.sc-hdr--on{background:var(--of-accent);color:var(--of-primary);font-weight:700}
.sc-row{display:flex;align-items:center;height:32px;padding:0 12px;gap:12px;border-bottom:1px solid var(--of-border)}
.sc-col-a{width:64px}.sc-col-b{width:200px;font-weight:600}.sc-col-c{flex:1;color:var(--of-muted-fg)}
.sc-dash{display:flex;flex:1;min-height:0}
.sc-side{width:220px;flex-shrink:0;padding:12px;background:var(--of-muted);border-right:1px solid var(--of-border);display:flex;flex-direction:column;gap:6px}
.sc-main{flex:1;padding:16px;display:flex;flex-direction:column;gap:12px;overflow:hidden}
.sc-cards{display:flex;gap:12px}
.sc-card{flex:1;padding:12px;border-radius:8px;border:1px solid var(--of-border);background:var(--of-card)}
`

func chadcnLightScenarioTheme() string {
	return chadcn.Light + chadcn.Stylesheet + chadcnScenarioExtraCSS
}

func chadcnDarkScenarioTheme() string {
	return chadcn.Dark + chadcn.Stylesheet + chadcnScenarioExtraCSS
}

func px(n int) string { return strconv.Itoa(n) + "px" }

func placeBox(e *ui.Element, x, y, w, h int) *ui.Element {
	return e.CSSRaw("left", px(x)).CSSRaw("top", px(y)).CSSRaw("width", px(w)).CSSRaw("height", px(h))
}

func buildFlexListTree(rows int) *ui.Element {
	body := ui.Div().ID("sc-list")
	for r := range rows {
		rs := strconv.Itoa(r)
		body.Children(
			ui.Div().ID("sc-row-"+rs).Class("sc-row").Children(
				ui.Span().Class("sc-col-a").Text(rs),
				ui.Span().Class("sc-col-b").Text("Record "+rs),
				ui.Span().ID("sc-val-"+rs).Class("sc-col-c").Text("Payload value for row "+rs),
			),
		)
	}
	return ui.Div().ID("sc-shell").Class("sc-shell").Children(
		ui.Div().ID("sc-bar").Class("sc-bar").Text("List Scenario"),
		body,
	)
}

type gridTreeOpts struct {
	scrollRow       int
	selRow          int
	selCol          int
	isolateOverlay  bool
	slotKeyed       bool
	editing         bool
	editVal         string
	mutatedLeafTick int
	mutatedLeafN    int
}

func buildDenseGridTree(opt gridTreeOpts) *ui.Element {
	canvas := ui.Div().ID("sc-canvas").Class("sc-canvas")
	for r := range gridBenchRows {
		dataRow := opt.scrollRow + r
		rs := strconv.Itoa(r)
		drs := strconv.Itoa(dataRow)
		for c := range gridBenchCols {
			cs := strconv.Itoa(c)
			id := "sc-c-" + drs + "-" + cs
			if opt.slotKeyed {
				id = "sc-slot-" + rs + "-" + cs
			}
			val := strconv.Itoa(dataRow*gridBenchCols + c)
			if opt.mutatedLeafN > 0 && r == 0 && c < opt.mutatedLeafN {
				val = strconv.Itoa(opt.mutatedLeafTick*100 + c)
			}
			cell := placeBox(ui.Div().ID(id).Class("sc-cell"), c*gridCellW, r*gridCellH, gridCellW, gridCellH).
				Children(ui.Span().ID(id + "-t").Text(val))
			canvas.Children(cell)
		}
	}

	var overlays []*ui.Element
	ax, ay := opt.selCol*gridCellW, opt.selRow*gridCellH
	if opt.editing {
		in := ui.Input().ID("sc-editor").Class("sc-editor").Attr("type", "text").Attr("value", opt.editVal)
		overlays = append(overlays, placeBox(in, ax, ay, gridCellW, gridCellH))
	} else {
		overlays = append(overlays,
			placeBox(ui.Div().ID("sc-active").Class("sc-active"), ax, ay, gridCellW, gridCellH),
			placeBox(ui.Div().ID("sc-handle").Class("sc-handle"), ax+gridCellW-4, ay+gridCellH-4, 6, 6),
		)
	}

	grid := ui.Div().ID("sc-grid").Class("sc-grid")
	if opt.isolateOverlay {
		grid.Children(canvas, ui.Div().ID("sc-overlay").Class("sc-overlay").Children(overlays...))
	} else {
		canvas.Children(overlays...)
		grid.Children(canvas)
	}

	colhdrs := ui.Div().ID("sc-colhdrs").Class("sc-colhdrs")
	for c := range gridBenchCols {
		cls := "sc-colh"
		if c == opt.selCol {
			cls += " sc-hdr--on"
		}
		colhdrs.Children(placeBox(ui.Div().ID("sc-ch-"+strconv.Itoa(c)).Class(cls).Text("C"+strconv.Itoa(c+1)), c*gridCellW, 0, gridCellW, gridColHdrH))
	}
	rowhdrs := ui.Div().ID("sc-rowhdrs").Class("sc-rowhdrs")
	for r := range gridBenchRows {
		dataRow := opt.scrollRow + r
		cls := "sc-rowh"
		if r == opt.selRow {
			cls += " sc-hdr--on"
		}
		id := "sc-rh-" + strconv.Itoa(dataRow)
		if opt.slotKeyed {
			id = "sc-rh-slot-" + strconv.Itoa(r)
		}
		rowhdrs.Children(placeBox(ui.Div().ID(id).Class(cls).Text(strconv.Itoa(dataRow+1)), 0, r*gridCellH, gridRowHdrW, gridCellH))
	}
	grid.Children(colhdrs, rowhdrs)

	bar := ui.Div().ID("sc-bar").Class("sc-bar").Children(
		ui.Input().ID("sc-name").Attr("type", "text").Attr("value", "R"+strconv.Itoa(opt.scrollRow+opt.selRow+1)+"C"+strconv.Itoa(opt.selCol+1)),
		ui.Input().ID("sc-fx").Attr("type", "text").Attr("value", strconv.Itoa((opt.scrollRow+opt.selRow)*gridBenchCols+opt.selCol)),
	)
	return ui.Div().ID("sc-shell").Class("sc-shell").Children(bar, grid)
}

func buildNestedDashboardTree() *ui.Element {
	side := ui.Div().ID("sc-side").Class("sc-side")
	for i := range 12 {
		side.Children(ui.Button("Section " + strconv.Itoa(i+1)).Class("of-btn of-btn-ghost"))
	}
	cards := ui.Div().Class("sc-cards")
	for i := range 4 {
		cards.Children(ui.Div().Class("sc-card").Children(
			ui.Div().Text("Metric "+strconv.Itoa(i+1)),
			ui.Div().Text("$"+strconv.Itoa((i+1)*12450)),
		))
	}
	table := ui.Div().ID("sc-dash-table")
	for r := range 40 {
		rs := strconv.Itoa(r)
		table.Children(ui.Div().ID("sc-dr-"+rs).Class("sc-row").Children(
			ui.Span().Class("sc-col-a").Text("#"+rs),
			ui.Span().Class("sc-col-b").Text("Order "+rs),
			ui.Span().Class("sc-col-c").Text("Processed"),
		))
	}
	return ui.Div().ID("sc-shell").Class("sc-shell").Children(
		ui.Div().ID("sc-bar").Class("sc-bar").Text("Dashboard"),
		ui.Div().Class("sc-dash").Children(
			side,
			ui.Div().Class("sc-main").Children(cards, table),
		),
	)
}

func mountScenarioView(ctx context.Context, b *testing.B, theme string, root *ui.Element) *ui.View {
	b.Helper()
	v, err := ui.Mount(ownframe.Config{
		Title:       "Scenario",
		Width:       scenarioWinW,
		Height:      scenarioWinH,
		Theme:       theme,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, root)
	if err != nil {
		b.Fatal(err)
	}
	if err := v.Page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	return v
}

// BenchmarkColdMount compares first-paint cost across stylesheet weights and
// DOM shapes (flex list, 450-cell positioned grid, nested dashboard).
func BenchmarkColdMount(b *testing.B) {
	ctx := b.Context()
	cases := []struct {
		name  string
		theme string
		build func() *ui.Element
	}{
		{"MinimalCSS_FlexList_200", minimalScenarioCSS, func() *ui.Element { return buildFlexListTree(200) }},
		{"ChadcnCSS_FlexList_200", chadcnLightScenarioTheme(), func() *ui.Element { return buildFlexListTree(200) }},
		{"MinimalCSS_GridCells_450", minimalScenarioCSS, func() *ui.Element { return buildDenseGridTree(gridTreeOpts{}) }},
		{"ChadcnCSS_GridCells_450", chadcnLightScenarioTheme(), func() *ui.Element { return buildDenseGridTree(gridTreeOpts{}) }},
		{"ChadcnCSS_Dashboard_Nested", chadcnLightScenarioTheme(), buildNestedDashboardTree},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				v, err := ui.Mount(ownframe.Config{
					Title:       tc.name,
					Width:       scenarioWinW,
					Height:      scenarioWinH,
					Theme:       tc.theme,
					RetainState: true,
					Scrollbar:   ownframe.ScrollbarHidden,
				}, tc.build())
				if err != nil {
					b.Fatal(err)
				}
				if err := v.Page.Redraw(ctx); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPatchEscalation isolates the Excel selection bottleneck: moving 2
// positioned overlays inside a 450-cell container vs. inside an isolated
// sibling overlay container, plus paint-only header toggles and leaf text edits.
func BenchmarkPatchEscalation(b *testing.B) {
	ctx := b.Context()
	theme := chadcnLightScenarioTheme()

	b.Run("SharedContainer_2GeoMoves_450Cells", func(b *testing.B) {
		v := mountScenarioView(ctx, b, theme, buildDenseGridTree(gridTreeOpts{selRow: 1, selCol: 1}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			opt := gridTreeOpts{selRow: 1 + (i % 10), selCol: 1 + ((i / 10) % 6)}
			if err := v.Update(ctx, buildDenseGridTree(opt)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("IsolatedOverlay_2GeoMoves_450Cells", func(b *testing.B) {
		v := mountScenarioView(ctx, b, theme, buildDenseGridTree(gridTreeOpts{selRow: 1, selCol: 1, isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			opt := gridTreeOpts{selRow: 1 + (i % 10), selCol: 1 + ((i / 10) % 6), isolateOverlay: true}
			if err := v.Update(ctx, buildDenseGridTree(opt)); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LeafText_10Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, theme, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, slotKeyed: true, mutatedLeafN: 10}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			opt := gridTreeOpts{isolateOverlay: true, slotKeyed: true, mutatedLeafN: 10, mutatedLeafTick: i + 1}
			if err := v.Update(ctx, buildDenseGridTree(opt)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkScrollPatterns compares coordinate-keyed virtualized cells (forces
// renderFull) against slot-keyed virtualized cells (stays in diffTree).
func BenchmarkScrollPatterns(b *testing.B) {
	ctx := b.Context()
	theme := chadcnLightScenarioTheme()

	b.Run("CoordKeyed_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, theme, buildDenseGridTree(gridTreeOpts{isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{scrollRow: i % 20, isolateOverlay: true})); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SlotKeyed_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, theme, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, slotKeyed: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{scrollRow: i % 20, isolateOverlay: true, slotKeyed: true})); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkOverlayAndFocus measures structural overlay mount/unmount,
// back-to-back Update+Focus (like Excel beginEdit/commitEdit), and theme toggle.
func BenchmarkOverlayAndFocus(b *testing.B) {
	ctx := b.Context()
	light := chadcnLightScenarioTheme()
	dark := chadcnDarkScenarioTheme()

	b.Run("MountUnmountOverlay_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, light, buildDenseGridTree(gridTreeOpts{isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: true, editVal: strconv.Itoa(i)})); err != nil {
				b.Fatal(err)
			}
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: false})); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RedrawThenFocus_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, light, buildDenseGridTree(gridTreeOpts{isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: true, editVal: strconv.Itoa(i)})); err != nil {
				b.Fatal(err)
			}
			if err := v.Page.Form().Focus(ctx, "sc-editor"); err != nil {
				b.Fatal(err)
			}
			if err := v.Page.Form().Focus(ctx, ""); err != nil {
				b.Fatal(err)
			}
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: false})); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RequestFocusWithRedraw_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, light, buildDenseGridTree(gridTreeOpts{isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			v.Page.Form().RequestFocus("sc-editor")
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: true, editVal: strconv.Itoa(i)})); err != nil {
				b.Fatal(err)
			}
			v.Page.Form().RequestFocus("")
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true, editing: false})); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ThemeToggle_450Cells_Chadcn", func(b *testing.B) {
		v := mountScenarioView(ctx, b, light, buildDenseGridTree(gridTreeOpts{isolateOverlay: true}))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			th := dark
			if i%2 == 1 {
				th = light
			}
			if err := v.Page.SetTheme(th); err != nil {
				b.Fatal(err)
			}
			if err := v.Update(ctx, buildDenseGridTree(gridTreeOpts{isolateOverlay: true})); err != nil {
				b.Fatal(err)
			}
		}
	})
}
