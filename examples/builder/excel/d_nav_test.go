package main

import (
	"context"
	"strings"
	"testing"
)

// ---- EXL-16: headers stay put on both axes ----

func TestDStickyHeadersHorizontalAndVertical(t *testing.T) {
	a, _ := newTestApp(t)
	row0, col0 := mustBox(t, a, "xl-grh-0"), mustBox(t, a, "xl-gch-0")
	corner := mustBox(t, a, "xl-gcorner")
	a.sheet().SetFreeze(0, 0)
	a.scrollCol, a.scrollRow = 8, 60
	a.selectCell(CellRef{Row: 62, Col: 9})
	row1, col1 := mustBox(t, a, "xl-grh-"+itoa(a.scrollRow)), mustBox(t, a, "xl-gch-"+itoa(a.scrollCol))
	if row1.X != row0.X || row1.W != row0.W {
		t.Fatalf("row header moved horizontally: %v -> %v", row0.X, row1.X)
	}
	if col1.Y != col0.Y || col1.H != col0.H {
		t.Fatalf("column header moved vertically: %v -> %v", col0.Y, col1.Y)
	}
	if c := mustBox(t, a, "xl-gcorner"); c.X != corner.X || c.Y != corner.Y {
		t.Fatalf("corner moved")
	}
	// The headers of the active cell carry the highlight, the others do not.
	for id, want := range map[string]bool{"xl-gch-9": true, "xl-gch-8": false, "xl-grh-62": true, "xl-grh-61": false} {
		if got := headerHighlighted(a, id); got != want {
			t.Errorf("%s highlighted = %v, want %v", id, got, want)
		}
	}
}

// headerHighlighted reports whether the header element with id carries the
// active-span class in the rendered HTML.
func headerHighlighted(a *app, id string) bool {
	html := a.page.HTML()
	i := strings.Index(html, `id="`+id+`"`)
	if i < 0 {
		return false
	}
	start := strings.LastIndex(html[:i], "<")
	end := strings.Index(html[i:], ">")
	return strings.Contains(html[start:i+end], "xl-hdr--on")
}

// ---- EXL-18: freeze commands ----

func TestDFreezeCommands(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	a.runAction("xl:freeze:none")
	if sh.FreezeRows() != 0 || sh.FreezeCols() != 0 {
		t.Fatalf("unfreeze left %d,%d", sh.FreezeRows(), sh.FreezeCols())
	}
	a.runAction("xl:freeze:row")
	if sh.FreezeRows() != 1 || sh.FreezeCols() != 0 {
		t.Fatalf("freeze top row: %d,%d", sh.FreezeRows(), sh.FreezeCols())
	}
	a.runAction("xl:freeze:col")
	if sh.FreezeCols() != 1 || sh.FreezeRows() != 1 {
		t.Fatalf("freeze first column: %d,%d", sh.FreezeRows(), sh.FreezeCols())
	}
	a.selectCell(CellRef{Row: 3, Col: 2})
	a.runAction("xl:freeze:both")
	if sh.FreezeRows() != 3 || sh.FreezeCols() != 2 {
		t.Fatalf("freeze at active cell: %d,%d", sh.FreezeRows(), sh.FreezeCols())
	}
	a.scrollRow, a.scrollCol = 300, 10
	a.redraw()
	fx, fy := mustBox(t, a, "xl-grh-2"), mustBox(t, a, "xl-gch-1")
	if fx.Y != float64(0)+mustBox(t, a, "xl-grh-0").Y+float64(sh.RowHeight(0)+sh.RowHeight(1)) || fy.X <= 0 {
		t.Fatalf("frozen headers not pinned: %v %v", fx.Y, fy.X)
	}
	for _, id := range []string{"xl-gfz-v", "xl-gfz-h"} {
		if _, ok := box(a, id); !ok {
			t.Errorf("divider %s missing", id)
		}
	}
	a.runAction("xl:freeze:none")
	for _, id := range []string{"xl-gfz-v", "xl-gfz-h"} {
		if _, ok := box(a, id); ok {
			t.Errorf("divider %s still drawn after unfreeze", id)
		}
	}
}

// ---- EXL-24: keyboard navigation details ----

func TestDKeyboardWrapAndRegionJumps(t *testing.T) {
	a, m := newTestApp(t)

	// Tab and Enter wrap inside a multi-cell selection.
	a.selectRange(Range{Min: CellRef{Row: 20, Col: 2}, Max: CellRef{Row: 21, Col: 3}})
	key(t, a, "tab")
	key(t, a, "tab")
	if a.active != (CellRef{Row: 21, Col: 2}) {
		t.Fatalf("tab wrap: %v", a.active)
	}
	key(t, a, "enter")
	if a.active != (CellRef{Row: 20, Col: 3}) {
		t.Fatalf("enter wrap: %v", a.active)
	}
	m.shift = true
	key(t, a, "enter")
	m.shift = false
	if a.active != (CellRef{Row: 21, Col: 2}) {
		t.Fatalf("shift+enter wrap: %v", a.active)
	}
	if a.sel.Max != (CellRef{Row: 21, Col: 3}) {
		t.Fatalf("selection changed while stepping: %v", a.sel)
	}

	// Ctrl+Arrow stops at the edges of a data region (the starter table is A1:H10).
	a.selectCell(CellRef{Row: 1, Col: 0})
	m.ctrl = true
	key(t, a, "arrowdown")
	if a.active != (CellRef{Row: 9, Col: 0}) {
		t.Fatalf("ctrl+down from A2: %v", a.active)
	}
	key(t, a, "arrowup")
	if a.active != (CellRef{Row: 0, Col: 0}) {
		t.Fatalf("ctrl+up: %v", a.active)
	}
	key(t, a, "arrowright")
	if a.active != (CellRef{Row: 0, Col: 7}) {
		t.Fatalf("ctrl+right: %v", a.active)
	}
	key(t, a, "arrowleft")
	if a.active != (CellRef{Row: 0, Col: 0}) {
		t.Fatalf("ctrl+left: %v", a.active)
	}
	m.ctrl = false

	// Moving past the visible window scrolls it.
	a.selectCell(CellRef{Row: 0, Col: 0})
	for range 60 {
		key(t, a, "arrowdown")
	}
	if a.active.Row != 60 || a.scrollRow == 0 {
		t.Fatalf("active %v scrollRow %d after 60 downs", a.active, a.scrollRow)
	}
	if _, ok := box(a, "xl-grh-60"); !ok {
		t.Fatal("active row is not on screen")
	}
}

// ---- EXL-25 and EXL-26: resize clamps, tooltip and double-click auto-fit ----

func TestDResizeClampTooltipAndDoubleClick(t *testing.T) {
	a, _ := newTestApp(t)
	ctx := context.Background()
	gx, gy := a.gridOrigin()
	sh := a.sheet()
	edge := gx + float64(rowHeaderWidth+sh.ColWidth(0)+sh.ColWidth(1))
	y := gy + 10
	if ok, err := a.page.Testing().BeginDrag(ctx, edge, y); err != nil || !ok {
		t.Fatalf("BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, edge+12, y); err != nil {
		t.Fatal(err)
	}
	want := "Width: " + itoa(sh.ColWidth(1)) + " px"
	if !strings.Contains(a.page.HTML(), want) {
		t.Fatalf("no live tooltip %q during the drag", want)
	}
	if err := a.page.Testing().MoveDrag(ctx, edge-400, y); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if got := sh.ColWidth(1); got != minColWidth {
		t.Fatalf("column clamped to %d, want %d", got, minColWidth)
	}
	if strings.Contains(a.page.HTML(), "xl-gtip") {
		t.Fatal("tooltip stays after the drag")
	}

	// Rows clamp at the minimum height too.
	redge := gy + float64(colHeaderHeight+sh.RowHeight(0))
	rx := gx + 10
	if ok, err := a.page.Testing().BeginDrag(ctx, rx, redge); err != nil || !ok {
		t.Fatalf("row BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, rx, redge-200); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if got := sh.RowHeight(0); got != minRowHeight {
		t.Fatalf("row clamped to %d, want %d", got, minRowHeight)
	}

	// Double-click on the border auto-fits the column to its content.
	sh.SetRaw(2, 4, "a rather long label that overflows the default width")
	sh.SetColWidth(4, defaultColWidth)
	a.redraw()
	right := gx + float64(rowHeaderWidth)
	cols := a.layoutCols(func() int { w, _ := a.canvasSize(); return w }())
	s, _ := slotFor(cols, 4)
	right += float64(s.pos + s.size)
	pointer(t, a, right, gy+10)
	pointer(t, a, right, gy+10)
	if got := sh.ColWidth(4); got <= defaultColWidth {
		t.Fatalf("double-click auto-fit left width %d", got)
	}
}

// ---- EXL-40 to EXL-43: the function library covers every listed name ----

func TestDFunctionLibraryCoverage(t *testing.T) {
	groups := map[string][]string{
		"math": {"SUM", "AVERAGE", "MIN", "MAX", "COUNT", "COUNTA", "COUNTBLANK", "ROUND", "ROUNDUP", "ROUNDDOWN",
			"ABS", "MOD", "POWER", "SQRT", "INT", "SUMIF", "COUNTIF", "SUMIFS"},
		"logical": {"IF", "IFS", "AND", "OR", "NOT", "XOR", "IFERROR", "IFNA", "ISBLANK", "ISNUMBER", "ISTEXT", "ISERROR"},
		"lookup":  {"VLOOKUP", "HLOOKUP", "XLOOKUP", "INDEX", "MATCH", "ROW", "COLUMN", "ROWS", "COLUMNS"},
		"text": {"CONCAT", "TEXTJOIN", "LEFT", "RIGHT", "MID", "LEN", "TRIM", "UPPER", "LOWER", "PROPER", "SUBSTITUTE",
			"FIND", "TEXT", "TODAY", "NOW", "DATE", "YEAR", "MONTH", "DAY", "DATEDIF"},
	}
	have := map[string]bool{}
	for _, n := range FunctionNames() {
		have[n] = true
	}
	total := 0
	for group, names := range groups {
		for _, n := range names {
			total++
			if !have[n] {
				t.Errorf("%s function %s is not registered", group, n)
			}
			if FunctionSignatures()[n] == "" {
				t.Errorf("%s has no autocomplete signature", n)
			}
		}
	}
	if total < 35 || len(FunctionNames()) < 35 {
		t.Fatalf("%d listed, %d registered", total, len(FunctionNames()))
	}
}

// ---- EXL-44: seeded formulas evaluate without errors on every sheet ----

func TestDStarterFormulasHaveNoErrors(t *testing.T) {
	w := NewWorkbook()
	formulas := 0
	for _, s := range w.Sheets {
		u := s.UsedRange()
		for r := u.Min.Row; r <= u.Max.Row; r++ {
			for c := u.Min.Col; c <= u.Max.Col; c++ {
				raw := s.Raw(r, c)
				if strings.HasPrefix(raw, "=") {
					formulas++
					if v := s.Value(r, c); v.Kind == KindError {
						t.Errorf("%s!%s %s = %s", s.Name, CellRef{Row: r, Col: c}.A1(), raw, v.Err)
					}
				}
			}
		}
	}
	if formulas == 0 {
		t.Fatal("no seeded formulas found")
	}
}
