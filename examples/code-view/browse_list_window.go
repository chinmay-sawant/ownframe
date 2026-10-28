package main

import "context"

// Browse folder list windowing. The engine scrolls the page as a whole: a
// wheel event goes to the Go handler first and otherwise pans the entire
// page (see internal/window/view.go wheel). Scroll state is page-level only,
// so there is no input path that moves an inner overflow:auto box. The
// browse dialog therefore windows its rows in Go the way the explorer tree
// does: buildView renders only the visible slice and the wheel handler
// below moves the offset. Entry markup, ids, and actions are unchanged, so
// clicks keep working.
const (
	// browseListVisible is the rows the panel shows: the 320px panel
	// max-height holds ten full 30px rows (see .file-item height).
	browseListVisible = 10
)

// browseListWindow is the vertical scroll model for the browse dialog
// folder list. Offset is the first visible row. Accum keeps the sub-row
// part of a wheel delta between events, so a trackpad moves the list once
// its fractions add up to a whole row.
type browseListWindow struct {
	Offset int
	Accum  float64
}

// windowed pins Offset to the scrollable range and returns the visible
// slice. A folder that fits leaves Offset at zero and shows every entry.
func (w *browseListWindow) windowed(entries []recentRootRow) []recentRootRow {
	w.Offset = browseClampOffset(w.Offset, len(entries))

	start := w.Offset
	end := min(start+browseListVisible, len(entries))
	if end < start {
		end = start
	}

	return entries[start:end]
}

// scrollBy moves Offset by rows, where a positive count moves toward the
// end of the list, and pins the result to the scrollable range.
func (w *browseListWindow) scrollBy(rows, total int) {
	w.Offset = browseClampOffset(w.Offset+rows, total)
}

// wheelRows turns a wheel delta measured in rows into whole rows to scroll,
// keeping the fraction in Accum for the next event. A positive delta
// scrolls toward the end, so a caller passes -dy multiplied by its rows
// per notch.
func (w *browseListWindow) wheelRows(delta float64) int {
	w.Accum += delta
	rows := int(w.Accum)
	w.Accum -= float64(rows)

	return rows
}

// reset parks the list back at the first row, used when the dialog opens
// or the root changes.
func (w *browseListWindow) reset() {
	w.Offset = 0
	w.Accum = 0
}

// browseClampOffset pins offset between the first row and the last full
// window. A list that fits the window stays at zero.
func browseClampOffset(offset, total int) int {
	maxOffset := total - browseListVisible
	if maxOffset < 0 {
		maxOffset = 0
	}

	return max(0, min(offset, maxOffset))
}

// browseWindowedEntries lists the browse root's subdirectories and returns
// only the visible window. It keeps entry order, markup, ids, and actions
// identical: buildView renders the slice the same way it rendered the full
// list, so clicks keep working.
func (a *App) browseWindowedEntries() []recentRootRow {
	return a.browseList.windowed(a.browseEntries())
}

// browseWheelAt scrolls the folder list when a vertical wheel lands on it.
// It reports handled true to swallow the wheel: the dialog is modal, so the
// page must not pan behind it. x and y are page coordinates, already
// adjusted for the page scroll offset by the caller.
func (a *App) browseWheelAt(ctx context.Context, x, y, dy float64) (bool, error) {
	if !a.browseOpen || dy == 0 {
		return false, nil
	}

	list, ok := findBoxID(a.page.Boxes(), "browse-list")
	if !ok || x < list.X || x > list.X+list.W || y < list.Y || y > list.Y+list.H {
		return false, nil
	}

	if rows := a.browseList.wheelRows(-dy * treeWheelRows); rows != 0 {
		a.browseList.scrollBy(rows, len(a.browseEntries()))
		a.refresh()
	}

	return true, a.page.Redraw(ctx)
}
