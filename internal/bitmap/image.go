package bitmap

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func clearWhite(img *image.NRGBA) {
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
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
	m := op.Transform()
	if !op.XformSet {
		m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
	}
	sx, sy := op.W*pxPerPt/float64(b.Dx()), op.H*pxPerPt/float64(b.Dy())
	x, y := transform(op, op.X, op.Y)
	matrix := f64.Aff3{m.A * sx, m.C * sy, x * pxPerPt, m.B * sx, m.D * sy, y * pxPerPt}
	mask := image.NewUniform(color.Alpha{A: masking.Channel(alpha)})
	xdraw.BiLinear.Transform(dst, matrix, src, b, draw.Over, &xdraw.Options{SrcMask: mask})
	return nil
}
