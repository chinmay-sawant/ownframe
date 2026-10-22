package ui

import (
	"context"
	"time"
)

// moveSelection moves the selected row and keeps it inside the viewport.
// shift keeps the range anchor; a plain arrow collapses the range.
func (a *App) moveSelection(ctx context.Context, delta int, extend bool) error {
	if a.pager.Empty() || a.detailOpen {
		return nil
	}

	idx, _ := a.pager.IndexOf(a.selected)
	idx += delta

	if idx < 0 {
		idx = 0
	}

	if idx >= a.pager.Len() {
		idx = a.pager.Len() - 1
	}

	a.selected = a.pager.Entries[idx].ID
	if !extend {
		a.selAnchor = a.selected
	}

	a.revealRow(idx)
	a.setWindow(a.winStart, a.winEnd)

	return a.draw(ctx)
}

// revealRow scrolls the minimum amount that shows row idx.
func (a *App) revealRow(idx int) {
	rowY := HeaderH + idx*RowH
	_, off := a.page.Scroll().ScrollOffset()
	viewH := a.viewHeight()

	if rowY < off {
		a.scrollTo(rowY)

		return
	}

	if rowY+RowH > off+viewH {
		a.scrollTo(rowY + RowH - viewH)
	}
}

// jumpOldest loads the first page of the frozen result set.
func (a *App) jumpOldest() {
	if a.pager.Empty() && a.pager.Total == 0 {
		return
	}

	a.freeze()
	a.follow.SetFollow(false)
	a.loadPage(intentOldest)
	a.saveSettings()
}

// sourcesInterval is how often the sidebar snapshot refreshes.
const sourcesInterval = 3 * time.Second
