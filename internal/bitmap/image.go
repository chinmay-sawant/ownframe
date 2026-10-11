package bitmap

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chinmay-sawant/blinkless/layout"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func clearWhite(img *image.NRGBA) {
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
}

func paintImage(dst *image.NRGBA, op *layout.DisplayOp) error {
	data, _, _ := op.ImageBytes()
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	b := src.Bounds()
	if b.Empty() || op.W <= 0 || op.H <= 0 {
		return nil
	}
	m := op.Transform()
	if !op.XformSet {
		m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
	}
	sx, sy := op.W*pxPerPt/float64(b.Dx()), op.H*pxPerPt/float64(b.Dy())
	x, y := transform(op, op.X, op.Y)
	matrix := f64.Aff3{m.A * sx, m.C * sy, x * pxPerPt, m.B * sx, m.D * sy, y * pxPerPt}
	mask := image.NewUniform(color.Alpha{A: channel(op.Opacity())})
	xdraw.BiLinear.Transform(dst, matrix, src, b, draw.Over, &xdraw.Options{SrcMask: mask})
	return nil
}
