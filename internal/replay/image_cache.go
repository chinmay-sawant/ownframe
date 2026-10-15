package replay

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
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
// Bytes to pixels, the empty-or-bad classify, and the alpha rule live in
// masking; this fork only uploads. It returns nil for a skip (nil op, empty
// payload, or empty decode) or a decode error, so the caller skips the draw.
func decodedImage(op *layout.DisplayOp) *ebiten.Image {
	if op == nil {
		return nil
	}
	key := imageBytesKey(op)
	images.mu.RLock()
	if img := images.entries[key]; img != nil {
		images.mu.RUnlock()
		return img
	}
	images.mu.RUnlock()
	src, _, skip, err := masking.DecodeImage(op)
	if skip || err != nil {
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
