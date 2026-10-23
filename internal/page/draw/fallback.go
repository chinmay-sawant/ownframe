package draw

import (
	"image"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Bitmap is the picture and the patch counts a bitmap frame records.
// Changed and Added are the operation count. Regions is 1. Full is true.
type Bitmap struct {
	Image   image.Image
	Changed int
	Added   int
	Regions int
	Full    bool
}

// PaintBitmap paints display, and when track is true stores the paint duration in paintTime.
// A nil display or a Picture error returns the error. A nil paintTime skips the store.
// Counts match what page.paintFallback wrote into DisplayPatch.
func PaintBitmap(track bool, paintTime *time.Duration, display *layout.Display) (Bitmap, error) {
	start := Clock(track)
	img, err := Picture(display)
	if err != nil || display == nil {
		return Bitmap{}, err
	}
	if paintTime != nil {
		Note(track, paintTime, start)
	}

	ops := len(display.Ops)

	return Bitmap{
		Image:   img,
		Changed: ops,
		Added:   ops,
		Regions: 1,
		Full:    true,
	}, nil
}
