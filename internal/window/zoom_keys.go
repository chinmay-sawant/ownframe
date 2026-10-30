package window

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// zoomStep is one keyboard zoom press or one wheel notch.
const zoomStep = 1.1

// zoomModifier reports that Control or Command is down, and not AltGr.
func zoomModifier() bool {
	command := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	altGr := ebiten.IsKeyPressed(ebiten.KeyAlt) && !ebiten.IsKeyPressed(ebiten.KeyMeta)

	return command && !altGr
}

// zoomKey applies one Control+=, Control+-, or Control+0 press. It reports
// whether the window consumed the key; the plain keys stay with the page.
// A page that refuses key zoom keeps the chords, so the page can run its
// own layout zoom (the code view does); the painted scale stays untouched.
func (s *shell) zoomKey(key ebiten.Key, down bool, mods modifiers) bool {
	if s.viewLocked() || !mods.command() || mods.Alt {
		return false
	}
	if !s.keyZoomAllowed() {
		return false
	}

	switch key {
	case ebiten.KeyEqual, ebiten.KeyKPAdd:
		if down {
			s.zoomBy(zoomStep)
		}
	case ebiten.KeyMinus, ebiten.KeyKPSubtract:
		if down {
			s.zoomBy(1 / zoomStep)
		}
	case ebiten.Key0, ebiten.KeyKP0:
		if down {
			s.zoomReset()
		}
	default:
		return false
	}

	return true
}

// wheelZoomFactor returns the factor for one wheel delta. Half a notch on a
// trackpad moves the zoom half as far.
func wheelZoomFactor(wheelY float64) float64 {
	if wheelY == 0 {
		return 1
	}

	return math.Pow(zoomStep, wheelY)
}

// zoomWheel applies Control+wheel to the page zoom and reports whether the
// wheel went to the zoom instead of the scroll.
// wheelMayZoom reports that this notch is allowed to scale the page. The
// modifier check stays in zoomWheel because it reads the live key state.
func (s *shell) wheelMayZoom(wheelY float64) bool {
	return !s.viewLocked() && wheelY != 0 && s.wheelZoomAllowed()
}

func (s *shell) zoomWheel(wheelY float64) bool {
	if !s.wheelMayZoom(wheelY) || !zoomModifier() {
		return false
	}

	s.zoomBy(wheelZoomFactor(wheelY))

	return true
}
