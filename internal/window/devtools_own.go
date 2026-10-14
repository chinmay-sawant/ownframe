package window

import "github.com/hajimehoshi/ebiten/v2"

// devOwns reports whether the overlay owns one key under these modifiers.
func (s *shell) devOwns(key ebiten.Key, mods modifiers) bool {
	insp := s.devInspector()
	if insp == nil {
		return false
	}

	// Keep the overlay available only for screens that opted in through
	// Config.DevTools or SetDevTools. A plain screen never sees F12.
	available := s.dev.available || s.dev.on || insp.DevTools()
	if !available {
		return false
	}
	s.dev.available = true

	switch key {
	case ebiten.KeyF12:
		return true
	case ebiten.KeyI:
		return mods.Control && mods.Shift
	case ebiten.KeyO:
		return s.dev.on && !mods.command()
	}

	return false
}

func devKeyIndex(key ebiten.Key) int {
	for i, k := range devKeys {
		if k == key {
			return i
		}
	}

	return -1
}

// devToggle flips the overlay through the page's own flag.
func (s *shell) devToggle() error {
	insp := s.devInspector()
	if insp == nil {
		return nil
	}

	on := !s.dev.on
	insp.SetDevTools(on)
	s.dev.available = true
	s.dev.on = on

	if !on {
		return s.devClose()
	}

	return nil
}

// devEats reports whether keyEvents must leave one key to the overlay.
func (s *shell) devEats(key ebiten.Key) bool {
	i := devKeyIndex(key)

	return i >= 0 && s.dev.eaten[i]
}

// pageKeyStep is the keyEvents step for one key. A key the overlay owns
// never reaches the page watch, so its press and release stay invisible.
func (s *shell) pageKeyStep(key ebiten.Key, pressed bool) (bool, bool) {
	if s.devEats(key) {
		return false, false
	}

	return s.watched.step(key, pressed)
}
