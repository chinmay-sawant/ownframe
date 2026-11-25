package main

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

type testMods struct{ shift, ctrl bool }

func newTestApp(t *testing.T) (*app, *testMods) {
	t.Helper()
	ownframe.UseMemory(true)
	ownframe.Clear()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	m := &testMods{}
	a.modsFn = func() (bool, bool) { return m.shift, m.ctrl }
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return a, m
}

func box(a *app, id string) (ownframe.Box, bool) {
	for _, b := range a.page.Boxes() {
		if b.ID == id {
			return b, true
		}
	}
	return ownframe.Box{}, false
}

func mustBox(t *testing.T, a *app, id string) ownframe.Box {
	t.Helper()
	b, ok := box(a, id)
	if !ok {
		t.Fatalf("no box %q on the page", id)
	}
	return b
}

// cellPoint is the window coordinate at the middle of a visible cell.
func cellPoint(t *testing.T, a *app, c CellRef) (float64, float64) {
	t.Helper()
	cw, ch := a.canvasSize()
	col, okc := slotFor(a.layoutCols(cw), c.Col)
	row, okr := slotFor(a.layoutRows(ch), c.Row)
	if !okc || !okr {
		t.Fatalf("cell %s is not visible", c.A1())
	}
	gx, gy := a.gridOrigin()
	return gx + float64(a.rowHdrW()+col.pos+col.size/2), gy + float64(a.colHdrH()+row.pos+row.size/2)
}

func pointer(t *testing.T, a *app, x, y float64) {
	t.Helper()
	ctx := context.Background()
	claimed, err := a.page.Testing().BeginDrag(ctx, x, y)
	if err != nil {
		t.Fatal(err)
	}
	if claimed {
		if err := a.page.Testing().EndDrag(ctx); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := a.page.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func clickCell(t *testing.T, a *app, c CellRef) {
	t.Helper()
	x, y := cellPoint(t, a, c)
	pointer(t, a, x, y)
}

func key(t *testing.T, a *app, k string) {
	t.Helper()
	if err := a.page.Edit().KeyDown(context.Background(), k); err != nil {
		t.Fatal(err)
	}
}

func typeText(t *testing.T, a *app, s string) {
	t.Helper()
	if err := a.page.Edit().Type(context.Background(), s); err != nil {
		t.Fatal(err)
	}
}

func framePNG(t *testing.T, a *app) image.Image {
	t.Helper()
	data := a.page.Probe().PNG()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode frame: %v", err)
	}
	return img
}

func TestNewAppMounts(t *testing.T) {
	a, _ := newTestApp(t)
	if a.lastErr != nil {
		t.Fatal(a.lastErr)
	}
	if w, h := a.page.Size(); w != windowWidth || h != windowHeight {
		t.Fatalf("frame %dx%d", w, h)
	}
	html := a.page.HTML()
	for _, want := range []string{"xl-titlebar", "xl-grid", "xl-fbar", "xl-tabs", "xl-status"} {
		if !strings.Contains(html, want) {
			t.Fatalf("page misses %s", want)
		}
	}
}

func TestShellFitsViewport(t *testing.T) {
	a, _ := newTestApp(t)
	_, h := a.page.Size()
	ids := []string{"xl-titlebar", "xl-ribbon-slot", "xl-fbar", "xl-grid", "xl-tabs", "xl-status"}
	prevBottom := 0.0
	for _, id := range ids {
		b := mustBox(t, a, id)
		if b.Y+0.5 < prevBottom {
			t.Fatalf("%s starts at %v, above the previous band bottom %v", id, b.Y, prevBottom)
		}
		prevBottom = b.Y + b.H
	}
	if prevBottom > float64(h)+0.5 {
		t.Fatalf("bands end at %v, past the %d px viewport", prevBottom, h)
	}
	grid := mustBox(t, a, "xl-grid")
	_, gh := a.gridViewportSize()
	if int(grid.H) != gh {
		t.Fatalf("grid box height %v, computed viewport %d", grid.H, gh)
	}
}

func TestGridNodeBudget(t *testing.T) {
	a, _ := newTestApp(t)
	n := 0
	for _, b := range a.page.Boxes() {
		if strings.HasPrefix(b.ID, "xl-gc-") {
			n++
		}
	}
	if n == 0 || n >= 400 {
		t.Fatalf("rendered %d cell elements, want 1..399", n)
	}
	if total := len(a.page.Boxes()); total > 900 {
		t.Fatalf("page has %d boxes, want a bounded window", total)
	}
}

func TestPointerSelectionFlows(t *testing.T) {
	a, m := newTestApp(t)
	clickCell(t, a, CellRef{Row: 2, Col: 1})
	if a.active != (CellRef{Row: 2, Col: 1}) || a.sel.Min != a.sel.Max {
		t.Fatalf("click: active %v sel %v", a.active, a.sel)
	}
	m.shift = true
	clickCell(t, a, CellRef{Row: 4, Col: 3})
	m.shift = false
	if a.sel.Min != (CellRef{Row: 2, Col: 1}) || a.sel.Max != (CellRef{Row: 4, Col: 3}) || a.active != (CellRef{Row: 2, Col: 1}) {
		t.Fatalf("shift+click: active %v sel %v", a.active, a.sel)
	}

	// Drag sweep.
	ctx := context.Background()
	x0, y0 := cellPoint(t, a, CellRef{Row: 6, Col: 0})
	x1, y1 := cellPoint(t, a, CellRef{Row: 8, Col: 2})
	if ok, err := a.page.Testing().BeginDrag(ctx, x0, y0); err != nil || !ok {
		t.Fatalf("BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x1, y1); err != nil {
		t.Fatal(err)
	}
	if got := a.nameBoxText(); got != "3R x 3C" {
		t.Fatalf("name box during drag = %q", got)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.sel.Min != (CellRef{Row: 6}) || a.sel.Max != (CellRef{Row: 8, Col: 2}) {
		t.Fatalf("drag sweep: sel %v", a.sel)
	}
	if got := a.nameBoxText(); got != "A7:C9" {
		t.Fatalf("name box after drag = %q", got)
	}

	// Column header, row header and corner.
	hx, hy := a.gridOrigin()
	pointer(t, a, hx+float64(rowHeaderWidth+84+40), hy+10) // column B
	if a.sel.Min.Col != 1 || a.sel.Max.Col != 1 || a.sel.Max.Row != MaxRows-1 {
		t.Fatalf("column header: sel %v", a.sel)
	}
	pointer(t, a, hx+10, hy+float64(colHeaderHeight+22*2+10)) // row 3
	if a.sel.Min.Row != 2 || a.sel.Max.Row != 2 || a.sel.Max.Col != MaxCols-1 {
		t.Fatalf("row header: sel %v", a.sel)
	}
	pointer(t, a, hx+10, hy+10)
	if a.sel.Min != (CellRef{}) || a.sel.Max != (CellRef{Row: MaxRows - 1, Col: MaxCols - 1}) {
		t.Fatalf("corner: sel %v", a.sel)
	}
}

func TestKeyboardNavigationMatrix(t *testing.T) {
	a, m := newTestApp(t)
	a.selectCell(CellRef{Row: 5, Col: 5})
	steps := []struct {
		key  string
		want CellRef
	}{
		{"arrowdown", CellRef{6, 5}},
		{"arrowup", CellRef{5, 5}},
		{"arrowright", CellRef{5, 6}},
		{"arrowleft", CellRef{5, 5}},
		{"tab", CellRef{5, 6}},
		{"enter", CellRef{6, 6}},
		{"home", CellRef{6, 0}},
		{"pagedown", CellRef{6 + a.visibleRowCount(), 0}},
		{"pageup", CellRef{6, 0}},
	}
	for _, s := range steps {
		key(t, a, s.key)
		if a.active != s.want {
			t.Fatalf("%s: active %v, want %v", s.key, a.active, s.want)
		}
	}
	m.shift = true
	key(t, a, "tab")
	m.shift = false
	if a.active != (CellRef{6, 0}) {
		t.Fatalf("shift+tab at column A: %v", a.active)
	}

	m.shift = true
	key(t, a, "arrowdown")
	key(t, a, "arrowright")
	m.shift = false
	if a.sel.Min != (CellRef{6, 0}) || a.sel.Max != (CellRef{7, 1}) {
		t.Fatalf("shift+arrows: sel %v", a.sel)
	}

	m.ctrl = true
	key(t, a, "home")
	if a.active != (CellRef{}) {
		t.Fatalf("ctrl+home: %v", a.active)
	}
	key(t, a, "end")
	used := a.sheet().UsedRange().Max
	if a.active != used {
		t.Fatalf("ctrl+end: %v, want %v", a.active, used)
	}
	a.selectCell(CellRef{})
	key(t, a, "arrowdown") // jumps down the first column's data block
	m.ctrl = false
	if a.active.Row < 1 {
		t.Fatalf("ctrl+down did not move: %v", a.active)
	}
}

func TestInlineEditLifecycle(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	target := CellRef{Row: 30, Col: 8}
	a.selectCell(target)

	// Typing replaces the cell and enters edit mode; Enter commits and moves down.
	typeText(t, a, "42")
	if !a.editing || a.editBuf != "42" {
		t.Fatalf("after typing: editing=%v buf=%q", a.editing, a.editBuf)
	}
	key(t, a, "enter")
	if a.editing || sh.Raw(30, 8) != "42" || a.active != (CellRef{31, 8}) {
		t.Fatalf("enter: editing=%v raw=%q active=%v", a.editing, sh.Raw(30, 8), a.active)
	}

	// Tab commits and moves right.
	typeText(t, a, "x")
	key(t, a, "tab")
	if sh.Raw(31, 8) != "x" || a.active != (CellRef{31, 9}) {
		t.Fatalf("tab: raw=%q active=%v", sh.Raw(31, 8), a.active)
	}

	// F2 keeps the content; Escape restores it.
	a.selectCell(target)
	key(t, a, "f2")
	if !a.editing || a.editBuf != "42" {
		t.Fatalf("f2: editing=%v buf=%q", a.editing, a.editBuf)
	}
	a.page.Form().SetFormValue("xl-editor", "99")
	if err := a.handleChange(context.Background(), ownframe.Box{ID: "xl-editor"}); err != nil {
		t.Fatal(err)
	}
	key(t, a, "escape")
	if a.editing || sh.Raw(30, 8) != "42" {
		t.Fatalf("escape: editing=%v raw=%q", a.editing, sh.Raw(30, 8))
	}

	// Delete clears without editing.
	key(t, a, "delete")
	if a.editing || sh.Raw(30, 8) != "" {
		t.Fatalf("delete: editing=%v raw=%q", a.editing, sh.Raw(30, 8))
	}

	// Backspace clears too.
	sh.SetRaw(30, 8, "7")
	if err := a.page.Edit().Backspace(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sh.Raw(30, 8) != "" {
		t.Fatalf("backspace: raw=%q", sh.Raw(30, 8))
	}
}

func TestDoubleClickEdits(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetRaw(1, 1, "hello")
	clickCell(t, a, CellRef{1, 1})
	clickCell(t, a, CellRef{1, 1})
	if !a.editing || a.editBuf != "hello" {
		t.Fatalf("double click: editing=%v buf=%q", a.editing, a.editBuf)
	}
}

func TestFormulaBarSync(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetRaw(2, 2, "=1+2")
	a.sheet().SetRaw(3, 2, "plain")
	a.selectCell(CellRef{2, 2})
	if got := a.page.Form().FormValue("xl-fx"); got != "=1+2" {
		t.Fatalf("formula input = %q", got)
	}
	if got := a.page.Form().FormValue("xl-name"); got != "C3" {
		t.Fatalf("name box = %q", got)
	}
	a.selectCell(CellRef{3, 2})
	if got := a.page.Form().FormValue("xl-fx"); got != "plain" {
		t.Fatalf("formula input = %q", got)
	}
}

func TestNameBoxJump(t *testing.T) {
	a, _ := newTestApp(t)
	ctx := context.Background()
	if err := a.page.Form().Focus(ctx, "xl-name"); err != nil {
		t.Fatal(err)
	}
	a.page.Form().SetFormValue("xl-name", "D25")
	key(t, a, "enter")
	if a.active != (CellRef{Row: 24, Col: 3}) {
		t.Fatalf("active = %v", a.active)
	}
	if a.scrollRow == 0 {
		t.Fatalf("grid did not scroll to row 25 (scrollRow %d)", a.scrollRow)
	}
	if err := a.page.Form().Focus(ctx, "xl-name"); err != nil {
		t.Fatal(err)
	}
	a.page.Form().SetFormValue("xl-name", "A1:C3")
	key(t, a, "enter")
	if a.sel.Min != (CellRef{}) || a.sel.Max != (CellRef{2, 2}) {
		t.Fatalf("range jump: sel %v", a.sel)
	}
}

func TestGridWheelScroll(t *testing.T) {
	a, m := newTestApp(t)
	gx, gy := a.gridOrigin()
	ctx := context.Background()
	if handled, err := a.handleWheel(ctx, gx+200, gy+200, 0, -3); err != nil || !handled {
		t.Fatalf("wheel handled=%v err=%v", handled, err)
	}
	if a.scrollRow != 9 {
		t.Fatalf("scrollRow = %d, want 9", a.scrollRow)
	}
	m.shift = true
	if _, err := a.handleWheel(ctx, gx+200, gy+200, 0, -2); err != nil {
		t.Fatal(err)
	}
	m.shift = false
	if a.scrollCol != 2 || a.scrollRow != 9 {
		t.Fatalf("shift+wheel: scrollRow %d scrollCol %d", a.scrollRow, a.scrollCol)
	}
	// Clamp at the top.
	if _, err := a.handleWheel(ctx, gx+200, gy+200, 0, 1000); err != nil {
		t.Fatal(err)
	}
	if a.scrollRow != a.sheet().FreezeRows() {
		t.Fatalf("scrollRow = %d after scrolling up, want the freeze boundary", a.scrollRow)
	}
	if handled, _ := a.handleWheel(ctx, gx+200, 5, 0, -3); handled {
		t.Fatalf("wheel over the title bar should not be consumed")
	}
	// Scrolled window shows later rows and bounded node count.
	a.scrollRow = 500
	a.redraw()
	if _, ok := box(a, "xl-grh-500"); !ok {
		t.Fatalf("row header 501 missing after scroll")
	}
	if _, ok := box(a, "xl-grh-5"); ok {
		t.Fatalf("row header 6 still rendered after scroll")
	}
}

func TestFreezePanesStayPinned(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetFreeze(1, 1)
	a.redraw()
	h0 := mustBox(t, a, "xl-grh-0")
	c0 := mustBox(t, a, "xl-gch-0")
	a.scrollRow, a.scrollCol = 200, 10
	a.redraw()
	h1 := mustBox(t, a, "xl-grh-0")
	c1 := mustBox(t, a, "xl-gch-0")
	if h0.Y != h1.Y || c0.X != c1.X {
		t.Fatalf("frozen headers moved: row %v -> %v, col %v -> %v", h0.Y, h1.Y, c0.X, c1.X)
	}
	next := mustBox(t, a, "xl-grh-200")
	if next.Y <= h1.Y {
		t.Fatalf("scrolled row 201 at y=%v is not below the frozen row at %v", next.Y, h1.Y)
	}
	if _, ok := box(a, "xl-gfz-v"); !ok {
		t.Fatalf("freeze divider missing")
	}
}

func TestStickyHeadersAndActiveHighlight(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{Row: 3, Col: 2})
	html := a.page.HTML()
	if !strings.Contains(html, "xl-hdr--on") {
		t.Fatalf("no active header highlight")
	}
	before := mustBox(t, a, "xl-gch-2")
	a.scrollRow = 100
	a.redraw()
	after := mustBox(t, a, "xl-gch-2")
	if before.Y != after.Y {
		t.Fatalf("column header moved vertically: %v -> %v", before.Y, after.Y)
	}
}

func TestSelectionOverlayGeometry(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectRange(Range{Min: CellRef{1, 1}, Max: CellRef{3, 2}})
	sel := mustBox(t, a, "xl-gov-sel")
	tl := mustBox(t, a, "xl-gc-1-1")
	_ = tl
	gx, gy := a.gridOrigin()
	sh := a.sheet()
	wantX := gx + float64(rowHeaderWidth+sh.ColWidth(0))
	wantY := gy + float64(colHeaderHeight+sh.RowHeight(0))
	if sel.X != wantX || sel.Y != wantY {
		t.Fatalf("selection at (%v,%v), want (%v,%v)", sel.X, sel.Y, wantX, wantY)
	}
	if sel.W != float64(sh.ColWidth(1)+sh.ColWidth(2)) || sel.H != float64(sh.RowHeight(1)+sh.RowHeight(2)+sh.RowHeight(3)) {
		t.Fatalf("selection size %vx%v", sel.W, sel.H)
	}
	act := mustBox(t, a, "xl-gov-active")
	if act.X != wantX-1 || act.Y != wantY-1 {
		t.Fatalf("active outline at (%v,%v)", act.X, act.Y)
	}
}

func TestHeaderResizeDrag(t *testing.T) {
	a, _ := newTestApp(t)
	ctx := context.Background()
	gx, gy := a.gridOrigin()
	w0, w1 := a.sheet().ColWidth(0), a.sheet().ColWidth(1)
	edge := gx + float64(rowHeaderWidth+w0+w1) // right edge of column B
	y := gy + 10
	if ok, err := a.page.Testing().BeginDrag(ctx, edge, y); err != nil || !ok {
		t.Fatalf("BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, edge+40, y); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if got := a.sheet().ColWidth(1); got != w1+40 {
		t.Fatalf("column B width = %d, want %d", got, w1+40)
	}
	b := mustBox(t, a, "xl-gch-1")
	if b.W != float64(w1+40) {
		t.Fatalf("header box width %v", b.W)
	}
}

func TestHeaderAutoFit(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetRaw(1, 5, "a rather long label that overflows the default width")
	a.redraw()
	a.autoFitCol(5)
	if got := a.sheet().ColWidth(5); got <= defaultColWidth {
		t.Fatalf("column F width %d, want wider than %d", got, defaultColWidth)
	}
}

func TestScrollbarDragUpdatesViewport(t *testing.T) {
	a, _ := newTestApp(t)
	ctx := context.Background()
	vw, vh := a.gridViewportSize()
	vt := a.vTrack(vw, vh)
	gx, gy := a.gridOrigin()
	x := gx + float64(vt[0]+vt[2]/2)
	y := gy + float64(vt[1]+4)
	if ok, err := a.page.Testing().BeginDrag(ctx, x, y); err != nil || !ok {
		t.Fatalf("BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x, y+120); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.scrollRow < 20 {
		t.Fatalf("scrollRow = %d after dragging the thumb", a.scrollRow)
	}
	// Track click pages down.
	before := a.scrollRow
	pointer(t, a, x, gy+float64(vt[1]+vt[3]-6))
	if a.scrollRow <= before {
		t.Fatalf("track click did not page down (%d -> %d)", before, a.scrollRow)
	}
}

func TestFormulaClickInsertAndF4(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{Row: 20, Col: 6})
	a.beginEdit("=SUM(")
	ctx := context.Background()
	x0, y0 := cellPoint(t, a, CellRef{1, 1})
	x1, y1 := cellPoint(t, a, CellRef{4, 1})
	if ok, err := a.page.Testing().BeginDrag(ctx, x0, y0); err != nil || !ok {
		t.Fatalf("BeginDrag claimed=%v err=%v", ok, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x1, y1); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if !a.editing || a.editBuf != "=SUM(B2:B5" {
		t.Fatalf("editing=%v buf=%q", a.editing, a.editBuf)
	}
	key(t, a, "f4")
	if a.editBuf != "=SUM($B$2:$B$5" {
		t.Fatalf("after F4: %q", a.editBuf)
	}
	key(t, a, "f4")
	if a.editBuf != "=SUM(B$2:B$5" {
		t.Fatalf("after F4 x2: %q", a.editBuf)
	}
	if got := cycleOneRef("$A1"); got != "A1" {
		t.Fatalf("cycleOneRef($A1) = %q", got)
	}
}

func TestFormulaRefHighlights(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{Row: 20, Col: 6})
	a.beginEdit("=A1+B2:B4")
	if _, ok := box(a, "xl-gov-ref-0"); !ok {
		t.Fatalf("highlight for A1 missing")
	}
	r1 := mustBox(t, a, "xl-gov-ref-1")
	if r1.H != float64(a.sheet().RowHeight(1)+a.sheet().RowHeight(2)+a.sheet().RowHeight(3)) {
		t.Fatalf("B2:B4 highlight height %v", r1.H)
	}
	if !strings.Contains(a.page.HTML(), "xl-ref-2") {
		t.Fatalf("second reference does not use the second palette color")
	}
}

func TestClipboardCopyPasteUndo(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	sh.SetRaw(40, 0, "10")
	sh.SetRaw(40, 1, "20")
	a.selectRange(Range{Min: CellRef{40, 0}, Max: CellRef{40, 1}})
	if _, _, err := a.handleCopy(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.copy(false)
	a.selectCell(CellRef{42, 0})
	a.paste()
	if sh.Raw(42, 0) != "10" || sh.Raw(42, 1) != "20" {
		t.Fatalf("paste: %q %q", sh.Raw(42, 0), sh.Raw(42, 1))
	}
	if a.sel.Max != (CellRef{42, 1}) {
		t.Fatalf("pasted selection %v", a.sel)
	}
	a.undo()
	if sh.Raw(42, 0) != "" {
		t.Fatalf("undo left %q", sh.Raw(42, 0))
	}
	a.redo()
	if sh.Raw(42, 0) != "10" {
		t.Fatalf("redo gave %q", sh.Raw(42, 0))
	}

	// Cut moves the source after paste.
	a.selectCell(CellRef{40, 0})
	a.copy(true)
	a.selectCell(CellRef{45, 0})
	a.paste()
	if sh.Raw(45, 0) != "10" || sh.Raw(40, 0) != "" {
		t.Fatalf("cut/paste: dest %q source %q", sh.Raw(45, 0), sh.Raw(40, 0))
	}
}

func TestThemeToggleAndTitleBarActions(t *testing.T) {
	a, _ := newTestApp(t)
	for _, act := range []string{"win-min", "win-max", "win-close"} {
		if !strings.Contains(a.page.HTML(), `data-action="`+act+`"`) {
			t.Fatalf("title bar misses %s", act)
		}
	}
	a.selectCell(CellRef{4, 4})
	light := framePNG(t, a)
	lt := light.At(300, 18)
	lg := light.At(700, 400)
	if err := a.handleClick(context.Background(), ownframe.Box{Action: "theme:toggle"}); err != nil {
		t.Fatal(err)
	}
	if !a.dark {
		t.Fatal("theme did not toggle")
	}
	if a.active != (CellRef{4, 4}) {
		t.Fatalf("selection lost on theme toggle: %v", a.active)
	}
	dark := framePNG(t, a)
	if lt == dark.At(300, 18) {
		t.Fatalf("title bar pixel unchanged between themes: %v", lt)
	}
	if lg == dark.At(700, 400) {
		t.Fatalf("grid pixel unchanged between themes: %v", lg)
	}
}

func TestSheetBarSwitchAndStatusStats(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{0, 0})
	if err := a.handleClick(context.Background(), ownframe.Box{Action: "sheet:1"}); err != nil {
		t.Fatal(err)
	}
	if a.wb.Active != 1 {
		t.Fatalf("active sheet %d", a.wb.Active)
	}
	if err := a.handleClick(context.Background(), ownframe.Box{Action: "sheet:add"}); err != nil {
		t.Fatal(err)
	}
	if a.wb.Active != len(a.wb.Sheets)-1 || len(a.wb.Sheets) < 4 {
		t.Fatalf("add sheet: active %d of %d", a.wb.Active, len(a.wb.Sheets))
	}
	a.sheet().SetRaw(0, 0, "2")
	a.sheet().SetRaw(1, 0, "4")
	a.selectRange(Range{Max: CellRef{Row: 1}})
	html := a.page.HTML()
	for _, want := range []string{"Average: 3", "Count: 2", "Sum: 6"} {
		if !strings.Contains(html, want) {
			t.Fatalf("status bar misses %q", want)
		}
	}
}
