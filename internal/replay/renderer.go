package replay

// Renderer owns the caches retained while drawing one page. Its zero value
// is ready to draw. Use it only on the owning Ebiten loop, including Close.
// Different renderers do not evict each other's images or display plans.
type Renderer struct {
	rasters  rasterCache
	fonts    textCache
	images   imageCache
	textRuns textRasterCache
	chunks   chunkStore
	plans    planCache
	layers   layerStore
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
}
