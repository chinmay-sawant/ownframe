package bitmap

import (
	"image"
	"slices"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Viewport paints a scrolled, zoomed view without moving fixed operations.
// Copies carry translated matrices; the retained engine operations stay intact.
func Viewport(d *layout.Display, rect image.Rectangle, zoom float64) (image.Image, error) {
	view := *d
	view.Width, view.Height = rect.Dx(), rect.Dy()
	view.Ops = slices.Clone(d.Ops)
	for i := range view.Ops {
		op := &view.Ops[i]
		m := op.Transform()
		if !op.XformSet {
			m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
		}
		m.A, m.B, m.C, m.D = m.A*zoom, m.B*zoom, m.C*zoom, m.D*zoom
		m.E, m.F = m.E*zoom, m.F*zoom
		if !op.Fixed {
			m.E -= float64(rect.Min.X) / pxPerPt
			m.F -= float64(rect.Min.Y) / pxPerPt
		}
		op.SetXform(m)
	}
	return Paint(&view)
}
