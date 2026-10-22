package page

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// warmBudgetHTML is a few elements: one filled box and one paragraph.
// The warm-redraw budget uses this page, not a hundreds-of-rows list.
const warmBudgetHTML = `<style>` +
	`#box{width:200px;height:80px;background:#3366cc}` +
	`p{color:#222222;font-size:14px;line-height:20px}` +
	`</style><div id="box"></div>` +
	`<p>Resize me. Resize me. Resize me. Resize me. Resize me.</p>`

const (
	warmBudgetWidth  = 640
	warmBudgetHeight = 480
)

// Measured one warm Redraw on this machine with Perf off: 640us and
// 100200 bytes. Ceilings use factor 20, so a noisy CI box still passes.
// A multi-second hang or a 100x regression still fails.
const (
	measuredWarmRedraw     = 640 * time.Microsecond
	warmRedrawTimeLimit    = measuredWarmRedraw * 20
	measuredWarmAllocBytes = 100200
	warmRedrawAllocLimit   = measuredWarmAllocBytes * 20
)

// TestWarmRedrawBudget fails when a warm Redraw of a small page, Perf off,
// takes longer than the time ceiling or allocates more than the byte ceiling.
// One cold Redraw fills the cache. The next Redraw is the sample.
func TestWarmRedrawBudget(t *testing.T) {
	ctx := context.Background()

	p, err := New(Config{
		HTML:   warmBudgetHTML,
		Width:  warmBudgetWidth,
		Height: warmBudgetHeight,
		Perf:   false,
	})
	if err != nil {
		t.Fatal(err)
	}

	if p.Perf() {
		t.Fatal("perf on")
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// A collection here keeps a leftover GC off the timed call.
	runtime.GC()

	var before, after runtime.MemStats

	runtime.ReadMemStats(&before)

	start := time.Now()
	err = p.Redraw(ctx)
	elapsed := time.Since(start)

	runtime.ReadMemStats(&after)

	if err != nil {
		t.Fatal(err)
	}

	allocs := after.TotalAlloc - before.TotalAlloc
	t.Logf("warm redraw %s, %d bytes", elapsed, allocs)

	if elapsed > warmRedrawTimeLimit {
		t.Fatalf("warm redraw %s, limit %s", elapsed, warmRedrawTimeLimit)
	}

	if allocs > warmRedrawAllocLimit {
		t.Fatalf("warm redraw allocs %d bytes, limit %d", allocs, warmRedrawAllocLimit)
	}
}
