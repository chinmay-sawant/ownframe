package window

import "github.com/chinmay-sawant/ownframe/internal/host"

// windowBounds returns the largest window the screen asks for, or -1 for no
// bound. A screen without host.MaxSizeProvider is unbounded, as it was
// before the cap became a window bound. The cap is window-only; the browser
// sink serves pictures and never reads it.
func windowBounds(app host.Screen) (int, int) {
	maxW, maxH := -1, -1

	sizer, ok := app.(host.MaxSizeProvider)
	if !ok {
		return maxW, maxH
	}

	maxW, maxH = sizer.MaxSize()
	if maxW <= 0 {
		maxW = -1
	}

	if maxH <= 0 {
		maxH = -1
	}

	return maxW, maxH
}
