package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintGroup draws one buffered group: the visible members into a layer the
// size of the group's window, then one composite with the group's mode.
func paintGroup(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker) {
	rect, base, ok := groupWindow(dst, d, n, pick)
	if !ok {
		return
	}

	layer := ebiten.NewImage(rect.Dx(), rect.Dy())
	defer layer.Deallocate()

	inner := func(op *layout.DisplayOp) (float64, float64, bool) {
		dx, dy, ok := pick(op)

		return dx - float64(rect.Min.X), dy - float64(rect.Min.Y), ok
	}

	paintItems(layer, d, n, inner)
	if n.group.Mask != nil {
		mask := groupMaskTexture(n)
		if mask == nil {
			return
		}
		opts := &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn}
		opts.GeoM.Translate(float64(n.maskBox.Min.X)+base[0]-float64(rect.Min.X),
			float64(n.maskBox.Min.Y)+base[1]-float64(rect.Min.Y))
		layer.DrawImage(mask, opts)
	}
	compositeLayer(dst, layer, rect.Min, n.group.Mode)
}
