package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"testing"
	"time"
)

func TestWantFrameAtRoundTrip(t *testing.T) {
	p := &Page{}
	at := time.Unix(123, 456)
	p.WantFrameAt(at)

	if got := p.NextFrameAt(); !got.Equal(at) {
		t.Fatalf("NextFrameAt = %v, want %v", got, at)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a second read = %v, want the zero time", got)
	}
}

func TestWantFrameInIsFuture(t *testing.T) {
	p := &Page{}
	p.WantFrameIn(50 * time.Millisecond)

	d := time.Until(p.NextFrameAt())
	if d <= 0 || d > 50*time.Millisecond {
		t.Fatalf("deadline %v out of range", d)
	}
}

func TestWantFrameAtZeroClears(t *testing.T) {
	p := &Page{}
	p.WantFrameAt(time.Unix(1, 0))
	p.WantFrameAt(time.Time{})

	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("NextFrameAt = %v, want the zero time", got)
	}
}

func TestStaticPageFrameEvents(t *testing.T) {
	p := tickBarPage(t)
	events := p.FrameEvents()
	select {
	case <-events:
	default:
	}
	if p.FramePolling() {
		t.Fatal("static page requested polling")
	}
	p.WantFrameIn(time.Second)
	select {
	case <-events:
	default:
		t.Fatal("external deadline did not notify host")
	}
	if p.NextFrameAt().IsZero() {
		t.Fatal("notification consumed deadline")
	}
	p.SetTick(func(context.Context) error { p.WantFrameIn(time.Second); return nil })
	if !p.FramePolling() {
		t.Fatal("legacy callback lost keep-alive")
	}
	select {
	case <-events:
	default:
	}
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-events:
		t.Fatal("tick notified itself instead of scheduling deadline")
	default:
	}
	p.SetTick(nil)
	if p.FramePolling() {
		t.Fatal("removed callback still polls")
	}
}

func TestCaretSuppliesIdleDeadline(t *testing.T) {
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	setCaret(p, 2, false)
	at := time.Now()
	formState(p).BlinkAt = at
	if p.FramePolling() {
		t.Fatal("caret unnecessarily enabled compatibility polling")
	}
	if got := p.NextFrameAt(); !got.Equal(at.Add(caretBlinkInterval)) {
		t.Fatalf("caret deadline = %v, want %v", got, at.Add(caretBlinkInterval))
	}
	earlier := at.Add(time.Millisecond)
	p.WantFrameAt(earlier)
	if got := p.NextFrameAt(); !got.Equal(earlier) {
		t.Fatalf("caret delayed earlier request: %v", got)
	}
	formState(p).All = true
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("selection scheduled a caret: %v", got)
	}
}

func tickBarPage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{
		HTML:   `<body style="margin:0"><div id="bar" style="width:40px;height:12px;background:#224466"></div></body>`,
		Width:  80,
		Height: 40,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
	if p.FrameDirty() {
		t.Fatal("Redraw left FrameDirty set")
	}

	return p
}

func TestTickFillPatchSetsFrameDirty(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	patched := false
	p.SetTick(func(context.Context) error {
		d := p.Display()
		if d == nil {
			return nil
		}
		bar := boxRect(boxByID(p, "bar"))
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind != DisplayOpFillRect || op.W <= 1 || op.H <= 1 {
				continue
			}
			if opBounds(op, d).Overlaps(bar) {
				op.R = 1
				patched = true
				break
			}
		}

		return nil
	})
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !patched {
		t.Fatal("tick did not patch a fill")
	}
	if !p.FrameDirty() {
		t.Fatal("patched fill must set FrameDirty")
	}
	r, ok := p.TakeDirty()
	if !ok || r.Empty() {
		t.Fatalf("TakeDirty %v ok=%v", r, ok)
	}
	if r.Intersect(boxRect(boxByID(p, "bar"))).Empty() {
		t.Fatalf("dirty %v misses bar", r)
	}
}

func TestTickNoOpStaysClean(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	p.SetTick(func(context.Context) error { return nil })
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.FrameDirty() {
		t.Fatal("no-op tick set FrameDirty")
	}
	if _, ok := p.TakeDirty(); ok {
		t.Fatal("no-op tick marked dirty")
	}
}

func TestUseFrameDirtyStillSetsFlag(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	p.SetTick(func(context.Context) error {
		p.UseFrameDirty()

		return nil
	})
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.FrameDirty() {
		t.Fatal("UseFrameDirty alias did not set the flag")
	}
	if _, ok := p.TakeDirty(); ok {
		t.Fatal("alias-only tick marked a rect")
	}
}

func TestTickRunsCallback(t *testing.T) {
	t.Parallel()

	var got context.Context

	ctx := context.Background()
	p := &Page{}
	p.SetTick(func(c context.Context) error {
		got = c

		return nil
	})

	if err := p.Tick(ctx); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	if got != ctx {
		t.Fatal("the callback did not receive the context")
	}
}

func TestTickWithoutCallback(t *testing.T) {
	t.Parallel()

	if err := (&Page{}).Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}
}

func TestTickNilRemovesCallback(t *testing.T) {
	t.Parallel()

	calls := 0
	p := &Page{}
	p.SetTick(func(context.Context) error {
		calls++

		return nil
	})
	p.SetTick(nil)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	if calls != 0 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestTickReturnsCallbackError(t *testing.T) {
	t.Parallel()

	want := errors.New("stop")
	p := &Page{}
	p.SetTick(func(context.Context) error { return want })

	if err := p.Tick(context.Background()); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

// makeBenchDisplay creates a Display with n layout.DisplayOpFillRect ops.
func makeBenchDisplay(n int) *layout.Display {
	ops := make([]DisplayOp, n)
	for i := 0; i < n; i++ {
		ops[i] = DisplayOp{
			Kind:  layout.DisplayOpFillRect,
			X:     float64((i % 40) * 20),
			Y:     float64((i / 40) * 20),
			W:     15,
			H:     15,
			R:     0.2,
			G:     0.4,
			B:     0.6,
			Alpha: 1.0,
		}
	}

	return &layout.Display{
		Width:         1000,
		Height:        1000,
		PixelPerPoint: 1,
		Ops:           ops,
	}
}

// makeBenchPage creates a Page populated with a Display having n ops.
func makeBenchPage(n int) *Page {
	p := &Page{
		pageGeom: pageGeom{width: 1000, height: 1000},
	}
	p.draw.Display = makeBenchDisplay(n)

	return p
}

func benchTickCaptureAndDamage(b *testing.B, opCount int, modify bool) {
	b.Helper()
	p := makeBenchPage(opCount)
	gen := p.Generation()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		if modify {
			p.tickMutated.Store(true)
			snap := captureTickOps(p)
			p.draw.Display.Ops[0].R = float64(i%2) * 0.8
			noteTickDamage(p, gen, snap)
			freeOps(p, snap)
		} else if p.tickMutated.Load() {
			snap := captureTickOps(p)
			noteTickDamage(p, gen, snap)
			freeOps(p, snap)
		}
		p.dirty.Rects = p.dirty.Rects[:0]
	}
}

func benchTickDiffByPosition(b *testing.B, opCount int) {
	b.Helper()
	prev := makeBenchDisplay(opCount)
	next := makeBenchDisplay(opCount + 1)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		diffDisplay(prev, next)
	}
}

func BenchmarkTick_CaptureAndDamage_100Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 100, false)
}

func BenchmarkTick_CaptureAndDamage_100Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 100, true)
}

func BenchmarkTick_CaptureAndDamage_500Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 500, false)
}

func BenchmarkTick_CaptureAndDamage_500Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 500, true)
}

func BenchmarkTick_CaptureAndDamage_2000Ops_Mod0(b *testing.B) {
	benchTickCaptureAndDamage(b, 2000, false)
}

func BenchmarkTick_CaptureAndDamage_2000Ops_Mod1(b *testing.B) {
	benchTickCaptureAndDamage(b, 2000, true)
}

func BenchmarkTick_DiffByPosition_100Ops(b *testing.B) {
	benchTickDiffByPosition(b, 100)
}

func BenchmarkTick_DiffByPosition_500Ops(b *testing.B) {
	benchTickDiffByPosition(b, 500)
}

func BenchmarkTick_DiffByPosition_2000Ops(b *testing.B) {
	benchTickDiffByPosition(b, 2000)
}

func benchFullTick(b *testing.B, opCount int, modify bool) {
	b.Helper()
	p := makeBenchPage(opCount)
	ctx := context.Background()

	i := 0
	p.SetTick(func(ctx context.Context) error {
		if modify {
			p.Display().Ops[0].R = float64(i%2) * 0.8
		}
		return nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		i++
		if err := p.Tick(ctx); err != nil {
			b.Fatal(err)
		}
		p.dirty.Rects = p.dirty.Rects[:0]
	}
}

func BenchmarkTick_FullTick_100Ops_Mod0(b *testing.B) {
	benchFullTick(b, 100, false)
}

func BenchmarkTick_FullTick_100Ops_Mod1(b *testing.B) {
	benchFullTick(b, 100, true)
}

func BenchmarkTick_FullTick_500Ops_Mod0(b *testing.B) {
	benchFullTick(b, 500, false)
}

func BenchmarkTick_FullTick_500Ops_Mod1(b *testing.B) {
	benchFullTick(b, 500, true)
}

func BenchmarkTick_FullTick_2000Ops_Mod0(b *testing.B) {
	benchFullTick(b, 2000, false)
}

func BenchmarkTick_FullTick_2000Ops_Mod1(b *testing.B) {
	benchFullTick(b, 2000, true)
}
