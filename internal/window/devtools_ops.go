package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

var (
	devFillInk   = color.RGBA{R: 0x1a, G: 0x56, B: 0xdb, A: 0xff}
	devStrokeInk = color.RGBA{R: 0xe5, G: 0x48, B: 0x4d, A: 0xff}
	devLineInk   = color.RGBA{R: 0xd9, G: 0x77, B: 0x06, A: 0xff}
	devTextInk   = color.RGBA{R: 0x17, G: 0x6b, B: 0x45, A: 0xff}
	devImageInk  = color.RGBA{R: 0x7c, G: 0x3a, B: 0xed, A: 0xff}
	devGridInk   = color.RGBA{R: 0x08, G: 0x91, B: 0xb2, A: 0xff}
)

// drawDevOps outlines every painted operation in paint order, colour by
// kind. Noop and unknown operations carry no paint and get no outline.
func (s *shell) drawDevOps(screen *ebiten.Image) {
	if s.display == nil {
		return
	}

	for _, index := range s.display.Order {
		if index < 0 || index >= len(s.display.Ops) {
			continue
		}

		op := &s.display.Ops[index]
		ink, ok := devOpInk(op.Kind)
		if !ok {
			continue
		}

		r := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
		vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 1, ink, false)
	}
}

// devOpInk returns the outline colour for one operation kind.
func devOpInk(kind layout.DisplayKind) (color.RGBA, bool) {
	switch kind {
	case layout.DisplayOpFillRect:
		return devFillInk, true
	case layout.DisplayOpStrokeRect:
		return devStrokeInk, true
	case layout.DisplayOpLine:
		return devLineInk, true
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return devTextInk, true
	case layout.DisplayOpImage:
		return devImageInk, true
	case layout.DisplayOpGridRun:
		return devGridInk, true
	}

	return color.RGBA{}, false
}
