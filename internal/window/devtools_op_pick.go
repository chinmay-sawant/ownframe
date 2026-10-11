package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawDevOpPick outlines the operation chosen in the Ops list in its kind
// colour, thicker than the ops view so it reads as the selection.
func (s *shell) drawDevOpPick(screen *ebiten.Image) {
	if s.display == nil || s.dev.opPick < 0 || s.dev.opPick >= len(s.display.Ops) {
		return
	}

	op := &s.display.Ops[s.dev.opPick]

	ink, ok := devOpInk(op.Kind)
	if !ok {
		return
	}

	r := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
	vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 3, ink, false)
}
