package bitmap

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/f64"
)

func (c faceCache) close() {
	for _, face := range c {
		if face != nil {
			face.Close()
		}
	}
}

func drawString(d *font.Drawer, s string, spacing float64) {
	if spacing == 0 {
		d.DrawString(s)
		return
	}
	for _, r := range s {
		d.DrawString(string(r))
		d.Dot.X += fixedSpacing(spacing)
	}
}

func paintTextOp(dst *image.NRGBA, op *layout.DisplayOp, faces faceCache) error {
	if !op.XformSet && op.RotateDeg == 0 && !op.FakeOblique {
		return paintText(dst, op, faces)
	}
	// Bound the untransformed glyph canvas, including ascent and italic overhang.
	pad := op.Size * pxPerPt * 2
	b := image.Rect(int(op.X*pxPerPt-pad), int(op.Y*pxPerPt-pad),
		int((op.X+op.W)*pxPerPt+pad), int(op.Y*pxPerPt+pad))
	depth := 1
	for g := op.Group(); g != nil; g = g.Parent {
		depth++
	}
	budget := maxPixels - depth*dst.Bounds().Dx()*dst.Bounds().Dy()
	if b.Empty() || b.Dx() > budget/max(1, b.Dy()) {
		return fmt.Errorf("bitmap: transformed text exceeds buffer budget")
	}
	tmp := image.NewNRGBA(b)
	if err := paintText(tmp, op, faces); err != nil {
		return err
	}
	m := op.Transform()
	if !op.XformSet {
		m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
	}
	angle := float64(op.RotateDeg) * math.Pi / 180
	c, s := math.Cos(angle), math.Sin(angle)
	shear := 0.0
	if op.FakeOblique {
		shear = -.2
	}
	a, bv, cv, d := c, s, c*shear-s, s*shear+c
	x, y := op.X*pxPerPt, op.Y*pxPerPt
	tx, ty := x-a*x-cv*y, y-bv*x-d*y
	matrix := f64.Aff3{m.A*a + m.C*bv, m.A*cv + m.C*d, m.A*tx + m.C*ty + m.E*pxPerPt,
		m.B*a + m.D*bv, m.B*cv + m.D*d, m.B*tx + m.D*ty + m.F*pxPerPt}
	xdraw.BiLinear.Transform(dst, matrix, tmp, tmp.Bounds(), draw.Over, nil)
	return nil
}
