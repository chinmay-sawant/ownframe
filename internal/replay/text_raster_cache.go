package replay

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// textRunMaxEntries and textRunMaxPixels bound what one page caches. A page
// with more runs than this keeps the most-recently-used; the
// least-recently-used rasters are disposed.
const (
	textRunMaxEntries = 512
	textRunMaxPixels  = 4 << 20
)

// textRasterEntry is one baked run, offset from the integer origin the key
// names. A zero entry is a remembered decline: the run cannot be cached, so
// later draws take the direct path without reshaping to learn that again.
type textRasterEntry struct {
	img    *ebiten.Image
	dx, dy int
}

// pixels is the entry's pixel count, zero for a decline.
func (e textRasterEntry) pixels() int {
	if e.img == nil {
		return 0
	}

	return e.img.Bounds().Dx() * e.img.Bounds().Dy()
}

// textRasterCache keeps baked text rasters, least-recently-used first. A hit
// moves to the back, so a repaint that reuses its runs never evicts them.
type textRasterCache struct {
	mu      sync.Mutex
	entries map[textRasterKey]textRasterEntry
	order   []textRasterKey
	pixels  int
}
