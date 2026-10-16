package page_test

import "image"

// samePicture reports whether two decoded pictures share bounds and every
// pixel, so encoder byte drift never fails a paint-equality check.
func samePicture(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}

	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}
