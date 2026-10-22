package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const badgeLabel = "bitmap fallback"

// badgePad is the badge box pad and screen margin.
const badgePad = 8.0

// badgeInkHex packs badgeInk as 0xRRGGBBAA.
const badgeInkHex = 0x1c1915e6

var (
	badgeFace = text.NewGoXFace(basicfont.Face7x13)
	badgeInk  = devRGBA(badgeInkHex)
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

	width = labelW + badgePad
	height = labelH + badgePad
	x = screenW - width - badgePad - inset
	if x < badgePad {
		x = badgePad
	}

	y = badgePad

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
	op.GeoM.Translate(x+devLabelInset, y+devLabelInset)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, badgeLabel, badgeFace, &op)
	s.frameNote(frameRect(x, y, width, height))
}
