package window

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devOutline strokes one box and labels it. The stroke hook lets a test read
// the computed rect without reading pixels.
func (s *shell) devOutline(screen *ebiten.Image, box layout.Box, ink color.RGBA) {
	r := s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})

	if s.dev.stroke != nil {
		s.dev.stroke(screen, r, ink)
	} else {
		vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 2, ink, false)
	}

	// The 2 px stroke reaches 1 px past each edge.
	lr := s.devLabel(screen, r, devBoxLabel(box))
	s.frameNote(frameRect(r.X-1, r.Y-1, r.W+2, r.H+2).Union(frameRect(lr.X, lr.Y, lr.W, lr.H)))
}

// devBoxLabel is the tag, id, and size line drawn over a box.
func devBoxLabel(box layout.Box) string {
	name := box.Tag
	if box.ID != "" {
		name += "#" + box.ID
	}

	return fmt.Sprintf("%s  %gx%g", name, box.W, box.H)
}

// devLabel draws a box label above the outline, or below it when the top
// edge would clip, and returns the rect it painted.
func (s *shell) devLabel(screen *ebiten.Image, r devRect, label string) devRect {
	w, h := text.Measure(label, badgeFace, 0)
	x, y := r.X, r.Y-h-4

	if y < 0 {
		y = r.Y + r.H + 4
	}

	vector.FillRect(screen, float32(x), float32(y), float32(w+8), float32(h+4), badgeInk, false)

	var op text.DrawOptions
	op.GeoM.Translate(x+4, y+2)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, label, badgeFace, &op)

	return devRect{X: x, Y: y, W: w + 8, H: h + 4}
}
