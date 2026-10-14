package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const badgeLabel = "bitmap fallback"

var (
	badgeFace = text.NewGoXFace(basicfont.Face7x13)
	badgeInk  = color.RGBA{R: 0x1c, G: 0x19, B: 0x15, A: 0xe6}
)

// badgeRect returns the top-right box that holds label on a screenW-wide
// frame with no dock.
func badgeRect(screenW float64, label string) (x, y, width, height float64) {
	return badgeRectInset(screenW, label, 0)
}

// badgeRectInset is badgeRect with the right edge inset pixels from the
// screen edge, so the devtools dock leaves the badge visible.
func badgeRectInset(screenW float64, label string, inset float64) (x, y, width, height float64) {
	labelW, labelH := text.Measure(label, badgeFace, 0)

	width = labelW + 8
	height = labelH + 8
	x = screenW - width - 8 - inset
	if x < 8 {
		x = 8
	}

	y = 8

	return x, y, width, height
}

// drawBadge paints the bitmap fallback notice on top of the frame. The
// devtools dock pushes it left.
func (s *shell) drawBadge(screen *ebiten.Image) {
	inset := 0.0
	if s.dev.on {
		inset = s.dev.panel.W
	}

	x, y, width, height := badgeRectInset(float64(s.screenW), badgeLabel, inset)

	vector.FillRect(
		screen,
		float32(x), float32(y), float32(width), float32(height),
		badgeInk, false,
	)

	var op text.DrawOptions
	op.GeoM.Translate(x+4, y+4)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, badgeLabel, badgeFace, &op)
	s.frameNote(frameRect(x, y, width, height))
}
