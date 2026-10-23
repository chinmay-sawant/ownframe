package snip

import (
	"fmt"
	"strconv"
)

// selectFields reads x/y/w/h, clamps into the shot, and re-crops.
func (a *App) selectFields() {
	w, h := a.shotSize()
	a.sel = Clamp(Rect{
		X: num(a.page.Form().FormValue("x")),
		Y: num(a.page.Form().FormValue("y")),
		W: num(a.page.Form().FormValue("w")),
		H: num(a.page.Form().FormValue("h")),
	}, Rect{W: w, H: h})
	a.recrop()
	a.view.Status = fmt.Sprintf("crop %dx%d+%d+%d",
		a.sel.W, a.sel.H, a.sel.X, a.sel.Y)
}

// clear selects the full capture and drops the overlays.
func (a *App) clear() {
	w, h := a.shotSize()
	a.sel = Rect{W: w, H: h}
	a.anns = ClearAnnotations(a.anns)
	a.recrop()
	a.view.Status = "cleared"
}

// num parses tolerant ints; blanks and junk read as zero.
func num(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
