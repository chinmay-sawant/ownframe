package ui

// Pin is the page scroll-window callback. It slices the visible rows plus
// overscan for a scroll offset and moves the pinned bars before the template
// executes. It returns true only when the frame content changed.
func (a *App) Pin(offsetY, viewH int) bool {
	wasH := a.lastViewH
	a.viewH = viewH

	off := offsetY
	if a.havePending {
		off = a.pendingOffset
		if off == offsetY {
			a.havePending = false
		}
	}

	changed := a.view.BarTop != off || a.view.SideH != viewH ||
		a.view.StatusTop != off+viewH-StatusH || wasH != viewH

	a.view.BarTop = off
	a.view.SideH = viewH
	a.view.StatusTop = off + viewH - StatusH
	a.lastViewH = viewH

	if a.view.Mode != "detail" {
		start, end := visibleRange(a.pager.Len(), off, viewH, HeaderH, RowH, Overscan)
		if start != a.winStart || end != a.winEnd || a.winEmpty != a.pager.Empty() {
			a.winStart, a.winEnd = start, end
			a.winEmpty = a.pager.Empty()
			a.setWindow(start, end)
			changed = true
		}
	}

	a.handleResize(wasH, viewH, offsetY)

	if !a.detailOpen {
		a.pager.SetAnchorFromOffset(off, HeaderH)
	}

	return changed
}

// handleResize keeps the reading anchor (or the bottom) when the viewport
// height changes under a scroll offset that the window left alone.
func (a *App) handleResize(wasH, viewH, offsetY int) {
	if a.detailOpen || wasH == 0 || wasH == viewH || a.havePending {
		return
	}

	if a.atBottom(offsetY, wasH) {
		a.scrollBottom()

		return
	}

	a.scrollTo(a.pager.AnchorOffset(HeaderH))
}

// setWindow writes the visible slice and its spacer heights into the view.
func (a *App) setWindow(start, end int) {
	a.view.Rows = buildRows(a.pager.Entries[start:end], a.width, a.selected)
	a.view.TopPad = HeaderH + padTop(start, RowH)
	a.view.BotPad = padBottom(a.pager.Len(), end, RowH) + StatusH

	if a.pager.Empty() {
		a.view.Empty = a.emptyText()
	}
}
