package bitmap

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

func paintOp(img *image.NRGBA, op *layout.DisplayOp, faces faceCache) error {
	switch op.Kind {
	case layout.DisplayOpFillRect, layout.DisplayOpStrokeRect, layout.DisplayOpLine:
		paintShape(img, op)
	case layout.DisplayOpGridRun:
		for _, seg := range op.Grid.Segs {
			line := *op
			line.Kind = layout.DisplayOpLine
			line.X, line.Y, line.W, line.H = seg.X, seg.Y, seg.W, seg.H
			line.Width, line.LineInset = seg.Width, seg.LineInset
			line.R, line.G, line.B = seg.R, seg.G, seg.B
			paintShape(img, &line)
		}
	case layout.DisplayOpImage:
		return paintImage(img, op)
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return paintTextOp(img, op, faces)
	}
	return nil
}
