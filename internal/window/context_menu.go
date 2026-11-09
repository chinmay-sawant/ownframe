package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/chinmay-sawant/ownframe/internal/host"
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
