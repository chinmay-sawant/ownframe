package main

import (
	"math"
	"unicode/utf8"
)

const (
	hscrollStep      = 40 // pixels per keyboard step
	hscrollWheelStep = 30 // pixels per wheel tick
)

// scrollHorizontal shifts horizontal scroll offset by dx pixels, clamped between
// 0 and the maximum line width in the active document.
func (a *App) scrollHorizontal(dx int) {
	if a.wrap {
		a.scrollX = 0

		return
	}

	a.setScrollX(a.scrollX + dx)
}

// ScrollHorizontal is the exported form of scrollHorizontal.
func (a *App) ScrollHorizontal(dx int) {
	a.scrollHorizontal(dx)
}

// ScrollX returns the current horizontal scroll offset in pixels.
func (a *App) ScrollX() int {
	if a.wrap {
		return 0
	}

	return a.scrollX
}

// SetScrollX sets the horizontal scroll offset in pixels, clamped to [0, maxLineWidth].
func (a *App) SetScrollX(x int) {
	a.setScrollX(x)
}

// setScrollX sets the horizontal scroll offset in pixels, clamped to [0, maxLineWidth].
func (a *App) setScrollX(x int) {
	if a.wrap {
		a.scrollX = 0

		return
	}

	maxW := a.maxLineWidth()
	if x < 0 {
		x = 0
	} else if x > maxW {
		x = maxW
	}

	a.scrollX = x
	a.refresh()
}

// maxLineWidth returns the maximum line width in pixels of the active document
// or diff rows.
func (a *App) maxLineWidth() int {
	if a.mode == "changes" {
		rows := a.currentDiffRows()
		maxRunes := 0

		for _, row := range rows {
			leftRunes := 0
			for _, seg := range row.Left {
				leftRunes += utf8.RuneCountInString(seg.Text)
			}

			if leftRunes > maxRunes {
				maxRunes = leftRunes
			}

			rightRunes := 0
			for _, seg := range row.Right {
				rightRunes += utf8.RuneCountInString(seg.Text)
			}

			if rightRunes > maxRunes {
				maxRunes = rightRunes
			}

			if row.Hunk != "" {
				if h := utf8.RuneCountInString(row.Hunk); h > maxRunes {
					maxRunes = h
				}
			}
		}

		return int(math.Ceil(float64(maxRunes) * diffCellW))
	}

	doc := a.activeDoc()
	if doc == nil {
		return 0
	}

	maxRunes := 0

	for _, line := range doc.Lines {
		if r := utf8.RuneCountInString(line); r > maxRunes {
			maxRunes = r
		}
	}

	return int(math.Ceil(float64(maxRunes) * codeCellW))
}

// MaxLineWidth returns the maximum line width in pixels of the active document.
func (a *App) MaxLineWidth() int {
	return a.maxLineWidth()
}

// isOverEditor reports whether point (x, y) is inside the editor pane.
func (a *App) isOverEditor(x, y float64) bool {
	for _, id := range []string{"code-body", "editor-window", "code-body-2", "editor-window-2"} {
		if box, ok := findBoxID(a.page.Boxes(), id); ok {
			if x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
				return true
			}
		}
	}

	// Fallback when layout boxes are not present (e.g. headless tests).
	w := float64(defaultWidth)
	if a.page != nil {
		if pw, _ := a.page.Size(); pw > 0 {
			w = float64(pw)
		}
	}

	left := float64(chromeLeft)

	return x >= left && x <= w
}
