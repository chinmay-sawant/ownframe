package window

import (
	"context"
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"os"
	"testing"
	"time"
)

func TestPerfIsOwnedByItsWindow(t *testing.T) {
	a := newDevShell(newDevScreen())
	a.perf = true
	a.wirePerf()
	before := devFrameText(a.devFrameRows(320))
	b := newDevShell(newDevScreen())
	b.wirePerf()
	if got := devFrameText(a.devFrameRows(320)); got != before {
		t.Fatal("another window replaced this window's performance hooks")
	}
}

func TestPerfRingMath(t *testing.T) {
	t.Parallel()
	var r perfRing
	if r.avg() != 0 || r.p95() != 0 || r.p99() != 0 || r.longFrames(perfLongThreshold) != 0 {
		t.Fatal("empty ring must read zero")
	}
	for i := 0; i < 10; i++ {
		r.record(10 * time.Millisecond)
	}
	if got := r.avg(); got != 10*time.Millisecond {
		t.Fatalf("avg = %v, want 10ms", got)
	}
	if got := r.p95(); got != 10*time.Millisecond {
		t.Fatalf("p95 = %v, want 10ms", got)
	}
	if got := r.p99(); got != 10*time.Millisecond {
		t.Fatalf("p99 = %v, want 10ms", got)
	}
	r.record(100 * time.Millisecond)
	if got := r.longFrames(perfLongThreshold); got != 1 {
		t.Fatalf("long = %d, want 1", got)
	}
	if got := r.p95(); got <= 10*time.Millisecond {
		t.Fatalf("p95 = %v, want above 10ms", got)
	}
	for i := 0; i < 200; i++ {
		r.record(time.Millisecond)
	}
	if got := r.count(); got != perfRingCap {
		t.Fatalf("count = %d, want %d", got, perfRingCap)
	}
}

// TestPerfRingSummaryNoAlloc pins the point of the scratch buffer: a summary
// on a full ring must not allocate, because the frame loop runs it.
func TestPerfRingSummaryNoAlloc(t *testing.T) {
	var r perfRing
	for i := range perfRingCap {
		r.record(time.Duration(i) * time.Millisecond)
	}

	allocs := testing.AllocsPerRun(100, func() {
		avg, p95, p99, long := r.summary()
		if avg < 0 || p95 < 0 || p99 < 0 || long < 0 {
			t.Fatal("negative summary")
		}
	})

	if allocs != 0 {
		t.Fatalf("summary allocates %.1f times per call, want 0", allocs)
	}
}

func TestPerfSamplerValues(t *testing.T) {
	t.Parallel()
	var p perfSampler
	if !p.maybe(time.Now()) {
		t.Fatal("first sample must run")
	}
	if p.maybe(time.Now()) {
		t.Fatal("resample must gate")
	}
	if p.goroutines < 1 {
		t.Fatalf("goroutines = %d, want >= 1", p.goroutines)
	}
	s := &shell{perf: true}
	s.perfSampleRuntime()
	_, rss, _, g := s.perfRuntime()
	if g < 1 {
		t.Fatalf("shell goroutines = %d, want >= 1", g)
	}
	if got := s.dev.sampler.rss; got != rss {
		t.Fatalf("rss = %d, want %d", rss, got)
	}

	off := &shell{}
	off.perfSampleRuntime()
	off.perfFrameSample()
	if a, p95, p99, l := off.perfSummary(); a != 0 || p95 != 0 || p99 != 0 || l != 0 {
		t.Fatal("perf off must not sample")
	}
}

func TestPerfStageZeroSafe(t *testing.T) {
	t.Parallel()
	var st perfStage
	if st != (perfStage{}) {
		t.Fatal("stage must start zero")
	}
	perfTimed(&st.draw, func() {})
	if st.draw < 0 {
		t.Fatalf("draw = %v, want >= 0", st.draw)
	}
	want := errors.New("tick boom")
	if err := perfTimedErr(&st.tick, func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("timed err = %v, want %v", err, want)
	}
	var s shell
	if a, p95, p99, l := s.perfSummary(); a != 0 || p95 != 0 || p99 != 0 || l != 0 {
		t.Fatal("empty summary must read zero")
	}
}

func TestPerfFrameCostExcludesParkedGap(t *testing.T) {
	s := &shell{perf: true}
	s.dev.frame = 125 * time.Millisecond
	s.dev.stages.pendingUpdate = 2 * time.Millisecond
	s.dev.stages.draw = 3 * time.Millisecond
	s.dev.stages.present = time.Millisecond
	s.perfFrameSample()
	if s.dev.stages.work != 6*time.Millisecond {
		t.Fatalf("last frame work = %s, want 6ms", s.dev.stages.work)
	}
	avg, _, p99, long := s.perfSummary()
	if avg != 6*time.Millisecond || p99 != 6*time.Millisecond || long != 0 {
		t.Fatalf("work avg=%s p99=%s long=%d", avg, p99, long)
	}
	if s.dev.stages.pendingUpdate != 0 {
		t.Fatal("update work counted twice")
	}
	s.dev.stages.draw = 20 * time.Millisecond
	s.perfFrameSample()
	if _, _, _, long := s.perfSummary(); long != 1 {
		t.Fatal("20ms draw must exceed the 60Hz budget")
	}
}

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
