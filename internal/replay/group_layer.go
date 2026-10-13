package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintGroup draws one buffered group: the visible members into a layer the
// size of the group's window, then one composite with the group's mode.
func paintGroup(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker) {
	rect, _, ok := groupWindow(dst, d, n, pick)
	if !ok {
		return
	}

	layer := ebiten.NewImage(rect.Dx(), rect.Dy())
	defer layer.Dispose()

	inner := func(op *layout.DisplayOp) (float64, float64, bool) {
		dx, dy, ok := pick(op)

		return dx - float64(rect.Min.X), dy - float64(rect.Min.Y), ok
	}

	paintItems(layer, d, n, inner)
	compositeLayer(dst, layer, rect.Min, n.group.Mode)
}

// groupWindow is the device rect a group's layer covers: its page box moved
// by the first visible member's translation and clipped to dst. base is that
// translation; ok is false when nothing in the group paints.
func groupWindow(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker) (image.Rectangle, [2]float64, bool) {
	base, ok := firstPick(n, pick)
	if !ok {
		return image.Rectangle{}, [2]float64{}, false
	}

	box, bounded := nodeBox(n, d.PointsPerPixel)
	if !bounded {
		box = dst.Bounds()
	} else {
		box = image.Rect(
			int(math.Floor(float64(box.Min.X)+base[0])),
			int(math.Floor(float64(box.Min.Y)+base[1])),
			int(math.Ceil(float64(box.Max.X)+base[0])),
			int(math.Ceil(float64(box.Max.Y)+base[1])),
		)
	}

	box = box.Intersect(dst.Bounds())
	if box.Empty() {
		return image.Rectangle{}, [2]float64{}, false
	}

	return box, base, true
}
