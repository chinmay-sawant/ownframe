package replay

import (
	"bytes"
	"fmt"
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// checkTextClip pins partial-repaint clipping of baked and direct runs.
func checkTextClip() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}

	display := &layout.Display{
		Width: 320, Height: 240, PointsPerPixel: pxPerPt,
		Ops: ops, Order: []int{0, 1, 2, 3, 4, 5},
	}
	rect := image.Rect(8, 40, 220, 232)

	direct := ebiten.NewImage(320, 240)
	cached := ebiten.NewImage(320, 240)
	defer direct.Dispose()
	defer cached.Dispose()

	clip, ok := direct.SubImage(rect).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("sub-image")
	}
	for i := range ops {
		drawTextDirect(&ops[i], clip)
	}
	DrawRect(cached, display, rect, 0, 0)

	a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("clipped text raster differs in %d channels", countDiff(a, b))
	}

	return nil
}

// drawTextDirect draws one text op through the uncached path only.
func drawTextDirect(op *layout.DisplayOp, dst *ebiten.Image) {
	face := textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	drawTextRun(dst, face, content, op.X*pxPerPt, op.Y*pxPerPt-face.Metrics().HAscent, op)
}

// phase is the run's keyed sub-pixel offset, built the way cachedTextRun
// builds it from the op and the face.
func phase(op *layout.DisplayOp, face *text.GoTextFace) [2]float64 {
	x := op.X * pxPerPt
	y := op.Y*pxPerPt - face.Metrics().HAscent

	return [2]float64{textPhase(x - math.Floor(x)), textPhase(y - math.Floor(y))}
}
