package main

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

// readmeTokens are the --xl-* tokens the plan README lists in its primary
// token table.
var readmeTokens = []string{
	"brand", "brand-hover", "brand-strong", "brand-tint", "titlebar-bg", "titlebar-fg",
	"ribbon-bg", "ribbon-group-bg", "ribbon-border", "formula-bg", "sheet-bg", "cell-text", "cell-muted",
	"gridline", "gridline-frozen", "header-bg", "header-fg", "header-active-bg", "header-active-fg",
	"selection-fill", "selection-border", "cut-dash", "tabbar-bg", "tab-active-bg", "statusbar-bg",
	"statusbar-fg", "error-fg", "comment-corner", "negative-num",
}

// ---- EXL-05 and EXL-06: token sheets ----

func TestDThemeTokensPresent(t *testing.T) {
	hex := regexp.MustCompile(`^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)
	for name, sheet := range map[string]string{"ExcelLight": ExcelLight, "ExcelDark": ExcelDark} {
		check := func(token string) {
			m := regexp.MustCompile(`--xl-` + regexp.QuoteMeta(token) + `:\s*([^;]+);`).FindStringSubmatch(sheet)
			if m == nil {
				t.Errorf("%s misses --xl-%s", name, token)
			} else if !hex.MatchString(strings.TrimSpace(m[1])) {
				t.Errorf("%s --xl-%s = %q is not a hex color", name, token, m[1])
			}
		}
		for _, tok := range readmeTokens {
			check(tok)
		}
		for i := 1; i <= 6; i++ {
			check("ref-" + itoa(i))
			check("ref-" + itoa(i) + "-fill")
		}
		for _, tok := range []string{"cf-red-bg", "cf-red-fg", "cf-yellow-bg", "cf-yellow-fg", "cf-green-bg", "cf-green-fg",
			"bar-blue", "bar-green", "bar-red", "bar-amber", "scale-low", "scale-mid", "scale-high"} {
			check(tok)
		}
	}
	// The sheets sit on top of the chadcn base and mount without error.
	for _, dark := range []bool{false, true} {
		a, _ := newTestApp(t)
		if dark {
			a.toggleTheme()
		}
		if a.lastErr != nil {
			t.Fatalf("dark=%v: %v", dark, a.lastErr)
		}
	}
	// Spot-check values from the README table.
	for _, want := range []string{"--xl-brand:#107c41", "--xl-gridline:#d4d4d4", "--xl-ref-3:#7c3aed"} {
		if !strings.Contains(ExcelLight, want) {
			t.Errorf("ExcelLight misses %s", want)
		}
	}
	for _, want := range []string{"--xl-brand:#185c37", "--xl-sheet-bg:#141417", "--xl-ref-3:#a78bfa"} {
		if !strings.Contains(ExcelDark, want) {
			t.Errorf("ExcelDark misses %s", want)
		}
	}
}

func TestDSwatchMatrixGeometry(t *testing.T) {
	hex := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	m := themeMatrix()
	if len(m) != 5 {
		t.Fatalf("matrix has %d rows", len(m))
	}
	for r, row := range m {
		if len(row) != 10 {
			t.Fatalf("row %d has %d columns", r, len(row))
		}
		for c, col := range row {
			if !hex.MatchString(col) {
				t.Errorf("swatch %d,%d = %q", r, c, col)
			}
		}
	}
	if len(officeStandardColors()) != 10 {
		t.Fatalf("standard colors: %d", len(officeStandardColors()))
	}
	for _, col := range officeStandardColors() {
		if !hex.MatchString(col) {
			t.Errorf("standard color %q", col)
		}
	}
}

// ---- EXL-07: runtime theme switch keeps workbook state ----

func TestDThemeSwitchKeepsState(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetRaw(3, 3, "keep")
	a.selectRange(Range{Min: CellRef{Row: 2, Col: 2}, Max: CellRef{Row: 4, Col: 4}})
	a.scrollRow = 20
	a.redraw()
	a.toggleTheme()
	a.toggleTheme()
	if a.sheet().Raw(3, 3) != "keep" || a.sel.Max != (CellRef{Row: 4, Col: 4}) || a.scrollRow != 20 {
		t.Fatalf("state after two toggles: raw %q sel %v scrollRow %d", a.sheet().Raw(3, 3), a.sel, a.scrollRow)
	}
}

// ---- EXL-03: title bar actions ----

func TestDTitleBarActions(t *testing.T) {
	a, _ := newTestApp(t)
	html := a.page.HTML()
	for _, want := range []string{`data-action="qat:save"`, `data-action="qat:undo"`, `data-action="qat:redo"`,
		`data-action="search:open"`, `data-action="theme:toggle"`, `data-action="win-min"`,
		`data-action="win-max"`, `data-action="win-close"`, "Quarterly_Model.xlsx"} {
		if !strings.Contains(html, want) {
			t.Errorf("title bar misses %s", want)
		}
	}
	ctx := context.Background()
	if err := a.handleClick(ctx, ownframe.Box{Action: "search:open"}); err != nil || a.overlay != "command" {
		t.Fatalf("search: overlay %q err %v", a.overlay, err)
	}
	a.setOverlay("")
	if err := a.handleClick(ctx, ownframe.Box{Action: "qat:save"}); err != nil || a.overlay != "save" {
		t.Fatalf("save: overlay %q err %v", a.overlay, err)
	}
	a.setOverlay("")
	a.wb.Edit(a.sheet(), 0, 0, "undo me")
	if err := a.handleClick(ctx, ownframe.Box{Action: "qat:undo"}); err != nil || a.sheet().Raw(0, 0) == "undo me" {
		t.Fatalf("undo left %q (err %v)", a.sheet().Raw(0, 0), err)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "qat:redo"}); err != nil || a.sheet().Raw(0, 0) != "undo me" {
		t.Fatalf("redo left %q (err %v)", a.sheet().Raw(0, 0), err)
	}
}

// ---- EXL-01: the drag guard yields to overlays ----

func TestDDragGuardYieldsToOverlays(t *testing.T) {
	a, _ := newTestApp(t)
	if a.overlaysOpen() {
		t.Fatal("overlay open at start")
	}
	a.setOverlay("shortcuts")
	if !a.overlaysOpen() {
		t.Fatal("overlaysOpen false with a dialog open")
	}
	a.setOverlay("")
	if a.overlaysOpen() {
		t.Fatal("overlaysOpen true after close")
	}
}

// ---- EXL-12: ribbon tab switch and collapse ----

func TestDRibbonTabSwitchAndCollapse(t *testing.T) {
	a, _ := newTestApp(t)
	grid0 := mustBox(t, a, "xl-grid")
	a.runAction("xl:tab:Insert")
	if a.ribbonTab != "Insert" || !strings.Contains(a.page.HTML(), `aria-selected="true"`) {
		t.Fatalf("tab %q", a.ribbonTab)
	}
	slot := mustBox(t, a, "xl-ribbon-slot")
	if int(slot.H) != ribbonHeightFull {
		t.Fatalf("expanded ribbon %v", slot.H)
	}
	a.runAction("xl:ribbon:toggle")
	slot = mustBox(t, a, "xl-ribbon-slot")
	grid1 := mustBox(t, a, "xl-grid")
	if int(slot.H) != ribbonHeightMin || int(grid1.H-grid0.H) != ribbonHeightFull-ribbonHeightMin {
		t.Fatalf("collapsed ribbon %v, grid grew by %v", slot.H, grid1.H-grid0.H)
	}
	a.runAction("xl:ribbon:toggle")
	if grid2 := mustBox(t, a, "xl-grid"); grid2.H != grid0.H {
		t.Fatalf("grid height %v after expanding, want %v", grid2.H, grid0.H)
	}
}

// ---- EXL-15: node budget across a sheet that fills the grid ----

func TestDGridNodeBudgetFullSheet(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	for r := 0; r < MaxRows; r += 7 {
		for c := 0; c < MaxCols; c += 3 {
			sh.SetRaw(r, c, "v")
		}
	}
	a.redraw()
	n := 0
	for _, b := range a.page.Boxes() {
		if strings.HasPrefix(b.ID, "xl-gc-") {
			n++
		}
	}
	if n >= 400 {
		t.Fatalf("%d cell elements on a 1000x52 sheet", n)
	}
	a.scrollRow = 900
	a.redraw()
	if m := countPrefix(a, "xl-gc-"); m >= 400 {
		t.Fatalf("%d cell elements after scrolling", m)
	}
}

func countPrefix(a *app, prefix string) int {
	n := 0
	for _, b := range a.page.Boxes() {
		if strings.HasPrefix(b.ID, prefix) {
			n++
		}
	}
	return n
}

// ---- EXL-21: selection modes ----

func TestDSelectionModes(t *testing.T) {
	a, m := newTestApp(t)
	a.selectCell(CellRef{Row: 3, Col: 3})
	if a.sel.Min != a.sel.Max || a.anchor != a.active {
		t.Fatalf("single: %v anchor %v", a.sel, a.anchor)
	}
	m.shift = true
	a.extendTo(CellRef{Row: 5, Col: 6})
	m.shift = false
	if a.sel != (Range{Min: CellRef{Row: 3, Col: 3}, Max: CellRef{Row: 5, Col: 6}}) || a.active != (CellRef{Row: 3, Col: 3}) {
		t.Fatalf("extend: %v active %v", a.sel, a.active)
	}
	a.extendTo(CellRef{Row: 1, Col: 1})
	if a.sel != (Range{Min: CellRef{Row: 1, Col: 1}, Max: CellRef{Row: 3, Col: 3}}) {
		t.Fatalf("extend up-left: %v", a.sel)
	}
	a.selectRange(Range{Min: CellRef{Col: 2}, Max: CellRef{Row: allCellsMaxRow, Col: 2}})
	if a.sel.Min.Row != 0 || a.sel.Max.Row != allCellsMaxRow || a.sel.Min.Col != 2 {
		t.Fatalf("column: %v", a.sel)
	}
	a.selectRange(Range{Min: CellRef{Row: 4}, Max: CellRef{Row: 4, Col: allCellsMaxCol}})
	if a.sel.Min.Col != 0 || a.sel.Max.Col != allCellsMaxCol || a.sel.Min.Row != 4 {
		t.Fatalf("row: %v", a.sel)
	}
}

// ---- EXL-20: scrollbar track click pages ----

func TestDScrollbarTrackClickPages(t *testing.T) {
	a, _ := newTestApp(t)
	vw, vh := a.gridViewportSize()
	vt := a.vTrack(vw, vh)
	gx, gy := a.gridOrigin()
	pointer(t, a, gx+float64(vt[0]+vt[2]/2), gy+float64(vt[1]+vt[3]-6))
	if a.scrollRow != a.visibleRowCount() {
		t.Fatalf("track click scrolled to row %d, want one page %d", a.scrollRow, a.visibleRowCount())
	}
	ht := a.hTrack(vw, vh)
	pointer(t, a, gx+float64(ht[0]+ht[2]-6), gy+float64(ht[1]+ht[3]/2))
	if a.scrollCol == 0 {
		t.Fatal("horizontal track click did not scroll")
	}
}

// ---- a tiny render check for the mounted tree ----

func TestDGridTreeRenders(t *testing.T) {
	a, _ := newTestApp(t)
	html, err := ui.Render(a.gridView())
	if err != nil || !strings.Contains(html, "xl-gcanvas") {
		t.Fatalf("render: %v", err)
	}
}
