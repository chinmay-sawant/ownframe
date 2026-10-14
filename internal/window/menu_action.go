package window

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/clipboard"
)

const (
	menuRowH = 22
	menuPadX = 10
	menuPadY = 4
	menuMinW = 96
)

// menuRect returns the menu box for its rows, pulled inside the screen.
func (s *shell) menuRect() (x, y, w, h float64) {
	for _, item := range s.menu.items {
		tw, _ := menuTextWidth(item.Label)
		if tw+2*menuPadX > w {
			w = tw + 2*menuPadX
		}
	}

	if w < menuMinW {
		w = menuMinW
	}

	h = float64(len(s.menu.items)) * menuRowH
	x = float64(s.menu.x)
	y = float64(s.menu.y)
	viewW, viewH := s.viewSize()

	if viewW > 0 && x+w > float64(viewW) {
		x = float64(viewW) - w
	}

	if viewH > 0 && y+h > float64(viewH) {
		y = float64(viewH) - h
	}

	if x < 0 {
		x = 0
	}

	if y < 0 {
		y = 0
	}

	return x, y, w, h
}

// menuIndex returns the row at a point, or -1 outside the rows.
func (s *shell) menuIndex(x, y int) int {
	mx, my, mw, _ := s.menuRect()
	if float64(x) < mx || float64(x) > mx+mw || float64(y) < my {
		return -1
	}

	row := int((float64(y) - my) / menuRowH)
	if row < 0 || row >= len(s.menu.items) {
		return -1
	}

	return row
}

// menuAction runs one context menu row through the existing screen methods.
func (s *shell) menuAction(id string) error {
	switch id {
	case "cut":
		return s.copy(true)
	case "copy":
		return s.copy(false)
	case "paste":
		text := strings.ReplaceAll(clipboard.Read(), "\r", "")
		text = strings.ReplaceAll(text, "\n", "")

		return s.app.Paste(s.ctx, text)
	case "select-all":
		return s.app.SelectAll(s.ctx)
	case "undo":
		return s.app.Undo(s.ctx)
	case "redo":
		return s.app.Redo(s.ctx)
	}

	return nil
}
