package ui

// scrollWindow recomputes the rendered window for a scroll offset. The
// window owns the offset and calls this before Redraw and after every
// offset change. It always returns true: the frozen chrome is positioned
// from the offset, so any move needs a redraw and fresh hit boxes.
//
// The rendered window is exactly the visible cells; the fetched window
// adds FetchOver rows and columns, so a small scroll reuses the tiles.
func (a *App) scrollWindow(offsetY, viewH int) bool {
	offsetX, _ := a.page.Scroll().ScrollOffset()
	a.scrollX, a.scrollY = offsetX, offsetY

	sh := a.sheet()
	w, _ := a.page.Size()
	a.win = computeWindow(a.active, offsetX, offsetY, w, viewH, sh.Rows, sh.Cols, 0)

	fetch := computeWindow(a.active, offsetX, offsetY, w, viewH, sh.Rows, sh.Cols, FetchOver)
	if fetch != a.fetchWin {
		a.fetchWin = fetch
		a.requestArea(fetch.Area())
	}

	a.refresh()

	return true
}

// ensureVisible queues a scroll that brings the active cell inside the
// viewport, with one cell of margin.
func (a *App) ensureVisible() {
	s := a.selection()
	w, h := a.page.Size()
	x, y := a.scrollX, a.scrollY
	cx, cy := cellX(s.ActiveC), cellY(s.ActiveR)
	if cx < x+HeadW {
		x = max(0, cx-HeadW)
	} else if cx+ColW > x+w {
		x = cx + ColW - w
	}

	if cy < y+ChromeH {
		y = max(0, cy-ChromeH)
	} else if cy+RowH > y+h {
		y = cy + RowH - h
	}

	if x != a.scrollX || y != a.scrollY {
		a.page.ScrollTo(x, y)
	}
}
