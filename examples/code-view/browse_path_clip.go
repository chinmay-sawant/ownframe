package main

// The browse-head constants mirror components/chrome.css: the 560 px panel
// keeps 10 px side padding and a 1 px border on each side, the Up button
// carries its label plus 10 px side padding (reserved with slack for wider
// faces), the row keeps a 6 px gap, and the path keeps 4 px side padding.
// The panel shrinks below 600 px windows through max-width: calc(100vw -
// 40px). browseCellW covers the 10 px mono face with margin over the
// measured advance so a clipped name cannot reach the panel edge even when
// the face resolves a little wide.
const (
	browsePanelW      = 560
	browsePanelChrome = 22
	browseUpW         = 44
	browseHeadGap     = 6
	browsePathPad     = 8
	browseGutter      = 40
	browseCellW       = 6.5
)

// clipBrowsePath keeps the folder browser's displayed path on one line
// within the browse panel and marks the cut with an ellipsis. The engine
// paints a long label past its measured box (see clipCmdbarText), so the
// CSS ellipsis trio alone cannot hold a long folder name and the name is
// clipped in Go. The full path stays in the row's title attribute, so
// nothing is lost. Short names pass through unchanged.
func clipBrowsePath(windowWidth int, name string) string {
	panel := browsePanelW
	if max := windowWidth - browseGutter; max < panel {
		panel = max
	}

	budget := int(float64(panel-browsePanelChrome-browseUpW-browseHeadGap-browsePathPad) / browseCellW)
	if budget < 1 {
		budget = 1
	}

	runes := []rune(name)
	if len(runes) <= budget {
		return name
	}

	if budget == 1 {
		return "…"
	}

	return string(runes[:budget-1]) + "…"
}
