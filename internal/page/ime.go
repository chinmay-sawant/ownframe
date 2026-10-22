package page

import (
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Page is an IME target for the window.
var _ host.IME = (*Page)(nil)

// IMEContext returns the focused text control's box, the caret as a byte
// offset into the value, and the value around the caret. ok is false when
// no text control is focused.
func (p *Page) IMEContext() (layout.Box, int, string, string, bool) {
	c, ok := p.typingTarget()
	if !ok {
		return layout.Box{}, 0, "", "", false
	}

	runes := []rune(c.Value)
	caret := clampPos(p.formState().caret, len(runes))
	before := string(runes[:caret])

	// The window reports the caret to the platform in screen coordinates,
	// so take the scroll offset out of the box. Ebiten pans the canvas when
	// the keyboard would cover this rectangle.
	box := p.boxByID(c.ID)
	box.X -= float64(p.scroll.Windowing.X)
	box.Y -= float64(p.scroll.Windowing.Y)

	return box, len(before), before, string(runes[caret:]), true
}
