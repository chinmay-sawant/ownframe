package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

var _ host.IMECaretAnchor = (*Page)(nil)

// IMEAnchor returns the focused glyph caret in CSS pixels, adjusted for page
// scrolling. False means the current display cannot measure the caret.
func (p *Page) IMEAnchor() (layout.Box, bool) {
	c, ok := p.typingTarget()
	if !ok || p.display == nil {
		return layout.Box{}, false
	}
	box, ok := caretBox(p.display, c.ID)
	if !ok {
		return layout.Box{}, false
	}
	text, offset := p.imeCaretText(c)
	sample, x, ok := textrun.Caret(p.display, box, text, offset)
	if !ok {
		sample, x = emptyCaret(p.display, box)
	}
	pt := p.display.PixelPerPoint
	return layout.Box{X: x*pt - float64(p.windowing.x),
		Y: (sample.Y-sample.Size)*pt - float64(p.windowing.y),
		W: p.display.PointsPerPixel * pt, H: sample.Size * pt}, true
}

func (p *Page) imeCaretText(c Control) (string, int) {
	kind := strings.ToLower(strings.TrimSpace(c.Type))
	if p.formState().preeditID != c.ID {
		return shownText(kind, c.Value), clampPos(p.formState().caret, runeLen(c.Value))
	}
	base := []rune(c.Value)
	start, end := p.formState().bounds(len(base))
	value := string(base[:start]) + p.formState().preedit + string(base[end:])
	offset := start + runeAtByte(p.formState().preedit, p.formState().preeditStart)
	return shownText(kind, value), offset
}
