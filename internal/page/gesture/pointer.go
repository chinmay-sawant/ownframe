package gesture

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// SameHit reports whether the pointer is still on the same element.
func SameHit(current, next string) bool { return current == next }

// RunPress reports whether the press handler should see this control.
// A disabled control is skipped. A hit that is not a control still runs.
func RunPress(hasHandler, hasControl, disabled bool) bool {
	return hasHandler && (!hasControl || !disabled)
}

// CanPaintHover reports whether a hover handler can paint into the display.
func CanPaintHover(hasHandler, hasDisplay bool) bool {
	return hasHandler && hasDisplay
}

// SkipHoverRedraw reports whether the hover change should not relayout.
// A handler that painted, or a page without :hover, skips the redraw.
func SkipHoverRedraw(handled, pseudo bool) bool {
	return handled || !pseudo
}

// ClaimsWord reports whether a long press selects the word in field.
// ok is false when the point is not a control.
func ClaimsWord(field form.Control, ok bool) bool {
	return ok && !field.Disabled && form.CanEdit(field)
}

// Swipe returns dx, dy when a swipe handler is registered.
func Swipe(hasHandler bool, dx, dy float64) (float64, float64, bool) {
	if !hasHandler {
		return 0, 0, false
	}

	return dx, dy, true
}
