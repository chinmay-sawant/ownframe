package page

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
	"strings"
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

var _ host.IMECaretAnchor = (*Page)(nil)

// IMEAnchor returns the focused glyph caret in CSS pixels, adjusted for page
// scrolling. False means the current display cannot measure the caret.
func (p *Page) IMEAnchor() (layout.Box, bool) {
	c, ok := p.typingTarget()
	if !ok || p.draw.Display == nil {
		return layout.Box{}, false
	}
	box, ok := caretBox(p.draw.Display, c.ID)
	if !ok {
		return layout.Box{}, false
	}
	text, offset := p.imeCaretText(c)
	sample, x, ok := textrun.Caret(p.draw.Display, box, text, offset)
	if !ok {
		sample, x = emptyCaret(p.draw.Display, box)
	}
	pt := p.draw.Display.PixelPerPoint
	return layout.Box{X: x*pt - float64(p.scroll.Windowing.X),
		Y: (sample.Y-sample.Size)*pt - float64(p.scroll.Windowing.Y),
		W: p.draw.Display.PointsPerPixel * pt, H: sample.Size * pt}, true
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

// SetIMEComposition paints temporary IME text without storing or notifying
// the field value. start and end select bytes within text.
func (p *Page) SetIMEComposition(ctx context.Context, text string, start, end int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.formState() == nil {
		return nil
	}
	if text == "" {
		if p.formState().preedit == "" {
			return nil
		}
		p.formState().preeditID, p.formState().preedit = "", ""
		p.formState().preeditStart, p.formState().preeditEnd = 0, 0
		return p.Redraw(ctx)
	}
	if _, ok := p.typingTarget(); !ok {
		return nil
	}
	if p.formState().preeditID == p.formState().focusID && p.formState().preedit == text &&
		p.formState().preeditStart == start && p.formState().preeditEnd == end {
		return nil
	}
	p.formState().preeditID, p.formState().preedit = p.formState().focusID, text
	p.formState().preeditStart, p.formState().preeditEnd = start, end
	return p.Redraw(ctx)
}

// IMEReplace replaces the focused value's bytes [start, end) with insert
// and puts the caret at the byte offset caret within insert. It draws when
// the value changed.
func (p *Page) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	runes := []rune(c.Value)
	startRune := runeAtByte(c.Value, start)
	endRune := runeAtByte(c.Value, end)
	if startRune == endRune && insert == "" {
		return nil
	}

	ins := []rune(insert)
	pos := startRune + clampPos(runeAtByte(insert, caret), len(ins))

	return p.editField(ctx, nil, func(c Control, _, _ int) (Control, int, int, bool) {
		out := make([]rune, 0, len(runes)-(endRune-startRune)+len(ins))
		out = append(out, runes[:startRune]...)
		out = append(out, ins...)
		out = append(out, runes[endRune:]...)

		return setRunes(c, out), pos, pos, true
	})
}
