package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// decodedImage decodes op's payload once and caches the Ebiten image.
// Bytes to pixels, caps, and alpha live in masking; this fork only uploads.
// It returns nil for a skip or decode error, so the caller skips the draw.
func decodedImage(op *layout.DisplayOp) *ebiten.Image {
	if op == nil {
		return nil
	}
	key := imageBytesKey(op)
	if img := images.get(key); img != nil {
		return img
	}
	src, _, skip, err := masking.DecodeImage(op)
	if skip || err != nil {
		return nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if masking.PixelsTooBig(w, h) {
		return nil
	}
	return images.put(key, src, w*h)
}
