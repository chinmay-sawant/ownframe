package main

// Tab strip windowing. A tab keeps its full measured width so the whole
// filename shows, which means a narrow window cannot hold every tab. Go
// decides which tabs render: the strip shows a window of full-width tabs
// with ‹ › arrows to move it, instead of squeezing tabs behind the strip
// tools. Widths mirror components/editor.css; tabMonoPx stays conservative
// (the painted 9.5 px mono advance measures ~5.7 px) so a visible tab can
// never reach its neighbour.
const (
	tabPadLeftPx   = 9  // .tab-main left padding
	tabPadRightPx  = 25 // .tab-main right padding reserves the close button
	tabIconPx      = 16 // .ficon.sm
	tabGapPx       = 6  // .tab-main gap between children
	tabStatusPx    = 16 // .tab-status, always rendered
	tabDotPx       = 6  // .tab-dot, only on dirty tabs
	tabBorderPx    = 1  // .tab border-right
	tabMonoPx      = 6.0
	tabNavPx       = 26 // .tab-nav scroll arrow width
	tabNavMarginPx = 3  // .tab-nav right margin
	tabFillMinPx   = 4  // .tabs-fill minimum
	// diffTabEstimatePx holds the Working tree tab that leads the strip in
	// changes mode: padding, the 12 px icon, gaps, a 12-rune name, and the
	// border, with slack for wider faces.
	diffTabEstimatePx = 140
)

// tabWidthPx estimates one tab's strip width for a filename and its dirty
// mark. Children are the badge, the name, the status, and the dot on dirty
// tabs; the absolutely positioned close button lives in the right padding.
func (a *App) tabWidthPx(name string, dirty bool) int {
	width := a.zx(tabPadLeftPx) + a.zx(tabIconPx) + a.zx(tabGapPx) +
		int(float64(len([]rune(name)))*a.zf(tabMonoPx)) +
		a.zx(tabGapPx) + a.zx(tabStatusPx) + a.zx(tabPadRightPx) + a.zx(tabBorderPx)
	if dirty {
		width += a.zx(tabDotPx) + a.zx(tabGapPx)
	}

	return width
}

// tabStripBudget is the pixel budget for file tabs in one group's strip:
// the pane width minus the tool buttons, the fill slack, the scroll arrow
// already showing, and the Working tree tab in changes mode.
func (a *App) tabStripBudget(group int) int {
	budget := a.groupEditorWidth(group) - 2*a.zx(tabNavPx) - 2*a.zx(tabNavMarginPx) - a.zx(tabFillMinPx)
	if a.tabStart[group] > 0 {
		budget -= a.zx(tabNavPx) + a.zx(tabNavMarginPx)
	}

	if group == 0 && a.mode == "changes" {
		budget -= a.zx(diffTabEstimatePx)
	}

	return budget
}

// tabWindowEnd lays tabs out from start inside budget and reports the first
// index past the visible window. The start tab always renders, even past
// the budget, so the strip never goes blank.
func (a *App) tabWindowEnd(rows []tabRow, start, budget int, reserveNext bool) int {
	avail := budget
	if reserveNext {
		avail -= a.zx(tabNavPx) + a.zx(tabNavMarginPx)
	}

	end := start

	for end < len(rows) {
		width := a.tabWidthPx(rows[end].Name, rows[end].Dirty)
		if end > start && avail < width {
			break
		}

		avail -= width
		end++
	}

	if end == start && end < len(rows) {
		end = start + 1
	}

	return end
}

// windowTabRows renders one group's visible tab window. It returns the rows
// plus whether the ‹ and › arrows show. Rows carry their full-list index so
// tab actions keep working on a windowed strip.
func (a *App) windowTabRows(group int) ([]tabRow, bool, bool) {
	rows := a.tabRowsFor(group)
	if len(rows) == 0 {
		a.tabStart[group] = 0

		return nil, false, false
	}

	start := a.tabStart[group]
	if start < 0 {
		start = 0
	}

	if start >= len(rows) {
		start = len(rows) - 1
	}

	a.tabStart[group] = start

	budget := a.tabStripBudget(group)

	end := a.tabWindowEnd(rows, start, budget, false)
	if end < len(rows) {
		end = a.tabWindowEnd(rows, start, budget, true)
	}

	return rows[start:end], start > 0, end < len(rows)
}

// ensureTabVisible moves one group's window so its active tab renders. Tab
// selection, opens, and closes call it; arrow scrolling deliberately does
// not, so the window stays where the user put it.
func (a *App) ensureTabVisible(group int) {
	rows := a.tabRowsFor(group)
	if len(rows) == 0 {
		a.tabStart[group] = 0

		return
	}

	active := a.groupActive(group)
	if active < 0 {
		active = 0
	}

	if active >= len(rows) {
		active = len(rows) - 1
	}

	start := a.tabStart[group]
	if start < 0 {
		start = 0
	}

	if start >= len(rows) {
		start = len(rows) - 1
	}

	if start > active {
		start = active
	}

	// The budget follows the live offset, so write each step back before
	// measuring: the final render measures the same budget from the same
	// start and cannot push the active tab back out.
	a.tabStart[group] = start

	for a.tabStart[group] < active {
		start := a.tabStart[group]
		budget := a.tabStripBudget(group)

		end := a.tabWindowEnd(rows, start, budget, false)
		if end < len(rows) {
			end = a.tabWindowEnd(rows, start, budget, true)
		}

		if active < end {
			break
		}

		a.tabStart[group] = start + 1
	}
}

// scrollTabs moves one group's tab window by delta tabs and repaints.
func (a *App) scrollTabs(group, delta int) {
	start := a.tabStart[group] + delta
	if start < 0 {
		start = 0
	}

	if last := len(a.groupTabs(group)) - 1; start > last {
		start = last
	}

	if start < 0 {
		start = 0
	}

	a.tabStart[group] = start
	a.refresh()
}
