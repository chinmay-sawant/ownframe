package window

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// windowKey lets the window act on a key before the page handler sees it.
// It reports whether the key was consumed, so the handler stays unaware of
// the keys the window owns.
func (s *shell) windowKey(key ebiten.Key, down bool, mods modifiers) (bool, error) {
	switch key {
	case ebiten.KeyTab:
		return s.tabKey(down, mods)
	case ebiten.KeyEscape:
		if down {
			wasOpen := s.menu.open
			if err := s.escape(); err != nil {
				return false, err
			}

			return wasOpen, nil
		}
	case ebiten.KeyF11:
		return s.f11(down), nil
	case ebiten.KeyEqual, ebiten.KeyKPAdd,
		ebiten.KeyMinus, ebiten.KeyKPSubtract,
		ebiten.Key0, ebiten.KeyKP0:
		return s.zoomKey(key, down, mods), nil
	}

	return false, nil
}

// tabKey moves focus with Tab and Shift+Tab. The key is consumed only when a
// field took the focus, so a page without fields keeps the key.
func (s *shell) tabKey(down bool, mods modifiers) (bool, error) {
	if !down {
		taken := s.tabEaten
		s.tabEaten = false

		return taken, nil
	}

	focuser, ok := s.app.(host.Focuser)
	if !ok {
		return false, nil
	}

	var err error
	if mods.Shift {
		err = focuser.FocusPrev(s.ctx)
	} else {
		err = focuser.FocusNext(s.ctx)
	}

	if err != nil {
		return false, err
	}

	if focuser.FocusID() == "" {
		return false, nil
	}

	s.tabEaten = true

	return true, nil
}

// escape closes the context menu first, then clears keyboard focus.
func (s *shell) escape() error {
	if s.menu.open {
		s.closeMenu()

		return nil
	}

	focuser, ok := s.app.(host.Focuser)
	if !ok || focuser.FocusID() == "" {
		return nil
	}
	if err := s.imeCancelSession(); err != nil {
		return err
	}

	return focuser.Focus(s.ctx, "")
}
