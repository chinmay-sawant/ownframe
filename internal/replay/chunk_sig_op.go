package replay

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// sigOp folds every op field that shapes its pixels into h. Fields a tick can
// rewrite in place are covered, so a mutated op misses the cached raster.
func sigOp(h uint64, op *layout.DisplayOp) uint64 {
	h = mix64(h, uint64(op.Kind))

	for _, v := range [15]float64{
		op.X, op.Y, op.W, op.H, op.R, op.G, op.B, op.Alpha,
		op.Width, op.Size, op.LetterSpacing, op.InkDescent,
		op.Radius, op.RadiusY, op.Opacity(),
	} {
		h = mixF(h, v)
	}

	for _, v := range [4]float64{
		op.RadiusTopLeft, op.RadiusTopRight, op.RadiusBottomRight, op.RadiusBottomLeft,
	} {
		h = mixF(h, v)
	}

	for _, v := range [4]float64{
		op.RadiusTopLeftY, op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY,
	} {
		h = mixF(h, v)
	}

	h = mix64(h, uint64(op.StrokeMask))
	h = mix64(h, uint64(op.LineInset))
	h = mix64(h, uint64(op.ZIndex))
	h = mix64(h, bit(op.Bold)|bit(op.FakeOblique)<<1|bit(op.Fixed)<<2|
		bit(op.ZIndexSet)<<3|bit(op.NoFakeBoldValue())<<4)
	h = mix64(h, uint64(math.Float32bits(op.RotateDeg)))
	h = mixStr(h, op.Text)
	h = mixStr(h, op.TextTransformValue())
	h = mixStr(h, op.TextLanguage())
	h = mix64(h, ptrBits(op.Font))
	h = mix64(h, ptrBits(op.Group()))

	return sigXform(sigImage(sigGrid(h, op), op), op)
}
