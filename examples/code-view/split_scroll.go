package main

// split_scroll.go implements independent vertical scrolling for split editor groups (Group 0 and Group 1).
// Each editor group maintains its own vertical scroll offset (a.scrollGroup[0] and a.scrollGroup[1]),
// wheel accumulator, and clamping bounds.

// windowSliceForGroup returns the row range the fixed body shows and the track offset for group (0 or 1).
func (a *App) windowSliceForGroup(group, total int) (int, int, int) {
	bodyH := a.diffBodyHeight()
	if bodyH < rowH {
		bodyH = rowH
	}

	offset := a.scrollY
	if a.split && (group == 0 || group == 1) {
		offset = a.scrollGroup[group]
	}

	first := offset / rowH
	start := first
	if start > total {
		start = total
	}

	end := start + bodyH/rowH + 1
	if end > total {
		end = total
	}

	return start, end, 0
}

// maxGroupScroll returns the maximum vertical scroll offset in pixels for group (0 or 1).
func (a *App) maxGroupScroll(group int) int {
	doc := a.groupDoc(group)
	rows := a.paneRowCount(group, doc)
	maxY := rows*rowH - a.bodyHeight()
	if maxY < 0 {
		return 0
	}

	return maxY
}

// clampGroupScroll clamps y to the valid vertical scroll range [0, maxGroupScroll] for group.
func (a *App) clampGroupScroll(group, y int) int {
	if y < 0 {
		return 0
	}

	maxY := a.maxGroupScroll(group)
	if y > maxY {
		return maxY
	}

	return y
}

// scrollGroupBy shifts group's vertical scroll offset by delta pixels, clamped to bounds.
func (a *App) scrollGroupBy(group, delta int) {
	if group != 0 && group != 1 {
		return
	}

	current := a.scrollGroup[group]
	if !a.split && group == 0 {
		current = a.scrollY
	}

	a.scrollGroupTo(group, current+delta)
}

// scrollGroupTo sets group's vertical scroll offset to target pixels, clamped to bounds.
func (a *App) scrollGroupTo(group, target int) {
	if group != 0 && group != 1 {
		return
	}

	target = a.clampGroupScroll(group, target)
	a.scrollGroup[group] = target

	if !a.split && group == 0 {
		a.scrollY = target
		a.scrollTarget = target
		a.scrollPending = true
		if a.page != nil {
			a.page.ScrollTo(0, target)
		}
	} else if a.split && group == 0 {
		a.scrollY = target
	}

	a.refresh()
}

// scrollFocusedGroup moves the focused editor group by delta pixels.
// In changes mode or single-pane explorer, it scrolls the main view.
func (a *App) scrollFocusedGroup(delta int) {
	if a.mode == "changes" || !a.split {
		a.scrollTo(a.scrollY + delta)

		return
	}

	group := a.focusGroup
	if group != 1 {
		group = 0
	}

	a.scrollGroupBy(group, delta)
}

// scrollFocusedGroupTo sets the scroll offset of the focused editor group.
func (a *App) scrollFocusedGroupTo(target int) {
	if a.mode == "changes" || !a.split {
		a.scrollTo(target)

		return
	}

	group := a.focusGroup
	if group != 1 {
		group = 0
	}

	a.scrollGroupTo(group, target)
}

// scrollFocusedGroupToEnd scrolls the focused editor group to the bottom.
func (a *App) scrollFocusedGroupToEnd() {
	if a.mode == "changes" || !a.split {
		a.scrollTo(a.contentHeight())

		return
	}

	group := a.focusGroup
	if group != 1 {
		group = 0
	}

	a.scrollGroupTo(group, a.maxGroupScroll(group))
}

// scrollEditorGroupByWheel applies a vertical mouse wheel delta (-dy) to editor group (0 or 1).
func (a *App) scrollEditorGroupByWheel(group int, dy float64) {
	if group != 0 && group != 1 {
		return
	}

	a.splitWheelAccum[group] += -dy * float64(scrollStep)
	step := int(a.splitWheelAccum[group])
	a.splitWheelAccum[group] -= float64(step)

	if step != 0 {
		a.scrollGroupBy(group, step)
	}
}

// isOverPane reports whether (x, y) in viewport coordinates lies over editor pane (0 or 1).
func (a *App) isOverPane(group int, x, y float64) bool {
	id := "editor-window"
	if group == 1 {
		id = "editor-window-2"
	}

	if box, ok := findBoxID(a.page.Boxes(), id); ok {
		return x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H
	}

	if group == 1 {
		if !a.split {
			return false
		}
		left, width := a.secondGeometry()
		top := float64(chromeTop)
		bottom := float64(a.viewportH - chromeBottom)

		return x >= float64(left) && x <= float64(left+width) && y >= top && y <= bottom
	}

	left, width := a.paneGeometry()
	top := float64(chromeTop)
	bottom := float64(a.viewportH - chromeBottom)

	return x >= float64(left) && x <= float64(left+width) && y >= top && y <= bottom
}
