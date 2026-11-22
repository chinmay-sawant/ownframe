package replay

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"go/parser"
	"go/token"
	"image"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestRenderersEvictAndCloseIndependently(t *testing.T) {
	var first, second Renderer
	defer first.Close()
	defer second.Close()
	target := imageOp(t, tinyPNG(t, 200), 2, 2, false)
	firstImage := first.decodedImage(&target)
	secondImage := second.decodedImage(&target)
	if firstImage == nil || secondImage == nil || firstImage == secondImage {
		t.Fatal("renderers did not retain independent images")
	}
	for i := range imageCacheLimit {
		filler := imageOp(t, tinyPNG(t, uint8(i)), 2, 2, false)
		if first.decodedImage(&filler) == nil {
			t.Fatal("filler did not decode")
		}
	}
	if first.images.get(imageBytesKey(&target)) != nil {
		t.Fatal("first renderer did not evict its oldest entry")
	}
	if second.decodedImage(&target) != secondImage {
		t.Fatal("first renderer evicted the second renderer's image")
	}

	display := &layout.Display{Width: 16, Height: 16}
	first.chunkPlanFor(display)
	second.chunkPlanFor(display)
	if first.plans.entries[0].root == second.plans.entries[0].root {
		t.Fatal("renderers shared a mutable paint plan")
	}
	box := image.Rect(0, 0, 2, 2)
	key := chunkKey{id: 1}
	first.chunks.get(key, 1, box, func(*ebiten.Image) {})
	retained := second.chunks.get(key, 1, box, func(*ebiten.Image) {})
	first.Close()
	first.Close()
	if len(first.images.entries) != 0 || len(first.chunks.entries) != 0 || first.plans.entries[0].root != nil {
		t.Fatal("Close retained page images or plans")
	}
	if second.decodedImage(&target) != secondImage || second.chunks.entries[key] != retained {
		t.Fatal("closing one renderer cleared another renderer")
	}
	if first.decodedImage(&target) == nil {
		t.Fatal("closed renderer could not be reused")
	}
}

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
	r.fonts.sources = map[any]*text.GoTextFaceSource{}
	r.fonts.mu.Unlock()

	r.SetIdle()

	r.fonts.mu.Lock()
	defer r.fonts.mu.Unlock()
	if r.fonts.sources == nil {
		t.Fatal("SetIdle must keep the font source cache")
	}
}

var (
	rasters  = &defaultRenderer.rasters
	fonts    = &defaultRenderer.fonts
	chunks   = &defaultRenderer.chunks
	textRuns = &defaultRenderer.textRuns
)

func chunkPlanFor(d *layout.Display) *chunkPlan {
	return defaultRenderer.chunkPlanFor(d)
}

func chunkRaster(d *layout.Display, p *chunkPlan, c *chunkRun) *chunkEntry {
	return defaultRenderer.chunkRaster(d, p, c)
}

func decodedImage(op *layout.DisplayOp) *ebiten.Image {
	return defaultRenderer.decodedImage(op)
}

func drawOp(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	defaultRenderer.drawOp(dst, op, dx, dy)
}

func drawTextRun(dst *ebiten.Image, face *text.GoTextFace, content string, x, y float64, op *layout.DisplayOp) {
	defaultRenderer.drawTextRun(dst, face, content, x, y, op)
}

func textFace(op *layout.DisplayOp) *text.GoTextFace {
	return defaultRenderer.textFace(op)
}

func TestPinZSeparatesScrollingAndViewportOperations(t *testing.T) {
	body := &layout.DisplayOp{ZIndex: 1, ZIndexSet: true}
	bar := &layout.DisplayOp{ZIndex: 2, ZIndexSet: true}
	fixed := &layout.DisplayOp{Fixed: true, ZIndex: 3, ZIndexSet: true}
	if !scrollContentOp(body, 2) || scrollContentOp(bar, 2) || scrollContentOp(fixed, 2) {
		t.Fatal("scroll buffer included a pinned or fixed operation")
	}
	if atPinZ(body, 2) || !atPinZ(bar, 2) {
		t.Fatal("viewport pin threshold selected the wrong layer")
	}
}

// TestScissorOverflowClipsTransformedPaint pins the composite-time clip: an
// op with a recorded overflow box draws only inside it, and a clip that
// erases the op skips the draw. Ops without a clip keep the full dst.
func TestScissorOverflowClipsTransformedPaint(t *testing.T) {
	dst := ebiten.NewImage(64, 64)

	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 48, H: 48}
	op.SetOverflowClip(6, 6, 24, 24)

	sub, paint := scissorOverflow(dst, &op, 0, 0)
	if !paint {
		t.Fatal("clipped op skipped")
	}

	if got := sub.Bounds(); got != image.Rect(8, 8, 40, 40) {
		t.Fatalf("scissor = %v, want (8,8)-(40,40)", got)
	}

	outside := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8}
	outside.SetOverflowClip(48, 48, 8, 8)

	if _, paint := scissorOverflow(dst, &outside, 0, 0); paint {
		t.Fatal("fully clipped op not skipped")
	}

	plain := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8}
	if _, _, _, _, ok := plain.OverflowClip(); ok {
		t.Fatal("plain op reports a clip")
	}

	if sub, paint := scissorOverflow(dst, &plain, 0, 0); !paint || sub.Bounds() != dst.Bounds() {
		t.Fatalf("plain op dst changed: paint=%v bounds=%v", paint, sub.Bounds())
	}
}

func TestVisiblePreservesOrderAndUnboundedInk(t *testing.T) {
	d := &layout.Display{
		PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 20, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 1500, W: 20, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 750, W: 20, H: 20},
			{Kind: layout.DisplayOpText, Y: 9000, RotateDeg: 20},
		},
		Order: []int{3, 2, 1, 0, -1, 99},
	}
	rect := visibleRect(image.Rect(0, 0, 100, 100), 0, -1000)
	got := []int{}
	visitVisible(d, rect, func(i int) { got = append(got, i) })
	if !reflect.DeepEqual(got, []int{3, 2}) {
		t.Fatalf("visible paint order = %v, want [3 2]", got)
	}
}

func TestVisibleRectRoundsFractionalScrollOutward(t *testing.T) {
	got := visibleRect(image.Rect(5, 10, 105, 110), -0.5, -20.25)
	if want := image.Rect(5, 30, 106, 131); got != want {
		t.Fatalf("rect = %v, want %v", got, want)
	}
}

const replayBannedRender = "github.com/chinmay-sawant/ownframe/internal/render"

var replayRenderAllow = map[string]bool{}

func TestReplayRenderGate(t *testing.T) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller unknown")
	}
	dir := filepath.Dir(self)
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ".go") {
			continue
		}
		if strings.HasSuffix(n, "_test.go") {
			continue
		}
		if replayRenderAllow[n] {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, n), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if p == replayBannedRender {
				t.Errorf("%s imports %s; keep replay below render", n, p)
			}
		}
	}
}

// TestCanvasUnitsUsePointsPerPixel locks the conversion the op filter relies
// on: a canvas of display.Width CSS pixels is display.Width * PointsPerPixel
// wide in op points.
func TestCanvasUnitsUsePointsPerPixel(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<html><body style="margin:0;background:#ffffff;height:100px"></body></html>`,
		320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	wantPts := float64(display.Width) * display.PointsPerPixel

	var found *layout.DisplayOp
	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && op.X == 0 && op.Y == 0 {
			found = op
		}
	}

	if found == nil {
		t.Fatal("no canvas fill")
	}

	if math.Abs(found.W-wantPts) > 0.5 {
		t.Fatalf("fill W = %v points, want %v", found.W, wantPts)
	}

	// In canvas pixels the fill must span the whole canvas.
	box, ok := opBounds(found, display.PointsPerPixel)
	if !ok {
		t.Fatal("fill unbounded")
	}

	if box.Dx() < display.Width {
		t.Fatalf("bounds %v are not in canvas pixels, want width %d", box, display.Width)
	}
}

func TestRendererCloseAndIdleReleaseBackdropState(t *testing.T) {
	t.Parallel()

	var r Renderer
	s0, s1, s2, ok := r.backdrop.acquireScratch(32, 24)
	if !ok || s0 == nil || s1 == nil || s2 == nil {
		t.Fatal("acquireScratch failed")
	}
	// Smaller request reuses the same pooled scratch images.
	r0, _, _, ok := r.backdrop.acquireScratch(16, 12)
	if !ok || r0 == nil || r.backdrop.size != image.Pt(32, 24) {
		t.Fatalf("smaller acquireScratch resized pool: size=%v", r.backdrop.size)
	}

	r.SetIdle()
	if r.backdrop.scratch[0] != nil || r.backdrop.size != (image.Point{}) {
		t.Fatal("SetIdle did not release backdrop scratch pool")
	}

	// Simulate initialized shader state and verify Close resets everything.
	r.backdrop.acquireScratch(16, 16)
	r.backdrop.mu.Lock()
	r.backdrop.inited = true
	r.backdrop.mu.Unlock()

	r.Close()
	if r.backdrop.inited || r.backdrop.shader != nil || r.backdrop.scratch[0] != nil || r.backdrop.size != (image.Point{}) {
		t.Fatal("Close did not reset backdrop shader and scratch state")
	}
}
