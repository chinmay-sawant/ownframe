package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// devRowBack paints the row band: blue on the chosen operation, grey under
// the pointer on any actionable row.
func (s *shell) devRowBack(screen *ebiten.Image, row devRow, y, lineH float64) {
	picked := row.act == devActOp && s.dev.haveOp && row.arg == s.dev.opPick
	if !picked && !(row.act != devActNone && s.devRowHot(y, lineH)) {
		return
	}

	ink := devRowHover
	if picked {
		ink = devRowPick
	}

	vector.FillRect(screen, float32(s.dev.panel.X), float32(y), float32(s.dev.panel.W), float32(lineH), ink, false)
}

// devDrawLine draws one row's spans in order, swatches included. width is the
// line's content width, so a right-aligned span starts at the right edge.
func (s *shell) devDrawLine(screen *ebiten.Image, line devLine, x, y, width float64) {
	start := x

	for _, span := range line.spans {
		if span.swatch > 0 {
			top := y + (devLineHeight()-span.swatch)/devCenterDiv
			vector.FillRect(screen, float32(x), float32(top), float32(span.swatch), float32(span.swatch), span.ink, false)
			x += span.swatch + devSwatchGap

			continue
		}

		if span.right {
			x = start + width - span.w
		}

		devTextAt(screen, span.text, x, y, span.ink)

		if span.w > 0 {
			x += span.w

			continue
		}

		w, _ := text.Measure(span.text, badgeFace, 0)
		x += w
	}
}

// drawDevScrollbar paints a thumb when the rows overflow the viewport.
func (s *shell) drawDevScrollbar(screen *ebiten.Image) {
	c := s.dev.content
	total := float64(len(s.dev.rows)) * devLineHeight()
	if total <= c.H || c.H <= 0 {
		return
	}

	const thick = 4.0

	length := c.H * c.H / total
	if length < devScrollMin {
		length = devScrollMin
	}

	pos := c.Y + (c.H-length)*s.dev.scroll/(total-c.H)
	x := s.dev.panel.X + s.dev.panel.W - thick - devScrollMargin
	vector.FillRect(screen, float32(x), float32(pos), thick, float32(length), devBorder, false)
}
