package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// ---- EXL-14: the visible window follows the sheet axes ----

func TestDGridWindowVariableAxes(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	sh.SetColWidth(1, 120)
	sh.SetColWidth(2, 30)
	sh.SetRowHidden(2, true)
	sh.SetRowHeight(3, 40)

	cw, ch := a.canvasSize()
	cols := a.layoutCols(cw)
	pos := 0
	for _, c := range cols {
		if c.pos != pos {
			t.Fatalf("column %d at %d, want prefix sum %d", c.idx, c.pos, pos)
		}
		pos += c.size
	}
	if cols[1].size != 120 || cols[2].size != 30 || pos > cw {
		t.Fatalf("column slots %+v overflow canvas %d", cols[:3], cw)
	}

	rows := a.layoutRows(ch)
	for _, r := range rows {
		if r.idx == 2 {
			t.Fatal("hidden row 3 is in the window")
		}
	}
	if rows[2].idx != 3 || rows[2].size != 40 || rows[2].pos != 2*defaultRowHeight {
		t.Fatalf("row slots %+v", rows[:4])
	}

	// Deep in a tall sheet the window still holds only the rows that fit.
	a.scrollRow = MaxRows - 5
	a.clampScroll()
	deep := a.layoutRows(ch)
	// Row 1 is frozen on the starter sheet, then the window starts at scrollRow.
	if deep[1].idx != a.scrollRow || len(deep) > ch/defaultRowHeight+2 || deep[len(deep)-1].idx != MaxRows-1 {
		t.Fatalf("deep window %d..%d (%d rows)", deep[0].idx, deep[len(deep)-1].idx, len(deep))
	}
}

// ---- EXL-17: merged cells ----

func TestDMergedCellLayoutAndHitTest(t *testing.T) {
	a, _ := newTestApp(t)
	sh := a.sheet()
	sh.SetRaw(1, 1, "merged")
	sh.Merge(Range{Min: CellRef{Row: 1, Col: 1}, Max: CellRef{Row: 2, Col: 2}})
	a.redraw()

	anchor := mustBox(t, a, "xl-gc-1-1")
	wantW := sh.ColWidth(1) + sh.ColWidth(2) - 1
	wantH := sh.RowHeight(1) + sh.RowHeight(2) - 1
	if int(anchor.W) != wantW || int(anchor.H) != wantH {
		t.Fatalf("merged box %vx%v, want %dx%d", anchor.W, anchor.H, wantW, wantH)
	}
	for _, id := range []string{"xl-gc-1-2", "xl-gc-2-1", "xl-gc-2-2"} {
		if _, ok := box(a, id); ok {
			t.Fatalf("covered cell %s still has its own element", id)
		}
	}

	// A click on a covered cell selects the anchor, by pointer and by id.
	clickCell(t, a, CellRef{Row: 2, Col: 2})
	if a.active != (CellRef{Row: 1, Col: 1}) {
		t.Fatalf("click in merge selected %v", a.active)
	}
	a.selectCell(CellRef{Row: 5, Col: 5})
	a.pressSeen = time.Time{}
	if err := a.handleClick(context.Background(), ownframe.Box{ID: "xl-gc-1-1"}); err != nil {
		t.Fatal(err)
	}
	if a.active != (CellRef{Row: 1, Col: 1}) {
		t.Fatalf("id click selected %v", a.active)
	}

	// The active outline spans the whole merge.
	act := mustBox(t, a, "xl-gov-active")
	if act.W < anchor.W || act.H < anchor.H {
		t.Fatalf("active outline %vx%v is smaller than the merge %vx%v", act.W, act.H, anchor.W, anchor.H)
	}

	// Scrolled so the anchor row is gone, the visible part still draws once.
	a.scrollRow = 2
	a.redraw()
	if _, ok := box(a, "xl-gc-1-1"); !ok {
		t.Fatal("merge vanished when its anchor row scrolled off")
	}
}

// ---- zoom ----

func TestDZoomScalesGrid(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetRaw(0, 0, "x")
	a.redraw()
	base := mustBox(t, a, "xl-gc-0-0")

	a.x.zoom = 150
	a.redraw()
	big := mustBox(t, a, "xl-gc-0-0")
	if int(big.W) != a.zs(a.sheet().ColWidth(0))-1 || int(big.H) != a.zs(a.sheet().RowHeight(0))-1 || big.W <= base.W*1.4 {
		t.Fatalf("150%% cell %vx%v, base %vx%v", big.W, big.H, base.W, base.H)
	}
	if a.zs(defaultColWidth) != 126 || a.zs(defaultRowHeight) != 33 {
		t.Fatalf("zs: %d x %d", a.zs(defaultColWidth), a.zs(defaultRowHeight))
	}
	// The whole UI scales with the zoom, so the headers grow with it.
	if h := mustBox(t, a, "xl-gch-0"); int(h.H) != a.zs(colHeaderHeight) {
		t.Fatalf("column header height %v", h.H)
	}
	// Fewer cells fit, and a click on the scaled grid still lands on the right cell.
	clickCell(t, a, CellRef{Row: 4, Col: 3})
	if a.active != (CellRef{Row: 4, Col: 3}) {
		t.Fatalf("click at 150%% selected %v", a.active)
	}
	// Resizing a column by dragging changes the sheet width in unzoomed pixels.
	before := a.sheet().ColWidth(1)
	a.drag, a.dragIndex, a.dragOrig = dragResizeCol, 1, a.unz(a.zs(before))
	a.dragAccX = 30
	if err := a.handleDragMove(context.Background(), 0, 0); err != nil {
		t.Fatal(err)
	}
	if got := a.sheet().ColWidth(1); got != before+20 {
		t.Fatalf("30 px drag at 150%% gave width %d, want %d", got, before+20)
	}
	a.drag = dragNone

	a.x.zoom = 50
	a.redraw()
	if small := mustBox(t, a, "xl-gc-0-0"); small.W >= base.W {
		t.Fatalf("50%% cell %v wide, base %v", small.W, base.W)
	}
}

// ---- filter chevrons ----

func TestDFilterChevrons(t *testing.T) {
	a, _ := newTestApp(t)
	if _, ok := box(a, "xl-gfc-1"); ok {
		t.Fatal("chevron shown with no filter")
	}
	a.x.filterOn = true
	a.redraw()
	chev := mustBox(t, a, "xl-gfc-1")
	hdr := mustBox(t, a, "xl-gch-1")
	if chev.X < hdr.X || chev.X+chev.W > hdr.X+hdr.W {
		t.Fatalf("chevron %v..%v is outside its header %v..%v", chev.X, chev.X+chev.W, hdr.X, hdr.X+hdr.W)
	}
	// A press on the chevron opens that column's popover instead of selecting the column.
	a.selectCell(CellRef{Row: 5, Col: 5})
	pointer(t, a, chev.X+chev.W/2, chev.Y+chev.H/2)
	if a.overlay != "filter" || a.x.filterCol != 1 {
		t.Fatalf("overlay %q column %d", a.overlay, a.x.filterCol)
	}
	if a.active != (CellRef{Row: 5, Col: 5}) {
		t.Fatalf("chevron press moved the selection to %v", a.active)
	}
	a.setOverlay("")

	// A column with its own filter keeps a chevron once the sheet filter is off.
	a.x.filterOn = false
	a.sheet().SetFilter(2, map[string]bool{"x": true})
	a.redraw()
	if _, ok := box(a, "xl-gfc-2"); !ok {
		t.Fatal("filtered column lost its chevron")
	}
	if _, ok := box(a, "xl-gfc-3"); ok {
		t.Fatal("unfiltered column shows a chevron")
	}
}

// ---- right click ----

func TestDRightClickSelectsAndOpensMenu(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{})
	a.sheet().SetRaw(3, 3, "v")
	a.redraw()

	rows := a.rightClick("xl-gc-3-3")
	if len(rows) == 0 || a.active != (CellRef{Row: 3, Col: 3}) {
		t.Fatalf("cell: %d rows, active %v", len(rows), a.active)
	}
	// A cell already inside the selection keeps the whole selection.
	a.selectRange(Range{Min: CellRef{Row: 3, Col: 3}, Max: CellRef{Row: 6, Col: 6}})
	a.rightClick("xl-gc-5-5")
	if a.sel.Max != (CellRef{Row: 6, Col: 6}) {
		t.Fatalf("right click inside the selection collapsed it to %v", a.sel)
	}
	if got := a.rightClick("xl-gch-2"); len(got) == 0 || a.x.ctxKind != "col" || a.sel.Min.Col != 2 || a.sel.Max.Col != 2 {
		t.Fatalf("column header: kind %q sel %v", a.x.ctxKind, a.sel)
	}
	if got := a.rightClick("xl-grh-4"); len(got) == 0 || a.x.ctxKind != "row" || a.sel.Min.Row != 4 || a.sel.Max.Row != 4 {
		t.Fatalf("row header: kind %q sel %v", a.x.ctxKind, a.sel)
	}
	if got := a.rightClick("xl-titlebar"); got != nil {
		t.Fatalf("title bar returned %d menu rows", len(got))
	}

	// A long press opens the in-page menu at the cell.
	a.selectCell(CellRef{})
	if err := a.handleLongPress(context.Background(), ownframe.Box{ID: "xl-gc-3-3", X: 200, Y: 300, W: 80, H: 20}); err != nil {
		t.Fatal(err)
	}
	if a.overlay != "ctx" || !strings.Contains(a.page.HTML(), "rb-ctx-menu") || a.active != (CellRef{Row: 3, Col: 3}) {
		t.Fatalf("long press: overlay %q active %v", a.overlay, a.active)
	}
}

// ---- comment hover ----

func TestDCommentHoverPopover(t *testing.T) {
	a, _ := newTestApp(t)
	a.sheet().SetComment(2, 2, "check the Q3 number")
	a.redraw()
	if strings.Contains(a.page.HTML(), "check the Q3 number") {
		t.Fatal("comment text shown without a hover")
	}
	ctx := context.Background()
	x, y := cellPoint(t, a, CellRef{Row: 2, Col: 2})
	if err := a.page.Edit().Hover(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if a.hoverNote == nil || *a.hoverNote != (CellRef{Row: 2, Col: 2}) {
		t.Fatalf("hoverNote = %v", a.hoverNote)
	}
	pop := mustBox(t, a, "xl-gnote-pop")
	cell := mustBox(t, a, "xl-gc-2-2")
	if pop.X < cell.X+cell.W-1 && pop.X+pop.W > cell.X {
		t.Fatalf("note %v..%v overlaps its cell %v..%v", pop.X, pop.X+pop.W, cell.X, cell.X+cell.W)
	}
	if !strings.Contains(a.page.HTML(), "check the Q3 number") {
		t.Fatal("note text missing")
	}
	// Moving to a cell without a comment hides it.
	x, y = cellPoint(t, a, CellRef{Row: 6, Col: 6})
	if err := a.page.Edit().Hover(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if a.hoverNote != nil {
		t.Fatalf("hoverNote kept after leaving: %v", a.hoverNote)
	}
	if _, ok := box(a, "xl-gnote-pop"); ok {
		t.Fatal("note still on the page")
	}
}

// ---- web preview ----

func TestDWebEmptyCellsClickable(t *testing.T) {
	a, _ := newTestApp(t)
	a.web = true
	a.redraw()
	if _, ok := box(a, "xl-ge-20-4"); !ok {
		t.Fatal("no hit target for an empty cell in the web preview")
	}
	// The preview posts a bare click: no press precedes it.
	x, y := cellPoint(t, a, CellRef{Row: 12, Col: 6})
	if err := a.page.Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
	if a.active != (CellRef{Row: 12, Col: 6}) {
		t.Fatalf("empty cell click selected %v", a.active)
	}
	// The desktop build adds no hit targets.
	a.web = false
	a.redraw()
	if _, ok := box(a, "xl-ge-20-4"); ok {
		t.Fatal("hit targets rendered outside the web preview")
	}
}

// ---- EXL-34: function autocomplete ----

func TestDFormulaAutocomplete(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{Row: 40, Col: 8})
	typeText(t, a, "=AV")
	if !a.editing || a.editBuf != "=AV" {
		t.Fatalf("editing=%v buf=%q", a.editing, a.editBuf)
	}
	m := a.acMatches()
	if len(m) == 0 || m[0] != "AVERAGE" {
		t.Fatalf("matches %v", m)
	}
	sel := mustBox(t, a, "xl-ac-AVERAGE")
	if sel.Y <= 0 || !strings.Contains(a.page.HTML(), "AVERAGE(number1, [number2], ...)") {
		t.Fatalf("popover box %+v or signature missing", sel)
	}
	if !strings.Contains(a.page.HTML(), `aria-selected="true"`) {
		t.Fatal("no highlighted row")
	}

	// Down then Up returns to the first row; Enter inserts the function and keeps editing.
	key(t, a, "arrowdown")
	if a.acSel != 1 && len(m) > 1 {
		t.Fatalf("acSel after down = %d", a.acSel)
	}
	key(t, a, "arrowup")
	key(t, a, "enter")
	if !a.editing || a.editBuf != "=AVERAGE(" {
		t.Fatalf("after enter: editing=%v buf=%q", a.editing, a.editBuf)
	}
	if a.sheet().Raw(40, 8) != "" || a.active != (CellRef{Row: 40, Col: 8}) {
		t.Fatalf("enter committed or moved: raw %q active %v", a.sheet().Raw(40, 8), a.active)
	}
	if got := a.page.Form().FormValue("xl-fx"); got != "=AVERAGE(" {
		t.Fatalf("formula bar shows %q", got)
	}
	if len(a.acMatches()) != 0 {
		t.Fatalf("list stays open after the name is complete: %v", a.acMatches())
	}

	// Tab picks the highlighted row, and a click on a row does too.
	a.cancelEdit()
	a.selectCell(CellRef{Row: 41, Col: 8})
	typeText(t, a, "=1+SUMI")
	key(t, a, "tab")
	if a.editBuf != "=1+SUMIF(" || a.active != (CellRef{Row: 41, Col: 8}) {
		t.Fatalf("tab: buf %q active %v", a.editBuf, a.active)
	}
	a.cancelEdit()
	a.selectCell(CellRef{Row: 42, Col: 8})
	typeText(t, a, "=CO")
	if err := a.handleClick(context.Background(), ownframe.Box{ID: "xl-ac-COUNTA"}); err != nil {
		t.Fatal(err)
	}
	if a.editBuf != "=COUNTA(" {
		t.Fatalf("click: buf %q", a.editBuf)
	}

	// No list for plain text, a cell reference, or a name after a digit.
	a.cancelEdit()
	typeText(t, a, "SUM")
	if len(a.acMatches()) != 0 {
		t.Fatalf("list shown outside a formula: %v", a.acMatches())
	}
	key(t, a, "escape")
	typeText(t, a, "=A1+2SU")
	if len(a.acMatches()) != 0 {
		t.Fatalf("list shown after a digit: %v", a.acMatches())
	}
}

// ---- EXL-02: mounts at both documented sizes ----

func TestDMountsAtTwoSizes(t *testing.T) {
	a, _ := newTestApp(t)
	for _, sz := range [][2]int{{1440, 960}, {1280, 800}} {
		a.page.SetSize(sz[0], sz[1])
		a.redraw()
		if a.lastErr != nil {
			t.Fatal(a.lastErr)
		}
		if w, h := a.page.Size(); w != sz[0] || h != sz[1] {
			t.Fatalf("size %dx%d, want %v", w, h, sz)
		}
		status := mustBox(t, a, "xl-status")
		if int(status.Y+status.H) > sz[1] {
			t.Fatalf("status bar ends at %v past %d", status.Y+status.H, sz[1])
		}
	}
}
