package window

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devOpPickWidth is the selection outline width; the stroke reaches half
// past each edge.
const (
	devOpPickWidth  = 3.0
	devOpPickExpand = 1.5
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
	vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devOpPickWidth, ink, false)
	// The 3 px stroke reaches 1.5 px past each edge.
	s.frameNote(frameRect(r.X-devOpPickExpand, r.Y-devOpPickExpand, r.W+devOpPickWidth, r.H+devOpPickWidth))
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
