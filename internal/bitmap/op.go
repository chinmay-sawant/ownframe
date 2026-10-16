package bitmap

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

func paintOp(img *image.NRGBA, op *layout.DisplayOp, faces faceCache) error {
	switch op.Kind {
	case layout.DisplayOpFillRect, layout.DisplayOpStrokeRect, layout.DisplayOpLine:
		paintShape(img, op)
	case layout.DisplayOpGridRun:
		if op.Grid == nil {
			return nil
		}
		for i := range op.Grid.Segs {
			line := masking.GridLine(op, i)
			paintShape(img, &line)
		}
	case layout.DisplayOpImage:
		return paintImage(img, op)
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return paintTextOp(img, op, faces)
	}
	return nil
}
