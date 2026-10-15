package replay

import (
	"bytes"
	"image"
	_ "image/jpeg" // registers the JPEG decoder for image.Decode
	_ "image/png"  // registers the PNG decoder for image.Decode
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// imageCacheLimit bounds the decoded images kept. A page carries far fewer
// images than this, so the cap only matters across many pages.
const imageCacheLimit = 64

// imageCache keeps decoded images by payload key, oldest insertion first.
// Hits resolve under a read lock; decode runs lock-free and only the GPU
// upload plus insert hold the write lock, with a second hit check.
type imageCache struct {
	mu      sync.RWMutex
	entries map[uint64]*ebiten.Image
	order   []uint64
}

var images imageCache

// decodedImage decodes op's encoded payload once and caches the Ebiten image.
// It returns nil for a nil op, an empty payload, or a decode error, so the
// caller skips the draw instead of failing the frame.
func decodedImage(op *layout.DisplayOp) *ebiten.Image {
	if op == nil {
		return nil
	}

	data, _, _ := op.ImageBytes()
	if len(data) == 0 {
		return nil
	}

	key := imageBytesKey(op)

	images.mu.RLock()
	if img := images.entries[key]; img != nil {
		images.mu.RUnlock()
		return img
	}
	images.mu.RUnlock()

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	images.mu.Lock()
	defer images.mu.Unlock()

	if img := images.entries[key]; img != nil {
		return img
	}

	img := ebiten.NewImageFromImage(src)

	if images.entries == nil {
		images.entries = make(map[uint64]*ebiten.Image)
	}

	images.entries[key] = img
	images.order = append(images.order, key)

	if len(images.order) > imageCacheLimit {
		images.entries[images.order[0]].Dispose()
		delete(images.entries, images.order[0])
		images.order = images.order[1:]
	}

	return img
}
