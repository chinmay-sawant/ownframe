package bitmap

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func clearWhite(img *image.NRGBA) {
	pix := img.Pix
	const seedSize = 32
	filled := min(len(pix), seedSize)
	for i := range pix[:filled] {
		pix[i] = channelMax
	}
	for filled < len(pix) {
		filled += copy(pix[filled:], pix[:filled])
	}
}

func paintImage(dst *image.NRGBA, op *layout.DisplayOp) error {
	src, alpha, skip, err := masking.DecodeImage(op)
	if skip {
		return nil
	}
	if err != nil {
		return err
	}
	if masking.SkipImageSize(op) {
		return nil
	}
	b := src.Bounds()
	maskAlpha := masking.Channel(alpha)
	var opts *xdraw.Options
	if maskAlpha < channelMax {
		opts = &xdraw.Options{SrcMask: image.NewUniform(color.Alpha{A: maskAlpha})}
	}

	m := op.Transform()
	if !op.XformSet {
		m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
	}

	// Axis-aligned (unrotated, unskewed) images use Scale/Draw directly instead
	// of general affine Transform, keeping identical geometry across both
	// XformSet=false and Viewport-translated XformSet=true paths.
	if m.B == 0 && m.C == 0 && m.A > 0 && m.D > 0 {
		x0, y0 := (m.A*op.X+m.E)*pxPerPt, (m.D*op.Y+m.F)*pxPerPt
		x1, y1 := (m.A*(op.X+op.W)+m.E)*pxPerPt, (m.D*(op.Y+op.H)+m.F)*pxPerPt
		dr := image.Rect(
			int(math.Round(x0)),
			int(math.Round(y0)),
			int(math.Round(x1)),
			int(math.Round(y1)),
		)
		if dr.Intersect(dst.Bounds()).Empty() {
			return nil
		}
		if dr.Dx() == b.Dx() && dr.Dy() == b.Dy() && opts == nil {
			draw.Draw(dst, dr, src, b.Min, draw.Over)
			return nil
		}
		xdraw.BiLinear.Scale(dst, dr, src, b, draw.Over, opts)
		return nil
	}

	sx, sy := op.W*pxPerPt/float64(b.Dx()), op.H*pxPerPt/float64(b.Dy())
	x, y := transform(op, op.X, op.Y)
	matrix := f64.Aff3{m.A * sx, m.C * sy, x * pxPerPt, m.B * sx, m.D * sy, y * pxPerPt}
	xdraw.BiLinear.Transform(dst, matrix, src, b, draw.Over, opts)
	return nil
}
