package main

import "context"

// inventoryWheelStep is the CSS pixels scrolled per wheel delta unit,
// matching the web preview viewport step.
const inventoryWheelStep = 48.0

// handleWheel scrolls the open search palette under the cursor, else the
// inner body viewport under the cursor. It consumes the event so the web
// preview does not pan the page itself (handled=true skips scrollPage);
// the static top bar never consumes.
func (a *app) handleWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 {
		return false, nil
	}
	if a.searchOpen {
		if consumed, changed := a.scrollSearchPanel(x, y, dy); consumed {
			if !changed {
				return true, nil
			}
			return true, a.refresh(ctx)
		}
	}
	if y < float64(inventoryTopbarHeight) {
		return false, nil
	}
	a.bodyFrac += -dy * inventoryWheelStep
	move := int(a.bodyFrac)
	a.bodyFrac -= float64(move)
	viewportH, contentH := a.bodyMetrics()
	next := clampBodyOffset(a.bodyOffset+move, viewportH, contentH)
	if next != a.bodyOffset {
		a.bodyOffset = next
		return true, a.refresh(ctx)
	}
	return true, nil
}

// scrollSearchPanel moves the palette window one row per wheel notch when
// the pointer is over the open search panel. It reports whether the event
// was consumed and whether the offset moved; the caller consumes the event
// either way so the page never pans under the panel.
func (a *app) scrollSearchPanel(x, y, dy float64) (consumed, changed bool) {
	over := false
	for _, box := range a.page.Boxes() {
		if box.ID != "inv-search-panel" {
			continue
		}
		if x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
			over = true
		}
		break
	}
	if !over {
		return false, false
	}
	step := 0
	if dy < 0 {
		step = 1
	} else if dy > 0 {
		step = -1
	}
	if step == 0 {
		return true, false
	}
	_, total := a.searchWindow()
	maxStart := total - inventorySearchVisible
	if maxStart < 0 {
		maxStart = 0
	}
	next := min(max(a.searchOffset+step, 0), maxStart)
	if next == a.searchOffset {
		return true, false
	}
	a.searchOffset = next
	return true, true
}

// bodyMetrics measures the inner viewport and content heights in CSS pixels.
// It falls back to the page height below the top bar until boxes are laid out.
func (a *app) bodyMetrics() (viewportH, contentH int) {
	_, height := a.page.Size()
	viewportH = max(height-inventoryTopbarHeight, 0)
	contentH = viewportH
	for _, box := range a.page.Boxes() {
		switch box.ID {
		case inventoryViewportID:
			viewportH = int(box.H)
		case inventoryContentID:
			contentH = int(box.H)
		}
	}
	return viewportH, contentH
}

// clampBodyOffset pins offset into [0, contentH-viewportH].
func clampBodyOffset(offset, viewportH, contentH int) int {
	maximum := contentH - viewportH
	if maximum < 0 {
		return 0
	}
	return min(max(offset, 0), maximum)
}
