package replay

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	etext "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TestRendererIdleReleasesPixelCaches pins the park contract: SetIdle
// empties every GPU-backed pixel cache, Wake pre-sizes the containers
// without allocating images, and the next paints re-bake from a clean index.
func TestRendererIdleReleasesPixelCaches(t *testing.T) {
	var r Renderer
	box := image.Rect(0, 0, 32, 32)

	r.chunks.get(chunkKey{id: 1}, 1, box, func(*ebiten.Image) {})
	r.textRuns.raster(textRasterKey{}, func() (textRasterEntry, bool) {
		return textRasterEntry{img: ebiten.NewImage(16, 16)}, true
	})
	r.images.put(1, image.NewRGBA(image.Rect(0, 0, 8, 8)), 64)
	r.rasters.raster(rasterKey{variant: rasterFill, w: 8, h: 8}, image.Pt(12, 12), func(*ebiten.Image) {})

	r.layers.mu.Lock()
	r.layers.entries = map[layerKey]*Layer{
		{id: "a", box: 1}: {Tex: ebiten.NewImage(8, 8), Valid: true},
	}
	r.layers.pixels = 64
	r.layers.mu.Unlock()

	r.SetIdle()

	if len(r.chunks.entries) != 0 || r.chunks.pixels != 0 || r.chunks.order != nil {
		t.Fatal("SetIdle must empty the chunk cache")
	}
	if len(r.textRuns.entries) != 0 || r.textRuns.pixels != 0 {
		t.Fatal("SetIdle must empty the text run cache")
	}
	if len(r.images.entries) != 0 || r.images.total != 0 {
		t.Fatal("SetIdle must empty the image cache")
	}
	if len(r.rasters.entries) != 0 || r.rasters.pixels != 0 {
		t.Fatal("SetIdle must empty the raster cache")
	}
	r.layers.mu.Lock()
	layered := len(r.layers.entries) != 0 || r.layers.pixels != 0
	r.layers.mu.Unlock()
	if layered {
		t.Fatal("SetIdle must empty the layer store")
	}

	r.Wake()

	if r.chunks.entries == nil || r.textRuns.entries == nil || r.images.entries == nil ||
		r.rasters.entries == nil {
		t.Fatal("Wake must pre-size the pixel-cache indexes")
	}

	got := r.chunks.get(chunkKey{id: 1}, 1, box, func(*ebiten.Image) {})
	if got == nil || got.img == nil {
		t.Fatal("a post-wake paint must re-bake into the clean index")
	}
}

// TestRendererIdleKeepsFonts pins that a park keeps the parsed face cache:
// re-parsing sources on wake would cost more than the bounded pages they
// hold.
func TestRendererIdleKeepsFonts(t *testing.T) {
	var r Renderer
	r.fonts.mu.Lock()
	r.fonts.sources = map[any]*etext.GoTextFaceSource{}
	r.fonts.mu.Unlock()

	r.SetIdle()

	r.fonts.mu.Lock()
	defer r.fonts.mu.Unlock()
	if r.fonts.sources == nil {
		t.Fatal("SetIdle must keep the font source cache")
	}
}
