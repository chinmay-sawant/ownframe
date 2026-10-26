package replay

import (
	"runtime/debug"

	"github.com/hajimehoshi/ebiten/v2"
)

// idleWarmCaps bound the containers Wake pre-sizes after a park. They cover
// one frame of repaints without regrowing, not the full cache budgets, so a
// wake costs kilobytes and the next paints regrow lazily.
const (
	idleWarmRasters  = 16
	idleWarmTextRuns = 64
	idleWarmImages   = 8
	idleWarmChunks   = 16
	idleWarmLayers   = 4
)

// SetIdle releases the renderer's GPU-backed pixel caches when the window
// parks, then returns the freed pages to the OS. Parked keep-alive frames
// only blit the window's tile buffers and never touch these caches, so the
// next dirty paint re-bakes lazily. The font parse cache stays: its memory
// is bounded by the face caps and re-parsing sources on wake would cost more
// than the pages it holds. Call on the owning Ebiten loop, like Close. It is
// idempotent: a renderer that already idles releases nothing further.
func (renderer *Renderer) SetIdle() {
	renderer.rasters.releaseIdle()
	renderer.textRuns.releaseIdle()
	renderer.images.releaseIdle()
	renderer.chunks.releaseIdle()
	renderer.plans.releaseIdle()
	renderer.layers.releaseIdle()

	// FreeOSMemory scavenges the pages the cleared maps and order slices
	// freed, which is the Go runtime's MADV_DONTNEED path: without it the
	// allocator keeps the pages mapped and resident set size does not move.
	debug.FreeOSMemory()
}

// Wake pre-sizes the pixel-cache containers after a park so the first
// post-wake paints do not regrow maps and order slices under the draw lock.
// It allocates kilobytes of Go heap and no GPU images, and it never clobbers
// a cache that already holds entries. Call on the owning Ebiten loop.
func (renderer *Renderer) Wake() {
	renderer.rasters.warmIdle(idleWarmRasters)
	renderer.textRuns.warmIdle(idleWarmTextRuns)
	renderer.images.warmIdle(idleWarmImages)
	renderer.chunks.warmIdle(idleWarmChunks)
	renderer.layers.warmIdle(idleWarmLayers)
}

// releaseIdle deallocates every baked raster and forgets the index, so the
// pages return to the allocator before FreeOSMemory hands them to the OS.
func (c *rasterCache) releaseIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.entries {
		img.Deallocate()
	}

	c.entries = nil
	c.order = nil
	c.pixels = 0
}

// warmIdle pre-sizes the index when a park emptied it. A live cache keeps
// its entries and capacity.
func (c *rasterCache) warmIdle(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[rasterKey]*ebiten.Image, n)
	}
	if c.order == nil {
		c.order = make([]rasterKey, 0, n)
	}
}

// releaseIdle drops every baked run and remembered decline. It reuses the
// benchmark flush, which already deallocates and forgets the index.
func (c *textRasterCache) releaseIdle() {
	c.flush()
}

// warmIdle pre-sizes the index when a park emptied it. A live cache keeps
// its entries and capacity.
func (c *textRasterCache) warmIdle(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[textRasterKey]textRasterEntry, n)
	}
	if c.order == nil {
		c.order = make([]textRasterKey, 0, n)
	}
}

// releaseIdle deallocates every decoded image and forgets the index, so the
// upload memory returns to the allocator before FreeOSMemory.
func (c *imageCache) releaseIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.entries {
		img.Deallocate()
	}

	c.entries = nil
	c.sizes = nil
	c.order = nil
	c.total = 0
}

// warmIdle pre-sizes the index when a park emptied it. A live cache keeps
// its entries and capacity.
func (c *imageCache) warmIdle(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[uint64]*ebiten.Image, n)
		c.sizes = make(map[uint64]int, n)
	}
	if c.order == nil {
		c.order = make([]uint64, 0, n)
	}
}

// releaseIdle deallocates every baked chunk and forgets the index.
func (c *chunkStore) releaseIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, entry := range c.entries {
		entry.img.Deallocate()
	}

	c.entries = nil
	c.order = nil
	c.pixels = 0
}

// warmIdle pre-sizes the index when a park emptied it. A live cache keeps
// its entries and capacity.
func (c *chunkStore) warmIdle(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[chunkKey]*chunkEntry, n)
	}
	if c.order == nil {
		c.order = make([]chunkKey, 0, n)
	}
}

// releaseIdle disposes both plan slots and forgets the displays they came
// from. A display is rebuilt wholesale on redraw, so nothing here survives
// the next paint anyway.
func (c *planCache) releaseIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.entries {
		disposePlan(c.entries[i].root)
		c.entries[i] = planEntry{}
	}
	c.next = 0
}

// releaseIdle deallocates every retained layer texture and forgets the slots
// without reporting damage: the parked window blits its tiles and the next
// BeginLayers rebuilds from the frame reset. Cumulative counters keep their
// totals; only the live set reads empty.
func (store *layerStore) releaseIdle() {
	store.mu.Lock()
	defer store.mu.Unlock()

	for key, entry := range store.entries {
		entry.Tex.Deallocate()
		entry.Valid = false
		delete(store.entries, key)
	}
	store.entries = nil
	store.pixels = 0
	clear(store.churn)
	store.frame.reset()
	store.stats.Active = 0
	store.stats.Pixels = 0
	store.stats.Bytes = 0
}

// warmIdle pre-sizes the layer index when a park emptied it. A live set
// keeps its textures.
func (store *layerStore) warmIdle(n int) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.entries == nil {
		store.entries = make(map[layerKey]*Layer, n)
	}
}
