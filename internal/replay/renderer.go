package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"runtime/debug"
	"sync"
)

// Renderer owns the caches retained while drawing one page. Its zero value
// is ready to draw. Use it only on the owning Ebiten loop, including Close.
// Different renderers do not evict each other's images or display plans.
type Renderer struct {
	rasters  rasterCache
	corners  cornerCache
	fonts    textCache
	images   imageCache
	textRuns textRasterCache
	chunks   chunkStore
	plans    planCache
	layers   layerStore
	drawOpts ebiten.DrawImageOptions
	// backdropOnce guards the compiled backdrop blur program, which lives
	// with the renderer so pages without backdrop-filter never pay the
	// compile cost and no process-wide mutable state is needed.
	backdropOnce   sync.Once
	backdropShader *ebiten.Shader
	backdropErr    error
}

// Close releases retained GPU images and forgets cached plans and fonts.
// It is idempotent. A subsequent draw can reuse the renderer with empty caches.
func (renderer *Renderer) Close() {
	renderer.rasters.mu.Lock()
	for _, img := range renderer.rasters.entries {
		img.Deallocate()
	}
	clear(renderer.rasters.entries)
	renderer.rasters.order = nil
	renderer.rasters.pixels = 0
	renderer.rasters.mu.Unlock()

	renderer.corners.close()

	renderer.textRuns.flush()

	renderer.images.mu.Lock()
	for _, img := range renderer.images.entries {
		img.Deallocate()
	}
	clear(renderer.images.entries)
	clear(renderer.images.sizes)
	renderer.images.order = nil
	renderer.images.total = 0
	renderer.images.mu.Unlock()

	renderer.chunks.mu.Lock()
	for _, entry := range renderer.chunks.entries {
		entry.img.Deallocate()
	}
	clear(renderer.chunks.entries)
	renderer.chunks.order = nil
	renderer.chunks.pixels = 0
	renderer.chunks.mu.Unlock()

	renderer.plans.mu.Lock()
	for i := range renderer.plans.entries {
		disposePlan(renderer.plans.entries[i].root)
		renderer.plans.entries[i] = planEntry{}
	}
	renderer.plans.next = 0
	renderer.plans.mu.Unlock()

	renderer.layers.close()

	renderer.fonts.mu.Lock()
	clear(renderer.fonts.faces)
	clear(renderer.fonts.sources)
	renderer.fonts.mu.Unlock()

	renderer.drawOpts.GeoM.Reset()
	renderer.drawOpts.ColorScale.Reset()
}

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
	renderer.corners.releaseIdle()
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
	renderer.corners.warmIdle(idleWarmRasters)
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

// Package functions retain the shared default for callers without an owner.
// Window hosts use their own Renderer instead.
func DrawRect(dst *ebiten.Image, display *layout.Display, rect image.Rectangle, dx, dy float64) {
	defaultRenderer.DrawRect(dst, display, rect, dx, dy)
}

func DrawUnfixedZ(dst *ebiten.Image, d *layout.Display, z int) {
	defaultRenderer.DrawUnfixedZ(dst, d, z)
}

func DrawRectUnfixedZ(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
	defaultRenderer.DrawRectUnfixedZ(dst, d, rect, z)
}

func DrawVisibleUnfixedZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	defaultRenderer.DrawVisibleUnfixedZ(dst, d, dx, dy, z)
}

func DrawPinnedZ(dst *ebiten.Image, d *layout.Display, z int, dx, dy float64) {
	defaultRenderer.DrawPinnedZ(dst, d, z, dx, dy)
}

func Draw(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.Draw(dst, display, dx, dy)
}

func DrawUnfixed(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.DrawUnfixed(dst, display, dx, dy)
}

func DrawFixed(dst *ebiten.Image, display *layout.Display) {
	defaultRenderer.DrawFixed(dst, display)
}

func DrawVisible(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.DrawVisible(dst, display, dx, dy)
}

func DrawVisiblePinned(dst *ebiten.Image, display *layout.Display, dx, dy float64, pinZ int, pinDx, pinDy float64) {
	defaultRenderer.DrawVisiblePinned(dst, display, dx, dy, pinZ, pinDx, pinDy)
}

func (renderer *Renderer) DrawUnfixedZ(dst *ebiten.Image, d *layout.Display, z int) {
	renderer.paintChunked(dst, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, scrollContentOp(op, z)
	})
}

func (renderer *Renderer) DrawRectUnfixedZ(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
	if dst == nil || d == nil {
		return
	}

	rect = clipRect(d, rect).Intersect(dst.Bounds())
	if rect.Empty() {
		return
	}

	clip, ok := dst.SubImage(rect).(*ebiten.Image)
	if !ok {
		return
	}

	ppt := d.PointsPerPixel
	renderer.paintChunked(clip, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, scrollContentOp(op, z) && opTouches(op, ppt, rect)
	})
}

func (renderer *Renderer) DrawVisibleUnfixedZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	renderer.visitVisibleZ(dst, d, dx, dy, z, false)
}

func (renderer *Renderer) DrawPinnedZ(dst *ebiten.Image, d *layout.Display, z int, dx, dy float64) {
	renderer.visitVisibleZ(dst, d, dx, dy, z, true)
}

func (renderer *Renderer) visitVisibleZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int, pinned bool) {
	if dst == nil || d == nil {
		return
	}

	r := visibleRect(dst.Bounds(), dx, dy)
	ppt := d.PointsPerPixel

	renderer.paintChunked(dst, d, func(op *layout.DisplayOp) (float64, float64, bool) {
		if op.Fixed || atPinZ(op, z) != pinned || !opTouches(op, ppt, r) {
			return 0, 0, false
		}

		return dx, dy, true
	})
}

func atPinZ(op *layout.DisplayOp, z int) bool {
	return z > 0 && op.ZIndexSet && op.ZIndex >= z
}

func scrollContentOp(op *layout.DisplayOp, z int) bool {
	return !op.Fixed && !atPinZ(op, z)
}
