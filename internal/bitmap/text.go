package bitmap

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/chinmay-sawant/blinkless/layout"
)

func paintText(dst *image.NRGBA, op *layout.DisplayOp, cache faceCache) error {
	if op.Text == "" {
		return nil
	}

	data := op.Font.Bytes()
	if len(data) == 0 || op.Size <= 0 {
		return fmt.Errorf("bitmap: invalid text font or size")
	}

	face := cache.face(data, op.Size)
	if face == nil {
		return fmt.Errorf("bitmap: cannot decode text font")
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	x, y := origin(op)
	draw := font.Drawer{
		Dst: dst,
		Src: image.NewUniform(color.NRGBA{
			R: channel(op.R), G: channel(op.G), B: channel(op.B), A: channel(op.Opacity()),
		}),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.Int26_6(math.Round(x * pxPerPt * 64)),
			Y: fixed.Int26_6(math.Round(y * pxPerPt * 64)),
		},
	}
	drawString(&draw, content, op.LetterSpacing*pxPerPt)

	if layout.DisplayFakeBold(op) {
		draw.Dot = fixed.Point26_6{X: fixed.Int26_6(math.Round(x*pxPerPt*64)) + fixed.I(1), Y: fixed.Int26_6(math.Round(y * pxPerPt * 64))}
		drawString(&draw, content, op.LetterSpacing*pxPerPt)
	}
	return nil
}
