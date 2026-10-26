package replay

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// flushChunks empties the chunk cache so a test starts cold.
func flushChunks() {
	chunks.mu.Lock()
	defer chunks.mu.Unlock()

	for _, e := range chunks.entries {
		e.img.Deallocate()
	}

	chunks.entries = nil
	chunks.order = nil
	chunks.pixels = 0
}

// TestChunkCacheReusesAndRebuilds pins the two mutation paths: a signature
// change with the same box repaints the same image, and a box change builds
// a new one.
func TestChunkCacheReusesAndRebuilds(t *testing.T) {
	flushChunks()
	defer flushChunks()

	key := chunkKey{id: 7}
	box := image.Rect(0, 0, 32, 32)

	first := chunks.get(key, 1, box, func(*ebiten.Image) {})
	second := chunks.get(key, 2, box, func(*ebiten.Image) {})
	if second.img != first.img {
		t.Fatal("a signature change built a new image")
	}

	third := chunks.get(key, 2, image.Rect(0, 0, 48, 32), func(*ebiten.Image) {})
	if third.img == first.img {
		t.Fatal("a box change reused the old image")
	}

	if third.img.Bounds().Dx() != 48 {
		t.Fatalf("new image = %v", third.img.Bounds())
	}
}

// TestChunkCacheStaysBounded fills past the entry cap and requires eviction
// down to it.
func TestChunkCacheStaysBounded(t *testing.T) {
	flushChunks()
	defer flushChunks()

	box := image.Rect(0, 0, 16, 16)
	for i := range chunkMaxEntries + 8 {
		chunks.get(chunkKey{id: i}, uint64(i), box, func(*ebiten.Image) {})
	}

	if got := len(chunks.entries); got > chunkMaxEntries {
		t.Fatalf("entries = %d, want at most %d", got, chunkMaxEntries)
	}

	if chunks.pixels > chunkMaxPixels {
		t.Fatalf("pixels = %d, want at most %d", chunks.pixels, chunkMaxPixels)
	}
}

// TestChunkCacheEvictsLeastRecentlyUsed pins the recency policy: a hit marks
// the slot most-recently used, so the next eviction drops an untouched slot
// instead of the hot one.
func TestChunkCacheEvictsLeastRecentlyUsed(t *testing.T) {
	flushChunks()
	defer flushChunks()

	box := image.Rect(0, 0, 16, 16)
	for i := range chunkMaxEntries {
		chunks.get(chunkKey{id: i}, uint64(i), box, func(*ebiten.Image) {})
	}

	hot := chunkKey{id: 0}
	chunks.get(hot, 0, box, func(*ebiten.Image) {})
	chunks.get(chunkKey{id: chunkMaxEntries}, chunkMaxEntries, box, func(*ebiten.Image) {})

	if _, ok := chunks.entries[hot]; !ok {
		t.Fatal("recently used chunk 0 was evicted")
	}

	if _, ok := chunks.entries[chunkKey{id: 1}]; ok {
		t.Fatal("untouched chunk 1 survived ahead of recently used chunk 0")
	}
}
