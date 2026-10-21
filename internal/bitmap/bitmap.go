// Package bitmap paints display operations into a software image for fallback
// pages and PNG export. It consumes layout geometry without applying CSS.
package bitmap

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

const cssDPI = 96.0
const pointsPerInch = 72.0
const pxPerPt = cssDPI / pointsPerInch

// Paint returns the software painting. A nil or empty display paints
// nothing and reports no error; unsupported content is an error the caller
// handles, so no silent picture ever leaves here.
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
