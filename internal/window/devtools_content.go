package window

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// devLineGap is the leading between panel lines.
const devLineGap = 2.0

// devLineHeight is the badge face line height plus a little leading.
func devLineHeight() float64 {
	return devFaceHeight() + devLineGap
}

// devTextAt draws one run at an exact panel point.
func devTextAt(screen *ebiten.Image, label string, x, y float64, ink color.RGBA) {
	var op text.DrawOptions
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(screen, label, badgeFace, &op)
}

// drawDevContent paints the active tab's rows, clipped to the content rect,
// with a hover band and a blue band on the chosen operation.
func (s *shell) drawDevContent(screen *ebiten.Image) {
	c := s.dev.content
	if c.W <= 0 || c.H <= 0 {
		return
	}

	clip, ok := screen.SubImage(image.Rect(int(c.X), int(c.Y), int(c.X+c.W), int(c.Y+c.H))).(*ebiten.Image)
	if !ok {
		return
	}
	lineH := devLineHeight()
	y := c.Y - s.dev.scroll

	for _, row := range s.dev.rows {
		if y+lineH > c.Y && y < c.Y+c.H {
			s.devRowBack(screen, row, y, lineH)
			s.devDrawLine(clip, row.line, c.X, y, row.w)
		}

		y += lineH
	}
}

// devRowHot reports the pointer inside one row band.
func (s *shell) devRowHot(y, lineH float64) bool {
	cy := float64(s.cursorY)

	return cy >= y && cy < y+lineH &&
		s.cursorX >= int(s.dev.panel.X) && s.cursorX <= int(s.dev.panel.X+s.dev.panel.W)
}
