package dirty

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// labelGap is how near a label must sit to a toggle to repaint with it.
// It is wider than Gap: a label is not a neighbouring damage rectangle.
const labelGap = 12

// boxRound nudges a box's far edges to the nearest whole CSS pixel.
const boxRound = 0.5

// BoxRect is one box in whole CSS pixels. The origin is truncated; the
// far edges gain boxRound before the truncation.
func BoxRect(b layout.Box) image.Rectangle {
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+boxRound), int(b.Y+b.H+boxRound))
}

// GrowBox covers the boxes inside r, so a nested span repaints with its
// parent. An empty r is returned unchanged.
func GrowBox(r image.Rectangle, boxes []layout.Box) image.Rectangle {
	if r.Empty() {
		return r
	}

	for _, b := range boxes {
		inner := BoxRect(b)
		if !inner.Empty() && inner.In(r) {
			r = r.Union(inner)
		}
	}

	return r
}

// ToggleHalo is the margin around a checkbox or radio that covers an
// inline label the layout did not box separately.
func ToggleHalo(ctl image.Rectangle) image.Rectangle {
	return ctl.Inset(-labelGap)
}

// LabelNear reports b's box when b is a label within labelGap of the
// control. Otherwise it reports false.
func LabelNear(ctl image.Rectangle, b layout.Box) (image.Rectangle, bool) {
	if b.Tag != "label" {
		return image.Rectangle{}, false
	}

	r := BoxRect(b)
	if !touches(ctl, r, labelGap) {
		return image.Rectangle{}, false
	}

	return r, true
}
