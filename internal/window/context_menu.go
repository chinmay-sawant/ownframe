package window

import (
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// menuState is the shell context menu.
type menuState struct {
	open  bool
	x, y  int
	items []host.MenuItem
}

// openMenu asks the page for its rows and shows the menu at a point. A screen
// that names the element under the pointer gets first refusal; a screen with
// no rows for that element falls back to the state-wide menu. A page without
// a context menu keeps the right click.
func (s *shell) openMenu(x, y int) {
	items := s.menuRows(x, y)
	if len(items) == 0 {
		return
	}

	if presenter, ok := s.app.(host.MenuPresenter); ok && presenter.PresentMenu(x, y, items) {
		s.closeMenu()

		return
	}

	s.menu = menuState{open: true, x: x, y: y, items: items}
}

// menuRows asks the screen for the rows at a window point. The hit-tested box
// id goes to ContextMenuFor when the screen implements it. An empty answer
// falls back to ContextMenu, so an element the app does not customize keeps
// the built-in editing rows.
func (s *shell) menuRows(x, y int) []host.MenuItem {
	if s.app == nil {
		return nil
	}

	id := ""
	px, py := s.contentPointAt(x, y)
	if menu, ok := s.app.(interface {
		ContextMenuAt(x, y float64) []host.MenuItem
	}); ok {
		return menu.ContextMenuAt(px, py)
	}
	if box, ok := s.boxAt(px, py); ok {
		id = box.ID
	}

	if menu, ok := s.app.(host.ContextMenuFor); ok {
		if items := menu.ContextMenuFor(id); len(items) > 0 {
			return items
		}
	}

	if menu, ok := s.app.(host.ContextMenu); ok {
		return menu.ContextMenu()
	}

	return nil
}

// closeMenu hides the menu.
func (s *shell) closeMenu() {
	s.menu.open = false
	s.menu.items = nil
}

// menuClaim gives an open menu first refusal on a fresh left press. The
// pointer runs the custom-chrome window drag before the menu, so without
// this a press on a menu row inside a draggable strip starts a drag: the
// row never acts, and the release lands a stray page click while the menu
// stays open. Right presses still flow through so the menu can reopen at a
// new point. It reports whether the press was consumed.
func (s *shell) menuClaim(x, y int, left, down bool) (bool, error) {
	if !s.menu.open || !left {
		return false, nil
	}

	handled, err := s.menuClick(x, y)
	if handled {
		s.mouseDown = down
	}

	return handled, err
}

// menuPointer handles right and left presses over the menu. It reports
// whether the event was consumed, so the page does not also see the click.
func (s *shell) menuPointer(x, y int) (bool, error) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		s.openMenu(x, y)

		return true, nil
	}

	if !s.menu.open || !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return false, nil
	}

	return s.menuClick(x, y)
}

// menuClick acts on a left click at a point and closes the menu. It reports
// whether the menu consumed the click.
func (s *shell) menuClick(x, y int) (bool, error) {
	if !s.menu.open {
		return false, nil
	}

	row := s.menuIndex(x, y)
	if row >= 0 && s.menu.items[row].Enabled {
		err := s.menuAction(s.menu.items[row].ID)
		s.closeMenu()

		return true, err
	}

	s.closeMenu()

	return true, nil
}

const (
	menuRowH = 22
	menuPadX = 10
	menuPadY = 4
	menuMinW = 96
	// menuSides pads both the left and the right of a row label.
	menuSides = 2
)

// menuRect returns the menu box for its rows, pulled inside the screen.
func (s *shell) menuRect() (x, y, w, h float64) {
	for _, item := range s.menu.items {
		tw, _ := menuTextWidth(item.Label)
		if tw+menuSides*menuPadX > w {
			w = tw + menuSides*menuPadX
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
		text, err := clipboard.Read()
		if err != nil {
			return nil
		}

		return s.clip().Paste(s.ctx, text)
	case "select-all":
		return s.clip().SelectAll(s.ctx)
	case "undo":
		return s.clip().Undo(s.ctx)
	case "redo":
		return s.clip().Redo(s.ctx)
	}

	if app, ok := s.app.(host.MenuActioner); ok {
		return app.MenuAction(s.ctx, id)
	}

	return nil
}

const (
	menuFillHex     = 0xfffffff2
	menuEdgeHex     = 0xbdbdbdff
	menuInkHex      = 0x1c1915ff
	menuDisabledHex = 0x9c9c9cff
)

var (
	menuFill     = devRGBA(menuFillHex)
	menuEdge     = devRGBA(menuEdgeHex)
	menuInk      = devRGBA(menuInkHex)
	menuDisabled = devRGBA(menuDisabledHex)
)

// menuTextWidth measures one row label with the badge face.
func menuTextWidth(label string) (float64, float64) {
	return text.Measure(label, badgeFace, 0)
}

// drawMenu paints the context menu over the frame. The menu is chrome: it
// never enters the page picture and no redraw sees it.
func (s *shell) drawMenu(screen *ebiten.Image) {
	if !s.menu.open || len(s.menu.items) == 0 {
		return
	}

	x, y, w, h := s.menuRect()
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), menuFill, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, menuEdge, false)

	for i, item := range s.menu.items {
		ink := menuInk
		if !item.Enabled {
			ink = menuDisabled
		}

		var op text.DrawOptions
		op.GeoM.Translate(x+menuPadX, y+float64(i)*menuRowH+menuPadY)
		op.ColorScale.ScaleWithColor(ink)
		text.Draw(screen, item.Label, badgeFace, &op)
	}

	s.frameNote(frameRect(x, y, w, h))
}
