package replay

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// chunkMaxEntries and chunkMaxPixels bound the chunk images one renderer
// keeps. Oldest insertion evicts first, and its image is disposed. A paint
// only needs the chunks of the region it draws, so the cap covers the
// visible rows rather than the whole page.
const (
	chunkMaxEntries = 256
	chunkMaxPixels  = 4 << 20
)

// chunkKey names one chunk's cache slot: the sequence of the run in a
// display's chunk plan. A display is rebuilt on every redraw, so the key
// carries no display identity: the content signature decides whether the
// slot still holds the same pixels, so a redraw with the same content reuses
// the baked image instead of baking it again, and the cache no longer pins
// the displays it came from.
//
// The signature has to survive a redraw for that to pay off, so it folds
// values a layout reproduces (see chunk_sig_op.go) and never a per-run
// pointer. A renderer owns these slots, so a second page cannot replace them.
type chunkKey struct {
	id int
}

// chunkEntry is one baked chunk: the image, the content signature it holds,
// and the page box its origin maps to.
type chunkEntry struct {
	img *ebiten.Image
	sig uint64
	box image.Rectangle
}

type chunkStore struct {
	mu      sync.Mutex
	entries map[chunkKey]*chunkEntry
	order   []chunkKey
	pixels  int
}
