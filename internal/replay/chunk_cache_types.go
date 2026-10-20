package replay

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// chunkMaxEntries and chunkMaxPixels bound the chunk images one process
// keeps. Oldest insertion evicts first, and its image is disposed. The pixel
// cap has to hold the working set of a long page, or every frame rebakes
// every chunk and the process runs out of memory.
const (
	chunkMaxEntries = 256
	chunkMaxPixels  = 24 << 20
)

// chunkKey names one chunk's cache slot: the display it belongs to and its
// sequence in that display's plan.
type chunkKey struct {
	d  *layout.Display
	id int
}

// chunkEntry is one baked chunk: the image, the content signature it holds,
// and the page box its origin maps to.
type chunkEntry struct {
	img *ebiten.Image
	sig uint64
	box image.Rectangle
}

var chunks chunkStore

type chunkStore struct {
	mu      sync.Mutex
	entries map[chunkKey]*chunkEntry
	order   []chunkKey
	pixels  int
}
