package window

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
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

// devOpTextCap caps the quoted text runes in an Ops row.
const devOpTextCap = 40

// devOpRow builds one Ops-tab row for the operation at index i, the nth
// painted operation.
func devOpRow(op *layout.DisplayOp, i, n int, ink color.RGBA, ppt float64) devRow {
	x := devOpRect(op, ppt)
	sp := []devSpan{
		{text: fmt.Sprintf("#%d ", n), ink: devDim},
		{text: devOpsListName[op.Kind], ink: ink},
		{text: fmt.Sprintf(" x%g y%g %gx%g", x.X, x.Y, x.W, x.H), ink: devDim},
	}
	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		t := op.Text
		if len([]rune(t)) > devOpTextCap {
			t = string([]rune(t)[:devOpTextCap]) + "..."
		}
		sp = append(sp, devSpan{text: fmt.Sprintf(" %q", t), ink: devStrInk})
	}

	return devRow{line: devLine{spans: sp}, act: devActOp, arg: i}
}

// devOpsView is the page area an outline may land in: the window minus the
// dock, so a stroke under the panel or off the window is skipped. The dock
// covering the whole window leaves it empty.
func (s *shell) devOpsView() devRect {
	if s.dev.mobile {
		return devRect{W: float64(s.screenW), H: s.dev.panel.Y}
	}
	if s.dev.panel.W <= 0 {
		return devRect{W: float64(s.screenW), H: float64(s.screenH)}
	}

	return devRect{W: s.dev.panel.X, H: float64(s.screenH)}
}

// devRectsMeet reports whether two screen rects overlap.
func devRectsMeet(a, b devRect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}
