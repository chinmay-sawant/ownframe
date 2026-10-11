// Package bitmap paints display operations into a software image for fallback
// pages and PNG export. It consumes layout geometry without applying CSS.
package bitmap

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

const pxPerPt = 96.0 / 72.0

// Picture returns the software painting, or nil if painting is unsupported.
func Picture(d *layout.Display) image.Image {
	img, _ := Paint(d)
	return img
}

// Paint reports unsupported content instead of silently omitting it.
func Paint(d *layout.Display) (image.Image, error) {
	if d == nil || d.Width < 1 || d.Height < 1 {
		return nil, nil
	}
	if err := validate(d); err != nil {
		return nil, err
	}
	img := image.NewNRGBA(image.Rect(0, 0, d.Width, d.Height))
	clearWhite(img)
	faces := faceCache{}
	defer faces.close()
	if err := paintOrdered(img, d, faces); err != nil {
		return nil, err
	}
	return img, nil
}
