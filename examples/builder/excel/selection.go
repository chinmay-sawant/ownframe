package main

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	resizeSlop     = 3
	autoFitPad     = 14  // horizontal padding added to an auto-fit column
	maxColWidth    = 600 // widest column auto-fit may produce
	autoFitRowPad  = 9   // vertical padding added to an auto-fit row
	allCellsMaxRow = MaxRows - 1
	allCellsMaxCol = MaxCols - 1
)

func clampCell(c CellRef) CellRef {
	c.Row = min(max(c.Row, 0), MaxRows-1)
	c.Col = min(max(c.Col, 0), MaxCols-1)
	return c
}

func normRange(a, b CellRef) Range {
	return Range{
		Min: CellRef{Row: min(a.Row, b.Row), Col: min(a.Col, b.Col)},
		Max: CellRef{Row: max(a.Row, b.Row), Col: max(a.Col, b.Col)},
	}
}

// selectCell makes c the single active cell, committing any open edit.
func (a *app) selectCell(c CellRef) {
	a.commitEdit()
	c = clampCell(c)
	a.active, a.anchor = c, c
	a.sel = Range{Min: c, Max: c}
	a.ensureVisible(c)
	a.afterSelect()
	a.redraw()
}

// selectRange selects r with its top-left cell active.
func (a *app) selectRange(r Range) {
	a.commitEdit()
	r = normRange(clampCell(r.Min), clampCell(r.Max))
	a.active, a.anchor = r.Min, r.Min
	a.sel = r
	a.ensureVisible(r.Min)
	a.afterSelect()
	a.redraw()
}

// extendTo grows the selection from the anchor to c.
func (a *app) extendTo(c CellRef) {
	c = clampCell(c)
	a.sel = normRange(a.anchor, c)
	a.active = a.anchor
	a.ensureVisible(c)
}

// corner is the selection corner opposite the anchor, the one Shift+Arrow
// moves.
func (a *app) corner() CellRef {
	c := a.anchor
	if a.sel.Min.Row == a.anchor.Row {
		c.Row = a.sel.Max.Row
	} else {
		c.Row = a.sel.Min.Row
	}
	if a.sel.Min.Col == a.anchor.Col {
		c.Col = a.sel.Max.Col
	} else {
		c.Col = a.sel.Min.Col
	}
	return c
}

func (a *app) filled(r, c int) bool { return a.sheet().Raw(r, c) != "" }

// dataEdge implements Ctrl+Arrow: run to the end of the current block of
// filled cells, or to the next filled cell, or to the sheet edge.
func (a *app) dataEdge(from CellRef, dr, dc int) CellRef {
	in := func(r, c int) bool { return r >= 0 && r < MaxRows && c >= 0 && c < MaxCols }
	r, c := from.Row, from.Col
	nr, nc := r+dr, c+dc
	if !in(nr, nc) {
		return from
	}
	if a.filled(r, c) && a.filled(nr, nc) {
		for in(r+dr, c+dc) && a.filled(r+dr, c+dc) {
			r, c = r+dr, c+dc
		}
		return CellRef{Row: r, Col: c}
	}
	for in(r+dr, c+dc) {
		r, c = r+dr, c+dc
		if a.filled(r, c) {
			return CellRef{Row: r, Col: c}
		}
	}
	return CellRef{Row: r, Col: c}
}

// moveActive moves the active cell (or the extending corner) by one step.
func (a *app) moveActive(dr, dc int, extend, ctrl bool) {
	a.commitEdit()
	from := a.active
	if extend {
		from = a.corner()
	}
	var to CellRef
	if ctrl {
		to = a.dataEdge(from, dr, dc)
	} else {
		to = CellRef{Row: from.Row + dr, Col: from.Col + dc}
		for dr != 0 && to.Row >= 0 && to.Row < MaxRows && a.sheet().HiddenRow(to.Row) {
			to.Row += dr
		}
		to = clampCell(to)
	}
	if extend {
		a.extendTo(to)
		a.redraw()
		return
	}
	a.selectCell(to)
}

// stepInSelection moves the active cell inside a multi-cell selection,
// wrapping at its edges, as Enter and Tab do. A single cell just moves.
func (a *app) stepInSelection(dr, dc int) {
	if a.sel.Min == a.sel.Max {
		a.moveActive(dr, dc, false, false)
		return
	}
	n := a.active
	s := a.sel
	if dc != 0 {
		n.Col += dc
		if n.Col > s.Max.Col {
			n.Col = s.Min.Col
			n.Row++
		} else if n.Col < s.Min.Col {
			n.Col = s.Max.Col
			n.Row--
		}
	}
	if dr != 0 {
		n.Row += dr
		if n.Row > s.Max.Row {
			n.Row = s.Min.Row
			n.Col++
		} else if n.Row < s.Min.Row {
			n.Row = s.Max.Row
			n.Col--
		}
	}
	if n.Row > s.Max.Row {
		n.Row = s.Min.Row
	}
	if n.Row < s.Min.Row {
		n.Row = s.Max.Row
	}
	if n.Col > s.Max.Col {
		n.Col = s.Min.Col
	}
	if n.Col < s.Min.Col {
		n.Col = s.Max.Col
	}
	a.active = n
	a.ensureVisible(n)
	a.redraw()
}

// ---- editing ----

func (a *app) beginEdit(initial string) {
	if a.editing {
		return
	}
	a.editing = true
	a.editFromBar = false
	a.editBuf = initial
	a.editOrig = a.sheet().Raw(a.active.Row, a.active.Col)
	a.refActive = false
	a.ensureVisible(a.active)
	a.redraw()
	_ = a.focus("xl-editor")
}

// beginBarEdit starts an edit from the formula input, which already has
// focus because the user clicked it.
func (a *app) beginBarEdit() {
	if a.editing {
		return
	}
	a.editing = true
	a.editFromBar = true
	a.editOrig = a.sheet().Raw(a.active.Row, a.active.Col)
	a.editBuf = a.editOrig
	a.refActive = false
	a.redraw()
}

func (a *app) commitEdit() {
	if !a.editing {
		return
	}
	raw := a.editBuf
	a.editing, a.editFromBar, a.refActive = false, false, false
	if raw != a.editOrig {
		if ok, msg := a.validateEntry(a.active.Row, a.active.Col, raw); !ok {
			a.notify("Invalid entry", msg)
		} else {
			a.wb.Edit(a.sheet(), a.active.Row, a.active.Col, raw)
		}
	}
	_ = a.focus("")
	a.redraw()
}

func (a *app) cancelEdit() {
	if !a.editing {
		return
	}
	a.editing, a.editFromBar, a.refActive = false, false, false
	a.editBuf = ""
	_ = a.focus("")
	a.redraw()
}

func (a *app) clearSelection() {
	a.wb.ClearRange(a.sheet(), a.sel)
	a.redraw()
}

// applyFormat styles the selection through the workbook so undo sees it.
func (a *app) applyFormat(f func(*Style)) {
	a.wb.Format(a.sheet(), a.sel, f)
	a.redraw()
}

// ---- clipboard ----

func (a *app) copyText(cut bool) string {
	sel := a.sel
	tsv := a.sheet().CopyTSV(sel)
	if cut {
		a.cutRange, a.copyRange = &sel, nil
	} else {
		a.copyRange, a.cutRange = &sel, nil
	}
	a.redraw()
	return tsv
}

func (a *app) copy(cut bool) {
	ownframe.Write(a.copyText(cut))
}

func (a *app) paste() {
	text, err := ownframe.Read()
	if err != nil || text == "" {
		return
	}
	a.pasteText(text)
}

func (a *app) pasteText(text string) {
	if a.pasteCut() {
		return
	}
	if a.cutRange != nil {
		a.wb.ClearRange(a.sheet(), *a.cutRange)
		a.cutRange = nil
	}
	rng := a.wb.PasteTSV(a.sheet(), a.sel.Min, text)
	a.copyRange = nil
	a.selectRange(rng)
}

func (a *app) undo() {
	if a.editing {
		return
	}
	if a.wb.Undo() {
		a.redraw()
	}
}

func (a *app) redo() {
	if a.editing {
		return
	}
	if a.wb.Redo() {
		a.redraw()
	}
}

// ---- keyboard ----

// splitKey strips the shift+, ctrl+ and alt+ prefixes the window adds to
// the caret keys.
func splitKey(key string) (base string, shift, ctrl bool) {
	base = key
	for {
		switch {
		case strings.HasPrefix(base, "shift+"):
			shift, base = true, strings.TrimPrefix(base, "shift+")
		case strings.HasPrefix(base, "ctrl+"):
			ctrl, base = true, strings.TrimPrefix(base, "ctrl+")
		case strings.HasPrefix(base, "alt+"):
			base = strings.TrimPrefix(base, "alt+")
		default:
			return base, shift, ctrl
		}
	}
}

func (a *app) handleKey(ctx context.Context, key string) error {
	base, shift, ctrl := splitKey(key)
	mShift, mCtrl := a.mods()
	shift, ctrl = shift || mShift, ctrl || mCtrl
	focus := a.page.Form().FocusID()

	if a.overlay != "" {
		if base == "escape" {
			a.x.menu, a.x.toastOpen = "", false
			a.setOverlay("")
		}
		return nil
	}
	if focus == "xl-name" {
		switch base {
		case "enter":
			a.jumpToName()
		case "escape":
			_ = a.page.Form().Focus(ctx, "")
			a.page.Form().SetFormValue("xl-name", a.nameBoxText())
		}
		return nil
	}
	if ctrl && a.zoomKey(joinKey(base, shift, ctrl)) {
		// Zoom chords work while a cell is being edited too.
		a.redraw()
		return nil
	}
	if a.editing {
		a.editKey(base, shift)
		return nil
	}
	if focus != "" {
		return nil
	}
	if a.shortcutCtx(ctx, joinKey(base, shift, ctrl)) {
		return nil
	}
	if ctrl && a.ctrlKey(base, shift) {
		return nil
	}
	a.navKey(base, shift, ctrl)
	return nil
}

// joinKey rebuilds the "ctrl+shift+l" style name the shortcut table uses.
func joinKey(base string, shift, ctrl bool) string {
	name := base
	if shift {
		name = "shift+" + name
	}
	if ctrl {
		name = "ctrl+" + name
	}
	return name
}

func (a *app) editKey(base string, shift bool) {
	if a.acKey(base) {
		return
	}
	switch base {
	case "enter":
		a.commitEdit()
		a.stepRel(1, 0, shift)
	case "tab":
		a.commitEdit()
		a.stepRel(0, 1, shift)
	case "escape":
		a.cancelEdit()
	case "f4":
		a.cycleAbsolute()
	}
}

// stepRel moves down/right, or up/left with Shift.
func (a *app) stepRel(dr, dc int, reverse bool) {
	if reverse {
		dr, dc = -dr, -dc
	}
	a.stepInSelection(dr, dc)
}

func (a *app) ctrlKey(base string, shift bool) bool {
	switch base {
	case "p":
		a.setOverlay("command")
	case "s":
		a.setOverlay("save")
	case "home":
		a.selectOrExtend(CellRef{}, shift)
	case "end":
		a.selectOrExtend(a.sheet().UsedRange().Max, shift)
	default:
		return false
	}
	return true
}

func (a *app) selectOrExtend(c CellRef, extend bool) {
	if extend {
		a.extendTo(c)
		a.redraw()
		return
	}
	a.selectCell(c)
}

func (a *app) navKey(base string, shift, ctrl bool) {
	switch base {
	case "arrowup":
		a.moveActive(-1, 0, shift, ctrl)
	case "arrowdown":
		a.moveActive(1, 0, shift, ctrl)
	case "arrowleft":
		a.moveActive(0, -1, shift, ctrl)
	case "arrowright":
		a.moveActive(0, 1, shift, ctrl)
	case "tab":
		a.stepRel(0, 1, shift)
	case "enter":
		a.stepRel(1, 0, shift)
	case "home":
		a.selectOrExtend(CellRef{Row: a.active.Row}, shift)
	case "end":
		if ctrl {
			a.selectOrExtend(a.sheet().UsedRange().Max, shift)
		}
	case "pageup", "pagedown":
		n := a.visibleRowCount()
		if base == "pageup" {
			n = -n
		}
		to := clampCell(CellRef{Row: a.active.Row + n, Col: a.active.Col})
		a.scrollBy(n, 0)
		a.selectOrExtend(to, shift)
	case "f2":
		a.beginEdit(a.sheet().Raw(a.active.Row, a.active.Col))
	case "delete":
		a.clearSelection()
	case "escape":
		if a.cutRange != nil || a.copyRange != nil {
			a.cutRange, a.copyRange = nil, nil
			a.redraw()
		}
	}
}

// jumpToName reads the Name Box and selects the typed cell or range.
func (a *app) jumpToName() {
	text := strings.TrimSpace(a.page.Form().FormValue("xl-name"))
	rg, ok := ParseRange(strings.ToUpper(text))
	_ = a.focus("")
	if !ok {
		a.page.Form().SetFormValue("xl-name", a.nameBoxText())
		return
	}
	a.selectRange(rg)
}

// ---- formula reference insertion ----

var refTail = regexp.MustCompile(`(\$?[A-Za-z]{1,3}\$?[0-9]+)(:\$?[A-Za-z]{1,3}\$?[0-9]+)?$`)

// refInsertable reports whether a click should insert a reference: the
// buffer is a formula whose caret follows =, (, a comma, :, or an operator.
func refInsertable(buf string) bool {
	if !strings.HasPrefix(buf, "=") {
		return false
	}
	t := strings.TrimRight(buf, " ")
	if t == "" {
		return false
	}
	return strings.ContainsRune("=(,:+-*/^&<>%;", rune(t[len(t)-1]))
}

// insertRef appends (or, right after a previous point-and-click, replaces)
// the reference text for r.
func (a *app) insertRef(r Range) {
	text := r.String()
	if r.Min == r.Max {
		text = r.Min.A1()
	}
	if a.refActive && a.refEnd == len(a.editBuf) && a.refStart <= len(a.editBuf) {
		a.editBuf = a.editBuf[:a.refStart]
	} else {
		a.refStart = len(a.editBuf)
	}
	a.editBuf += text
	a.refEnd = len(a.editBuf)
	a.refActive = true
	form := a.page.Form()
	form.SetFormValue("xl-editor", a.editBuf)
	form.SetFormValue("xl-fx", a.editBuf)
	a.redraw()
}

// cycleAbsolute is F4: A1 -> $A$1 -> A$1 -> $A1 -> A1 on the reference
// that ends the buffer.
func (a *app) cycleAbsolute() {
	loc := refTail.FindStringSubmatchIndex(a.editBuf)
	if loc == nil {
		return
	}
	tok := a.editBuf[loc[0]:]
	parts := strings.Split(tok, ":")
	for i, p := range parts {
		parts[i] = cycleOneRef(p)
	}
	a.editBuf = a.editBuf[:loc[0]] + strings.Join(parts, ":")
	form := a.page.Form()
	form.SetFormValue("xl-editor", a.editBuf)
	form.SetFormValue("xl-fx", a.editBuf)
	a.redraw()
}

var refParts = regexp.MustCompile(`^(\$?)([A-Za-z]{1,3})(\$?)([0-9]+)$`)

func cycleOneRef(ref string) string {
	m := refParts.FindStringSubmatch(ref)
	if m == nil {
		return ref
	}
	colAbs, rowAbs := m[1] == "$", m[3] == "$"
	switch {
	case !colAbs && !rowAbs:
		colAbs, rowAbs = true, true
	case colAbs && rowAbs:
		colAbs, rowAbs = false, true
	case !colAbs && rowAbs:
		colAbs, rowAbs = true, false
	default:
		colAbs, rowAbs = false, false
	}
	out := ""
	if colAbs {
		out += "$"
	}
	out += m[2]
	if rowAbs {
		out += "$"
	}
	return out + m[4]
}

// ---- pointer ----

// handleDragStart claims presses that land on the grid so the page tracks
// the gesture and sends DragMove and DragEnd.
func (a *app) handleDragStart(_ context.Context, box ownframe.Box) (bool, error) {
	if a.chartDragStart(box) {
		a.redraw()
		return true, nil
	}
	return a.overlay == "" && isGridID(box.ID), nil
}

// pointer maps window coordinates to grid-viewport coordinates.
func (a *app) pointer(x, y float64) (lx, ly float64) {
	gx, gy := a.gridOrigin()
	return x - gx, y - gy
}

// tapEchoGap is how soon after a claimed drag ends a further Press is the
// window's own tap echo. After a mouse press the window ends the claimed
// gesture and then replays Press, SelectAt and Click for a tap, so one physical
// click reaches handlePress twice back to back. A person cannot click twice
// this fast.
const tapEchoGap = 40 * time.Millisecond

func (a *app) handlePress(_ context.Context, box ownframe.Box, x, y float64) (bool, error) {
	// The grid claims presses before any click arrives, so an open dropdown
	// would stay up over a cell click. Close it here.
	if a.x.menu != "" && !strings.HasPrefix(box.Action, "xl:") {
		a.x.menu = ""
		a.redraw()
	}
	if a.overlay == "" && isGridID(box.ID) && time.Since(a.dragEndAt) < a.tapEcho {
		a.pressSeen = time.Now()
		return true, nil
	}
	a.drag = dragNone
	a.dragMoved = false
	if box.ID == "xl-fx" && a.overlay == "" {
		a.beginBarEdit()
		return false, nil
	}
	if a.overlay != "" || !isGridID(box.ID) {
		return false, nil
	}
	a.pressSeen = time.Now()
	if box.ID == "xl-gov-handle" && !a.editing {
		a.drag, a.fillSet = dragFill, false
		a.dragPressX, a.dragPressY = a.pointer(x, y)
		a.dragAccX, a.dragAccY = 0, 0
		return true, nil
	}
	if a.clickFilterChevron(box.ID) {
		return true, nil
	}
	lx, ly := a.pointer(x, y)
	a.dragPressX, a.dragPressY = lx, ly
	a.dragAccX, a.dragAccY = 0, 0
	vw, vh := a.gridViewportSize()
	hx, hy := a.rowHdrW(), a.colHdrH()
	cw, ch := a.canvasSize()
	cols, rows := a.layoutCols(cw), a.layoutRows(ch)
	shift, _ := a.mods()

	if a.pressScrollbar(lx, ly, vw, vh) {
		return true, nil
	}
	switch {
	case !a.x.hideHeads && lx < float64(hx) && ly < float64(hy):
		a.selectRange(Range{Max: CellRef{Row: allCellsMaxRow, Col: allCellsMaxCol}})
		return true, nil
	case !a.x.hideHeads && ly < float64(hy):
		a.pressColHeader(int(lx)-hx, cols, shift)
		return true, nil
	case !a.x.hideHeads && lx < float64(hx):
		a.pressRowHeader(int(ly)-hy, rows, shift)
		return true, nil
	}
	cell, ok := a.cellAt(lx, ly, cols, rows)
	if !ok {
		return true, nil
	}
	a.pressCell(cell, lx, ly, shift)
	return true, nil
}

func (a *app) pressCell(cell CellRef, lx, ly float64, shift bool) {
	if a.editing && refInsertable(a.editBuf) {
		a.drag = dragRefRange
		a.anchor = cell // temporary anchor for the dragged reference
		a.insertRef(Range{Min: cell, Max: cell})
		return
	}
	if a.pressValidation(cell, lx, ly) {
		return
	}
	a.commitEdit()
	a.drag = dragCells
	now := time.Now()
	if !shift && cell == a.lastCell && now.Sub(a.lastPress) < doubleClickGap {
		a.lastPress = time.Time{}
		a.selectCell(cell)
		a.beginEdit(a.sheet().Raw(cell.Row, cell.Col))
		return
	}
	a.lastPress, a.lastCell = now, cell
	_ = a.focus("")
	if shift {
		a.extendTo(cell)
		a.redraw()
		return
	}
	a.selectCell(cell)
}

// pressValidation opens the dropdown when the active validated cell's
// arrow is hit.
func (a *app) pressValidation(cell CellRef, lx, ly float64) bool {
	if cell != a.active || len(a.sheet().ValidationList(cell.Row, cell.Col)) == 0 {
		return false
	}
	cw, ch := a.canvasSize()
	slot, ok := slotFor(a.layoutCols(cw), cell.Col)
	if !ok {
		return false
	}
	right := float64(a.rowHdrW() + slot.pos + slot.size)
	if lx >= right-17 && lx <= right {
		a.setOverlay("validation:list")
		_ = ch
		return true
	}
	return false
}

func (a *app) pressColHeader(px int, cols []axisSlot, shift bool) {
	for _, s := range cols {
		if abs(px-(s.pos+s.size)) <= resizeSlop {
			a.drag, a.dragIndex, a.dragOrig = dragResizeCol, s.idx, a.unz(s.size)
			if now := time.Now(); s.idx == a.lastCell.Col && now.Sub(a.lastPress) < doubleClickGap && a.lastCell.Row == -2 {
				a.autoFitCol(s.idx)
				a.drag = dragNone
				a.lastPress = time.Time{}
				return
			}
			a.lastCell, a.lastPress = CellRef{Row: -2, Col: s.idx}, time.Now()
			return
		}
	}
	s, ok := slotAt(cols, px)
	if !ok {
		return
	}
	a.drag = dragColHeader
	a.commitEdit()
	if shift {
		a.sel = Range{Min: CellRef{Col: min(a.anchor.Col, s.idx)}, Max: CellRef{Row: allCellsMaxRow, Col: max(a.anchor.Col, s.idx)}}
		a.redraw()
		return
	}
	a.active = CellRef{Row: a.scrollRow, Col: s.idx}
	a.active.Row = 0
	a.anchor = a.active
	a.sel = Range{Min: CellRef{Col: s.idx}, Max: CellRef{Row: allCellsMaxRow, Col: s.idx}}
	a.redraw()
}

func (a *app) pressRowHeader(py int, rows []axisSlot, shift bool) {
	for _, s := range rows {
		if abs(py-(s.pos+s.size)) <= resizeSlop {
			a.drag, a.dragIndex, a.dragOrig = dragResizeRow, s.idx, a.unz(s.size)
			if now := time.Now(); s.idx == a.lastCell.Row && now.Sub(a.lastPress) < doubleClickGap && a.lastCell.Col == -2 {
				a.autoFitRow(s.idx)
				a.drag = dragNone
				a.lastPress = time.Time{}
				return
			}
			a.lastCell, a.lastPress = CellRef{Row: s.idx, Col: -2}, time.Now()
			return
		}
	}
	s, ok := slotAt(rows, py)
	if !ok {
		return
	}
	a.drag = dragRowHeader
	a.commitEdit()
	if shift {
		a.sel = Range{Min: CellRef{Row: min(a.anchor.Row, s.idx)}, Max: CellRef{Row: max(a.anchor.Row, s.idx), Col: allCellsMaxCol}}
		a.redraw()
		return
	}
	a.active = CellRef{Row: s.idx}
	a.anchor = a.active
	a.sel = Range{Min: CellRef{Row: s.idx}, Max: CellRef{Row: s.idx, Col: allCellsMaxCol}}
	a.redraw()
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// autoFitCol sizes a column to its widest populated cell.
func (a *app) autoFitCol(col int) {
	sh := a.sheet()
	best := minColWidth
	used := sh.UsedRange()
	for r := used.Min.Row; r <= used.Max.Row; r++ {
		text := sh.Display(r, col)
		if text == "" {
			continue
		}
		size := 13
		if fs := sh.StyleAt(r, col).FontSize; fs > 0 {
			size = fs
		}
		best = max(best, len([]rune(text))*(size*6/10+1)+autoFitPad)
	}
	sh.SetColWidth(col, min(best, maxColWidth))
	a.redraw()
}

// autoFitRow sizes a row to its tallest font.
func (a *app) autoFitRow(row int) {
	sh := a.sheet()
	best := defaultRowHeight
	for c := 0; c < MaxCols; c++ {
		if sh.Raw(row, c) == "" {
			continue
		}
		if fs := sh.StyleAt(row, c).FontSize; fs > 0 {
			best = max(best, fs+autoFitRowPad)
		}
	}
	sh.SetRowHeight(row, best)
	a.redraw()
}

// cellAt maps grid-viewport coordinates to a cell. Points beyond the last
// visible slot clamp to it so a drag past the edge keeps extending.
func (a *app) cellAt(lx, ly float64, cols, rows []axisSlot) (CellRef, bool) {
	if len(cols) == 0 || len(rows) == 0 {
		return CellRef{}, false
	}
	px, py := int(lx)-a.rowHdrW(), int(ly)-a.colHdrH()
	c, okc := slotAt(cols, px)
	r, okr := slotAt(rows, py)
	if !okc || !okr {
		return CellRef{}, false
	}
	return a.mergeAnchor(CellRef{Row: r.idx, Col: c.idx}), true
}

// clampedSlot is slotAt that clamps to the first or last slot.
func clampedSlot(slots []axisSlot, p int) axisSlot {
	if s, ok := slotAt(slots, p); ok {
		return s
	}
	if p < 0 || len(slots) == 0 {
		if len(slots) == 0 {
			return axisSlot{}
		}
		return slots[0]
	}
	return slots[len(slots)-1]
}

func (a *app) handleDragMove(_ context.Context, dx, dy float64) error {
	if a.x.dragChart != 0 {
		if a.chartDragMove(dx, dy) {
			a.redraw()
		}
		return nil
	}
	if a.drag == dragNone {
		return nil
	}
	a.dragAccX += dx
	a.dragAccY += dy
	a.dragMoved = true
	lx, ly := a.dragPressX+a.dragAccX, a.dragPressY+a.dragAccY
	vw, vh := a.gridViewportSize()
	cw, ch := a.canvasSize()
	hx, hy := a.rowHdrW(), a.colHdrH()

	switch a.drag {
	case dragResizeCol:
		a.sheet().SetColWidth(a.dragIndex, max(minColWidth, a.dragOrig+a.unz(int(a.dragAccX))))
	case dragResizeRow:
		a.sheet().SetRowHeight(a.dragIndex, max(minRowHeight, a.dragOrig+a.unz(int(a.dragAccY))))
	case dragVThumb:
		a.dragThumb(true, vw, vh)
	case dragHThumb:
		a.dragThumb(false, vw, vh)
	case dragFill:
		cols, rows := a.layoutCols(cw), a.layoutRows(ch)
		a.fillTo = CellRef{
			Row: clampedSlot(rows, int(ly)-hy).idx,
			Col: clampedSlot(cols, int(lx)-hx).idx,
		}
		a.fillSet = true
	case dragCells, dragRefRange:
		// Auto-scroll when the pointer leaves the canvas.
		switch {
		case ly > float64(vh):
			a.scrollBy(1, 0)
		case ly < float64(hy):
			a.scrollBy(-1, 0)
		}
		switch {
		case lx > float64(vw):
			a.scrollBy(0, 1)
		case lx < float64(hx):
			a.scrollBy(0, -1)
		}
		cols, rows := a.layoutCols(cw), a.layoutRows(ch)
		cell := CellRef{
			Row: clampedSlot(rows, int(ly)-hy).idx,
			Col: clampedSlot(cols, int(lx)-hx).idx,
		}
		if a.drag == dragRefRange {
			a.insertRef(normRange(a.anchor, cell))
			return nil
		}
		a.extendTo(cell)
	case dragColHeader:
		cols := a.layoutCols(cw)
		s := clampedSlot(cols, int(lx)-hx)
		lo, hi := min(a.anchor.Col, s.idx), max(a.anchor.Col, s.idx)
		a.sel = Range{Min: CellRef{Col: lo}, Max: CellRef{Row: allCellsMaxRow, Col: hi}}
	case dragRowHeader:
		rows := a.layoutRows(ch)
		s := clampedSlot(rows, int(ly)-hy)
		lo, hi := min(a.anchor.Row, s.idx), max(a.anchor.Row, s.idx)
		a.sel = Range{Min: CellRef{Row: lo}, Max: CellRef{Row: hi, Col: allCellsMaxCol}}
	}
	a.redraw()
	return nil
}

func (a *app) handleDragEnd(_ context.Context) error {
	a.dragEndAt = time.Now()
	if a.chartDragEnd() {
		a.redraw()
		return nil
	}
	if a.drag == dragNone {
		return nil
	}
	if a.drag == dragFill && a.fillSet && !a.sel.Contains(a.fillTo) {
		rng := a.wb.AutoFill(a.sheet(), a.sel, a.fillTo)
		a.selectRange(rng)
	}
	a.drag = dragNone
	a.dragMoved = false
	a.redraw()
	return nil
}

// pressScrollbar handles a press on either scrollbar: a thumb press starts
// a drag, a track press pages toward the pointer.
func (a *app) pressScrollbar(lx, ly float64, vw, vh int) bool {
	vt, ht := a.vTrack(vw, vh), a.hTrack(vw, vh)
	mr, mc := a.maxScroll()
	fr, fc := a.sheet().FreezeRows(), a.sheet().FreezeCols()
	cw, _ := a.canvasSize()
	fw, _ := a.frozenExtent()
	inV := lx >= float64(vt[0]) && lx <= float64(vt[0]+vt[2]) && ly >= float64(vt[1]) && ly <= float64(vt[1]+vt[3])
	inH := lx >= float64(ht[0]) && lx <= float64(ht[0]+ht[2]) && ly >= float64(ht[1]) && ly <= float64(ht[1]+ht[3])
	if inV {
		off, length := thumbGeom(vt[3], a.scrollRow-fr, mr-fr, MaxRows-fr, a.visibleRowCount())
		rel := int(ly) - vt[1]
		switch {
		case rel >= off && rel <= off+length:
			a.drag, a.dragOrig = dragVThumb, a.scrollRow
		case rel < off:
			a.scrollBy(-a.visibleRowCount(), 0)
			a.redraw()
		default:
			a.scrollBy(a.visibleRowCount(), 0)
			a.redraw()
		}
		return true
	}
	if inH {
		hvis := max((cw-fw)/a.zs(defaultColWidth), 1)
		off, length := thumbGeom(ht[2], a.scrollCol-fc, mc-fc, MaxCols-fc, hvis)
		rel := int(lx) - ht[0]
		switch {
		case rel >= off && rel <= off+length:
			a.drag, a.dragOrig = dragHThumb, a.scrollCol
		case rel < off:
			a.scrollBy(0, -hvis)
			a.redraw()
		default:
			a.scrollBy(0, hvis)
			a.redraw()
		}
		return true
	}
	return false
}

// dragThumb converts thumb travel into a scroll position.
func (a *app) dragThumb(vertical bool, vw, vh int) {
	mr, mc := a.maxScroll()
	fr, fc := a.sheet().FreezeRows(), a.sheet().FreezeCols()
	cw, _ := a.canvasSize()
	fw, _ := a.frozenExtent()
	if vertical {
		vt := a.vTrack(vw, vh)
		_, length := thumbGeom(vt[3], 0, mr-fr, MaxRows-fr, a.visibleRowCount())
		travel := max(vt[3]-length, 1)
		a.scrollRow = a.dragOrig + int(a.dragAccY*float64(mr-fr)/float64(travel))
	} else {
		ht := a.hTrack(vw, vh)
		hvis := max((cw-fw)/a.zs(defaultColWidth), 1)
		_, length := thumbGeom(ht[2], 0, mc-fc, MaxCols-fc, hvis)
		travel := max(ht[2]-length, 1)
		a.scrollCol = a.dragOrig + int(a.dragAccX*float64(mc-fc)/float64(travel))
	}
	a.clampScroll()
}
