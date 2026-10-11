package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func DrawUnfixedZ(dst *ebiten.Image, d *layout.Display, z int) {
	visitZ(d, func(op *layout.DisplayOp) {
		if scrollContentOp(op, z) {
			drawOp(dst, op, 0, 0)
		}
	})
}

func DrawRectUnfixedZ(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
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
	visitZ(d, func(op *layout.DisplayOp) {
		if scrollContentOp(op, z) && opTouches(op, d.PointsPerPixel, rect) {
			drawOp(clip, op, 0, 0)
		}
	})
}

func DrawVisibleUnfixedZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	visitVisibleZ(dst, d, dx, dy, z, false)
}

func DrawPinnedZ(dst *ebiten.Image, d *layout.Display, z int, dx, dy float64) {
	visitVisibleZ(dst, d, dx, dy, z, true)
}

func visitZ(d *layout.Display, visit func(*layout.DisplayOp)) {
	if d == nil {
		return
	}
	for _, i := range d.Order {
		if i >= 0 && i < len(d.Ops) {
			visit(&d.Ops[i])
		}
	}
}

func visitVisibleZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int, pinned bool) {
	if dst == nil || d == nil {
		return
	}
	r := visibleRect(dst.Bounds(), dx, dy)
	visitZ(d, func(op *layout.DisplayOp) {
		if op.Fixed || atPinZ(op, z) != pinned || !opTouches(op, d.PointsPerPixel, r) {
			return
		}
		drawOp(dst, op, dx, dy)
	})
}

func atPinZ(op *layout.DisplayOp, z int) bool {
	return z > 0 && op.ZIndexSet && op.ZIndex >= z
}

func scrollContentOp(op *layout.DisplayOp, z int) bool {
	return !op.Fixed && !atPinZ(op, z)
}
