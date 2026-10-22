package ui

// scrollTo queues an absolute scroll target and remembers it for the next
// scroll-window callback, so a page swap lays out at the destination offset
// in one redraw.
func (a *App) scrollTo(y int) {
	if y < 0 {
		y = 0
	}

	a.pendingOffset = y
	a.havePending = true
	a.page.ScrollTo(0, y)
}

// scrollTop puts the first row at the top of the list.
func (a *App) scrollTop() { a.scrollTo(0) }

// scrollBottom puts the last row at the bottom of the viewport.
func (a *App) scrollBottom() {
	y := a.contentHeight() - a.viewHeight()
	if y < 0 {
		y = 0
	}

	a.scrollTo(y)
}

// listHeight is the laid-out height of the whole list page.
func (a *App) listHeight() int {
	return HeaderH + a.pager.Len()*RowH + StatusH
}

// contentHeight is the height of whatever mode the page is in.
func (a *App) contentHeight() int {
	if a.detailOpen {
		return HeaderH + len(a.detail.Lines)*18 + StatusH
	}

	return a.listHeight()
}

// viewHeight is the last viewport height the window reported.
func (a *App) viewHeight() int {
	if a.viewH > 0 {
		return a.viewH
	}

	if _, h := a.page.Size(); h > 0 {
		return h
	}

	return uiDefHeight
}

// atBottom reports whether a viewport at offsetY with height viewH shows
// the last row fully.
func (a *App) atBottom(offsetY, viewH int) bool {
	return offsetY+viewH >= a.listHeight()-2*RowH
}
