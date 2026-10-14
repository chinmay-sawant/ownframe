package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// groupWindow returns the group's translated paint box clipped to dst.
func groupWindow(dst *ebiten.Image, d *layout.Display, n *groupNode,
	pick opPicker) (image.Rectangle, [2]float64, bool) {
	base, ok := firstPick(n, pick)
	if !ok {
		return image.Rectangle{}, [2]float64{}, false
	}
	box, bounded := nodeBox(n, d.PointsPerPixel)
	if n.group != nil && n.group.Mask != nil && n.hasMask {
		box, bounded = n.maskBox, true
	}
	if !bounded {
		box = dst.Bounds()
	} else {
		box = image.Rect(int(math.Floor(float64(box.Min.X)+base[0])),
			int(math.Floor(float64(box.Min.Y)+base[1])),
			int(math.Ceil(float64(box.Max.X)+base[0])),
			int(math.Ceil(float64(box.Max.Y)+base[1])))
	}
	box = box.Intersect(dst.Bounds())
	if box.Empty() {
		return image.Rectangle{}, [2]float64{}, false
	}
	return box, base, true
}
