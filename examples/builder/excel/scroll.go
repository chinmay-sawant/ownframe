package main

import (
	"context"
	"math"
)

// Initial capacities for the visible column and row slot lists.
const (
	colSlotCap = 24
	rowSlotCap = 48

	zoomBase       = 100 // zoom percentage that leaves sizes unscaled
	minThumbLength = 24  // shortest scrollbar thumb in pixels
)

// zs scales a pixel size by the zoom percentage: cells, header bands and the
// chrome bands all go through it.
func (a *app) zs(n int) int {
	return scaleBy(n, a.zoomPct())
}

// zf and unzf are zs and unz for fractional pixel values.
func (a *app) zf(v float64) float64 { return v * float64(a.zoomPct()) / zoomBase }

func (a *app) unzf(v float64) float64 { return v * zoomBase / float64(a.zoomPct()) }

// unz is the inverse of zs, for turning a pointer distance back into a sheet
// size.
func (a *app) unz(n int) int { return n * zoomBase / a.zoomPct() }

// axisSlot is one visible row or column: its index, its pixel offset inside
// the canvas (the area right of the row headers and under the column
// headers) and its size.
type axisSlot struct{ idx, pos, size int }

// canvasSize is the pixel box of the cell canvas inside the grid viewport.
func (a *app) canvasSize() (int, int) {
	w, h := a.gridViewportSize()
	w -= a.rowHdrW()
	h -= a.colHdrH()
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return w, h
}

// rowHdrW and colHdrH collapse to zero when the headings are hidden.
func (a *app) rowHdrW() int {
	if a.x.hideHeads {
		return 0
	}
	return a.zs(rowHeaderWidth)
}

func (a *app) colHdrH() int {
	if a.x.hideHeads {
		return 0
	}
	return a.zs(colHeaderHeight)
}

// layoutCols returns the visible columns for a canvas width: frozen columns
// first, then the scrollable ones starting at scrollCol.
func (a *app) layoutCols(width int) []axisSlot {
	sh := a.sheet()
	freeze := sh.FreezeCols()
	slots := make([]axisSlot, 0, colSlotCap)
	pos := 0
	for c := 0; c < freeze && c < MaxCols; c++ {
		w := a.zs(sh.ColWidth(c))
		slots = append(slots, axisSlot{c, pos, w})
		pos += w
	}
	scrolled := 0
	for c := max(a.scrollCol, freeze); c < MaxCols && pos < width; c++ {
		w := a.zs(sh.ColWidth(c))
		if pos+w > width && scrolled > 0 {
			break // a partly visible column is not drawn: clipped text would bleed
		}
		scrolled++
		slots = append(slots, axisSlot{c, pos, w})
		pos += w
	}
	return slots
}

// layoutRows is layoutCols for rows. Filtered-out rows take no space.
func (a *app) layoutRows(height int) []axisSlot {
	sh := a.sheet()
	freeze := sh.FreezeRows()
	slots := make([]axisSlot, 0, rowSlotCap)
	pos := 0
	for r := 0; r < freeze && r < MaxRows; r++ {
		if sh.HiddenRow(r) {
			continue
		}
		h := a.zs(sh.RowHeight(r))
		slots = append(slots, axisSlot{r, pos, h})
		pos += h
	}
	scrolled := 0
	for r := max(a.scrollRow, freeze); r < MaxRows && pos < height; r++ {
		if sh.HiddenRow(r) {
			continue
		}
		h := a.zs(sh.RowHeight(r))
		if pos+h > height && scrolled > 0 {
			break
		}
		scrolled++
		slots = append(slots, axisSlot{r, pos, h})
		pos += h
	}
	return slots
}

// frozenExtent is the pixel size of the frozen columns and rows.
func (a *app) frozenExtent() (w, h int) {
	sh := a.sheet()
	for c := 0; c < sh.FreezeCols(); c++ {
		w += a.zs(sh.ColWidth(c))
	}
	for r := 0; r < sh.FreezeRows(); r++ {
		if !sh.HiddenRow(r) {
			h += a.zs(sh.RowHeight(r))
		}
	}
	return w, h
}

func slotFor(slots []axisSlot, idx int) (axisSlot, bool) {
	for _, s := range slots {
		if s.idx == idx {
			return s, true
		}
	}
	return axisSlot{}, false
}

// slotAt finds the slot under a canvas coordinate.
func slotAt(slots []axisSlot, p int) (axisSlot, bool) {
	for _, s := range slots {
		if p >= s.pos && p < s.pos+s.size {
			return s, true
		}
	}
	return axisSlot{}, false
}

// spanOf returns the pixel extent of the visible slots whose index lies in
// [lo, hi]. ok is false when none of them is on screen.
func spanOf(slots []axisSlot, lo, hi int) (pos, size int, ok bool) {
	first, last := -1, -1
	for i, s := range slots {
		if s.idx >= lo && s.idx <= hi {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return 0, 0, false
	}
	return slots[first].pos, slots[last].pos + slots[last].size - slots[first].pos, true
}

// maxScroll bounds scrollRow and scrollCol so the last row and column can
// still reach the top-left of the canvas area.
func (a *app) maxScroll() (rows, cols int) {
	w, h := a.canvasSize()
	_, fh := a.frozenExtent()
	fw, _ := a.frozenExtent()
	visRows := max((h-fh)/a.zs(defaultRowHeight), 1)
	visCols := max((w-fw)/a.zs(defaultColWidth), 1)
	return max(MaxRows-visRows, a.sheet().FreezeRows()), max(MaxCols-visCols, a.sheet().FreezeCols())
}

func (a *app) clampScroll() {
	mr, mc := a.maxScroll()
	a.scrollRow = min(max(a.scrollRow, a.sheet().FreezeRows()), mr)
	a.scrollCol = min(max(a.scrollCol, a.sheet().FreezeCols()), mc)
}

// scrollBy moves the window by whole rows and columns.
func (a *app) scrollBy(dRows, dCols int) bool {
	beforeR, beforeC := a.scrollRow, a.scrollCol
	a.scrollRow += dRows
	a.scrollCol += dCols
	a.clampScroll()
	return a.scrollRow != beforeR || a.scrollCol != beforeC
}

// ensureVisible scrolls just far enough to show the cell.
func (a *app) ensureVisible(c CellRef) {
	sh := a.sheet()
	w, h := a.canvasSize()
	fw, fh := a.frozenExtent()
	if c.Row >= sh.FreezeRows() {
		if c.Row < a.scrollRow {
			a.scrollRow = c.Row
		} else if !a.fitsFrom(c.Row, a.scrollRow, h-fh, false) {
			a.scrollRow = a.firstFor(c.Row, h-fh, false)
		}
	}
	if c.Col >= sh.FreezeCols() {
		if c.Col < a.scrollCol {
			a.scrollCol = c.Col
		} else if !a.fitsFrom(c.Col, a.scrollCol, w-fw, true) {
			a.scrollCol = a.firstFor(c.Col, w-fw, true)
		}
	}
	a.clampScroll()
}

func (a *app) axisSize(i int, cols bool) int {
	sh := a.sheet()
	if cols {
		return a.zs(sh.ColWidth(i))
	}
	if sh.HiddenRow(i) {
		return 0
	}
	return a.zs(sh.RowHeight(i))
}

// fitsFrom reports whether index target is fully visible when the scrolling
// region starts at first and has room pixels.
func (a *app) fitsFrom(target, first, room int, cols bool) bool {
	used := 0
	limit := MaxRows
	if cols {
		limit = MaxCols
	}
	for i := first; i < limit; i++ {
		used += a.axisSize(i, cols)
		if i == target {
			return used <= room
		}
		if used > room {
			return false
		}
	}
	return false
}

// firstFor returns the first scrolled index that leaves target fully visible
// in the last position.
func (a *app) firstFor(target, room int, cols bool) int {
	used := 0
	first := target
	for i := target; i >= 0; i-- {
		used += a.axisSize(i, cols)
		if used > room {
			break
		}
		first = i
	}
	return first
}

// visibleRowCount is how many whole scrolling rows fit the canvas.
func (a *app) visibleRowCount() int {
	_, h := a.canvasSize()
	_, fh := a.frozenExtent()
	n, used := 0, 0
	for r := max(a.scrollRow, a.sheet().FreezeRows()); r < MaxRows; r++ {
		used += a.axisSize(r, false)
		if used > h-fh {
			break
		}
		n++
	}
	return max(n, 1)
}

// handleWheel scrolls the grid. Vertical deltas move rows, Shift+wheel or a
// horizontal delta moves columns. Ebiten reports a negative dy for a wheel
// pushed toward the user, so content moves down for negative dy.
func (a *app) handleWheel(_ context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 {
		return false, nil
	}
	if _, ctrl := a.mods(); ctrl {
		// Ctrl+wheel zooms the whole UI instead of scrolling. Wheel away from
		// the user (positive dy) zooms in.
		if dy != 0 {
			step := zoomStepPct
			if dy < 0 {
				step = -step
			}
			a.stepZoom(step)
		}
		return true, nil
	}
	if a.overlay != "" {
		return false, nil
	}
	ox, oy := a.gridOrigin()
	w, h := a.gridViewportSize()
	if x < ox || y < oy || x > ox+float64(w) || y > oy+float64(h) {
		return false, nil
	}
	shift, _ := a.mods()
	if shift && dx == 0 {
		dx, dy = dy, 0
	}
	a.wheelRow += -dy * rowsPerNotch
	a.wheelCol += -dx * colsPerNotch
	moveR := int(math.Trunc(a.wheelRow))
	moveC := int(math.Trunc(a.wheelCol))
	a.wheelRow -= float64(moveR)
	a.wheelCol -= float64(moveC)
	if (moveR != 0 || moveC != 0) && a.scrollBy(moveR, moveC) {
		a.redraw()
	}
	return true, nil
}

// vThumb and hThumb return the scrollbar thumb geometry in canvas-space
// pixels along the track: offset and length. track is the track length.
func thumbGeom(track, scroll, maxScroll, total, visible int) (off, length int) {
	if total <= 0 {
		return 0, track
	}
	length = max(minThumbLength, track*visible/total)
	length = min(length, track)
	if maxScroll <= 0 {
		return 0, length
	}
	off = (track - length) * scroll / maxScroll
	return off, length
}
