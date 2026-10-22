package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// devOpsHalf is the half-pixel reach of the 1 px outline; devOpsFull
// grows the union box by the full pixel.
const (
	devOpsHalf = 0.5
	devOpsFull = 1.0
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
		union = union.Union(frameRect(r.X-devOpsHalf, r.Y-devOpsHalf, r.W+devOpsFull, r.H+devOpsFull))
	}

	if !union.Empty() {
		s.frameNote(union)
	}
}
