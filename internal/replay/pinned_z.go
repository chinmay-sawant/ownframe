package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func (renderer *Renderer) DrawUnfixedZ(dst *ebiten.Image, d *layout.Display, z int) {
	renderer.paintChunked(dst, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, scrollContentOp(op, z)
	})
}

func (renderer *Renderer) DrawRectUnfixedZ(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
	if dst == nil || d == nil {
		return
	}

	rect = clipRect(d, rect).Intersect(dst.Bounds())
	if rect.Empty() {
		return
	}

	clip, ok := dst.SubImage(rect).(*ebiten.Image)
	if !ok {
		return
	}

	ppt := d.PointsPerPixel
	renderer.paintChunked(clip, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, scrollContentOp(op, z) && opTouches(op, ppt, rect)
	})
}

func (renderer *Renderer) DrawVisibleUnfixedZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	renderer.visitVisibleZ(dst, d, dx, dy, z, false)
}

func (renderer *Renderer) DrawPinnedZ(dst *ebiten.Image, d *layout.Display, z int, dx, dy float64) {
	renderer.visitVisibleZ(dst, d, dx, dy, z, true)
}

func (renderer *Renderer) visitVisibleZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int, pinned bool) {
	if dst == nil || d == nil {
		return
	}

	r := visibleRect(dst.Bounds(), dx, dy)
	ppt := d.PointsPerPixel

	renderer.paintChunked(dst, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		if op.Fixed || atPinZ(op, z) != pinned || !opTouches(op, ppt, r) {
			return 0, 0, false
		}

		return dx, dy, true
	})
}

func atPinZ(op *layout.DisplayOp, z int) bool {
	return z > 0 && op.ZIndexSet && op.ZIndex >= z
}

func scrollContentOp(op *layout.DisplayOp, z int) bool {
	return !op.Fixed && !atPinZ(op, z)
}
