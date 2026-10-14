package window

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawDevOps outlines every painted operation in paint order, colour by
// kind, culled to the page area the dock leaves visible. Noop and unknown
// operations carry no paint and get no outline. Only the strokes are culled;
// the row list still holds every operation, so picking is not affected.
// The outlines are one chrome region: the union of the strokes that were
// actually painted is noted once.
func (s *shell) drawDevOps(screen *ebiten.Image) {
	if s.display == nil {
		return
	}

	view := s.devOpsView()
	var union image.Rectangle

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
		if !devRectsMeet(r, view) {
			continue
		}

		if s.dev.stroke != nil {
			s.dev.stroke(screen, r, ink)
		} else {
			vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 1, ink, false)
		}

		// The 1 px stroke reaches half a pixel past each edge.
		union = union.Union(frameRect(r.X-0.5, r.Y-0.5, r.W+1, r.H+1))
	}

	if !union.Empty() {
		s.frameNote(union)
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
