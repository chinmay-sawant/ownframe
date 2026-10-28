package window

import (
	"context"
	"image"
	"os"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func newBenchScreen(width, height int) *dirtyScreen {
	return &dirtyScreen{
		fakeScreen: &fakeScreen{width: width, height: height, minW: 1, minH: 1},
		display:    partialDisplay(width, height),
		gen:        1,
	}
}

// BenchmarkTileSubImageStaticBounds measures SubImage calls with identical bounds.
// Ebiten hits its internal subImageCache, but still acquires subImageCacheM mutex
// and wraps *ebiten.Image in an image.Image interface value.
func BenchmarkTileSubImageStaticBounds(b *testing.B) {
	img := ebiten.NewImage(256, 256)
	defer img.Deallocate()
	r := image.Rect(10, 10, 100, 100)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub, ok := img.SubImage(r).(*ebiten.Image)
		if !ok || sub == nil {
			b.Fatal("nil sub")
		}
	}
}

// BenchmarkTileSubImageUniqueBounds simulates continuous scrolling or animated
// damage rects where the sub-region rectangle changes every frame.
// Every new rect misses subImageCache, allocating &Image{} and map buckets on the heap.
func BenchmarkTileSubImageUniqueBounds(b *testing.B) {
	img := ebiten.NewImage(2048, 2048)
	defer img.Deallocate()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 1000
		r := image.Rect(x, 0, x+256, 256)
		sub, ok := img.SubImage(r).(*ebiten.Image)
		if !ok || sub == nil {
			b.Fatal("nil sub")
		}
	}
}

// BenchmarkTileDirectBounds measures direct drawing using existing image bounds
// without any SubImage slicing or mutex locks.
func BenchmarkTileDirectBounds(b *testing.B) {
	src := ebiten.NewImage(256, 256)
	defer src.Deallocate()
	dst := ebiten.NewImage(800, 600)
	defer dst.Deallocate()

	var op ebiten.DrawImageOptions
	op.Blend = ebiten.BlendCopy

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst.DrawImage(src, &op)
	}
}

// BenchmarkTilePaintFull measures painting all tiles in a 800x600 window (12 tiles).
func BenchmarkTilePaintFull(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	s.ensureTiles(800, 600)
	defer s.dropTiles()

	full := image.Rect(0, 0, 800, 600)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.paintTiles(screen.display, full)
	}
}

// BenchmarkTilePaintPartial measures repainting a single 50x50 dirty rect in one tile.
func BenchmarkTilePaintPartial(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	s.ensureTiles(800, 600)
	defer s.dropTiles()

	r := image.Rect(10, 10, 60, 60)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.paintTiles(screen.display, r)
	}
}

// BenchmarkTileBlitFull measures blitting all 12 tiles across a 800x600 screen.
func BenchmarkTileBlitFull(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	s.ensureTiles(800, 600)
	defer s.dropTiles()
	dst := partialBuf(800, 600)
	defer dst.Deallocate()

	for row := 0; row < s.partial.tiles.rows; row++ {
		for col := 0; col < s.partial.tiles.cols; col++ {
			s.partial.tiles.markClean(col, row)
		}
	}

	full := image.Rect(0, 0, 800, 600)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.blitTiles(dst, screen.display, full)
	}
}

// BenchmarkTileBlitPartial measures blitting a single dirty tile rect to the destination.
func BenchmarkTileBlitPartial(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	s.ensureTiles(800, 600)
	defer s.dropTiles()
	dst := partialBuf(800, 600)
	defer dst.Deallocate()

	s.partial.tiles.markClean(0, 0)
	r := image.Rect(10, 10, 60, 60)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.blitTiles(dst, screen.display, r)
	}
}

// BenchmarkTileDrawReplayPartialClean measures drawReplayPartial on an undamaged frame.
func BenchmarkTileDrawReplayPartialClean(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	dst := partialBuf(800, 600)
	defer dst.Deallocate()

	// Warm up first frame (rebuilds buffer)
	s.drawReplayPartial(dst)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.drawReplayPartial(dst)
	}
}

// BenchmarkTileDrawReplayPartialDamaged measures drawReplayPartial when a single dirty rect is flagged.
func BenchmarkTileDrawReplayPartialDamaged(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newDirtyShell(screen)
	dst := partialBuf(800, 600)
	defer dst.Deallocate()

	// Warm up first frame
	s.drawReplayPartial(dst)

	damage := image.Rect(20, 20, 50, 50)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		screen.click(damage)
		s.drawReplayPartial(dst)
	}
}

// BenchmarkTileUpdateClean measures shell.Update() on a clean frame (s.perf = false).
func BenchmarkTileUpdateClean(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newShell(context.Background(), screen)
	s.display = screen.display
	s.seq = screen.gen
	s.perf = false
	s.lastPoll = time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.lastPoll = time.Now()
		if err := s.Update(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTileUpdateWithPerf measures shell.Update() with s.perf = true.
func BenchmarkTileUpdateWithPerf(b *testing.B) {
	screen := newBenchScreen(800, 600)
	s := newShell(context.Background(), screen)
	s.display = screen.display
	s.seq = screen.gen
	s.perf = true
	s.lastPoll = time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.lastPoll = time.Now()
		if err := s.Update(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTileUpdateDamaged measures shell.Update() when screen reports ticking and damage.
func BenchmarkTileUpdateDamaged(b *testing.B) {
	screen := newBenchScreen(800, 600)
	screen.ticking = true
	s := newShell(context.Background(), screen)
	s.display = screen.display
	s.seq = screen.gen
	s.perf = false
	s.lastPoll = time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.lastPoll = time.Now()
		screen.frame = true
		if err := s.Update(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTileReloadOSStat measures os.Stat executing on the 60 Hz thread
// as performed every 250ms by pollReload.
func BenchmarkTileReloadOSStat(b *testing.B) {
	tmp, err := os.CreateTemp("", "bench-reload-*.html")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Close(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := os.Stat(tmp.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}
