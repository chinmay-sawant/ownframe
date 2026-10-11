package telegram

// composeH is the composer height and sheetH the attachment sheet height,
// both in CSS pixels. The thread reserves them, and Pin keeps the bars at
// the viewport edges while the messages scroll. The top bar height lives in
// components/thread.css.
const (
	composeH = 63
	sheetH   = 150
	// keyboardMin is the bottom inset above which the phone keyboard is up.
	keyboardMin = 100
)

// Pin positions the pinned tabs and thread bars for a scroll offset and a viewport
// height. The page's scroll-window callback runs it before each redraw, so
// the window's own translation lands them back at the viewport edges. It
// reports whether a scroll redraw is needed: never on the phone, where the
// replay draws the pinned z-layer at the viewport and one settle redraw
// bakes the new positions.
func (a *App) Pin(offsetY, viewH int) bool {
	// A phone list starts at the top when its viewport height changes.
	if a.view.Phone && a.view.Active == "" && a.viewportH != 0 && a.viewportH != viewH {
		a.page.ScrollTo(0, 0)
	}
	a.viewportH = viewH
	// The engine places an absolute box against its parent's content box,
	// which starts at InsetTop. The bar must land at offsetY + InsetTop in
	// document space so the window's translation puts it at the top edge.
	top := offsetY
	height := composeH
	if a.view.Active == "" {
		height = 60
	} else if a.view.AttachOpen {
		height += sheetH
	}

	compose := offsetY + viewH - a.view.InsetBottom - height - a.view.InsetTop
	// The status strip sits one inset above the top bar in document space,
	// because the engine anchors absolute boxes inside the app's padding.
	pad := offsetY - a.view.InsetTop
	changed := top != a.view.BarTop || compose != a.view.BottomTop || pad != a.view.PadTop

	a.view.BarTop, a.view.BottomTop, a.view.PadTop = top, compose, pad

	if a.view.Phone {
		return false
	}

	return changed
}
