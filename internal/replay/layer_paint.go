package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// layerFrame is the paint state BeginLayers installs for one display: the
// owned operation indices to skip, and the order position at which each layer
// composites. The map holds at most layerMaxEntries entries.
type layerFrame struct {
	display *layout.Display
	skip    []bool
	start   map[int]*Layer
	active  []*Layer
}

func (f *layerFrame) reset() {
	f.display = nil
	f.skip = f.skip[:0]
	if f.start != nil {
		clear(f.start)
	}
	f.active = f.active[:0]
}

// DrawVisibleUnfixedZLayers replays the display in paint order, but every
// operation owned by an active layer is skipped and the layer is composited
// once at the position its first owned op occupies. dst is clipped by Ebiten,
// so a tile repaint automatically clips the layer to its tile.
func (renderer *Renderer) DrawVisibleUnfixedZLayers(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	if dst == nil || d == nil {
		return
	}
	frame := &renderer.layers.frame
	if frame.display != d || len(frame.active) == 0 {
		renderer.DrawVisibleUnfixedZ(dst, d, dx, dy, z)

		return
	}

	rect := visibleRect(dst.Bounds(), dx, dy)
	ppt := d.PointsPerPixel

	for rank, index := range d.Order {
		if layer := frame.start[rank]; layer != nil && layerTouches(layer, rect) {
			renderer.drawLayerPose(dst, d, layer, dx, dy)
		}
		if index < 0 || index >= len(d.Ops) || index >= len(frame.skip) || frame.skip[index] {
			continue
		}
		op := &d.Ops[index]
		if op.Fixed || atPinZ(op, z) || !opTouches(op, ppt, rect) {
			continue
		}
		countOp()
		renderer.drawOp(dst, op, dx, dy)
	}
}

// DrawRectUnfixedZLayers is the layer-aware DrawRectUnfixedZ. It paints only
// the part of dst that rect covers, skips layer-owned operations, and
// composites each layer once, clipped to dst.
func (renderer *Renderer) DrawRectUnfixedZLayers(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
	if dst == nil || d == nil {
		return
	}
	frame := &renderer.layers.frame
	if frame.display != d || len(frame.active) == 0 {
		renderer.DrawRectUnfixedZ(dst, d, rect, z)

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
	for rank, index := range d.Order {
		if layer := frame.start[rank]; layer != nil && layerTouches(layer, rect) {
			renderer.drawLayerPose(clip, d, layer, 0, 0)
		}
		if index < 0 || index >= len(d.Ops) || index >= len(frame.skip) || frame.skip[index] {
			continue
		}
		op := &d.Ops[index]
		if op.Fixed || atPinZ(op, z) || !opTouches(op, ppt, rect) {
			continue
		}
		countOp()
		renderer.drawOp(clip, op, 0, 0)
	}
}

// DrawVisibleUnfixedZBackground replays the display without the operations a
// promoted layer owns, and does not composite the layers. The caller keeps the
// result as the page background: the owned operations leave holes, so a layer
// can move over the background without the page under it being repainted.
//
// A page with no active layers takes the ordinary replay, so the background is
// also the complete frame.
func (renderer *Renderer) DrawVisibleUnfixedZBackground(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	if dst == nil || d == nil {
		return
	}

	frame := &renderer.layers.frame
	if frame.display != d || len(frame.active) == 0 {
		renderer.DrawVisibleUnfixedZ(dst, d, dx, dy, z)

		return
	}

	rect := visibleRect(dst.Bounds(), dx, dy)
	ppt := d.PointsPerPixel

	for _, index := range d.Order {
		if index < 0 || index >= len(d.Ops) || index >= len(frame.skip) || frame.skip[index] {
			continue
		}
		op := &d.Ops[index]
		if op.Fixed || atPinZ(op, z) || !opTouches(op, ppt, rect) {
			continue
		}
		countOp()
		renderer.drawOp(dst, op, dx, dy)
	}
}

// DrawLayers composites every active retained layer onto dst in paint order.
// The caller must already have blitted the background DrawVisibleUnfixedZBackground
// produced, at the same dx and dy.
func (renderer *Renderer) DrawLayers(dst *ebiten.Image, d *layout.Display, dx, dy float64) {
	if dst == nil || d == nil {
		return
	}

	frame := &renderer.layers.frame
	if frame.display != d || len(frame.active) == 0 {
		return
	}

	for rank := range d.Order {
		layer := frame.start[rank]
		if layer == nil {
			continue
		}
		renderer.drawLayerPose(dst, d, layer, dx, dy)
	}
}

func (renderer *Renderer) drawLayerPose(dst *ebiten.Image, d *layout.Display, layer *Layer, dx, dy float64) {
	if layer == nil || layer.Alpha <= 0 {
		// A fully transparent layer composites nothing. The tile is still
		// damaged by BeginLayers, so the page beneath it is repainted without
		// a wasted blit.
		return
	}
	countBlit()
	compositeLayerPose(dst, layer.Tex, layerGeoM(layer, d, dx, dy), layer.Alpha, layer.Mode)
}
