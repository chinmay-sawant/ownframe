package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/leakcheck"
	"image"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	budgetInputs = 2000
	budgetDepth  = 200
	budgetMaxLag = 10 * time.Second
	budgetWide   = 500000
	budgetDeep   = 20000
	budgetGuard  = 1000
)

func adversarialWide() string {
	return "<style>" + strings.Repeat(".c,", budgetInputs) + ".last{width:10px}</style><div>" + strings.Repeat(`<input value="v">`, budgetInputs) + "</div>"
}

func adversarialDeep() string {
	return strings.Repeat("<div>", budgetDepth) + "x" + strings.Repeat("</div>", budgetDepth)
}

func redrawNew(html string) error {
	p, err := New(Config{HTML: html, Width: 800, Height: 600})
	if err != nil {
		return err
	}

	return p.Redraw(context.Background())
}

// TestComplexityBudget pins adversarial costs (GEN-27): wide selector
// lists with thousands of fields and near-cap nesting finish inside a
// hard deadline and allocation budget.
func TestComplexityBudget(t *testing.T) {
	for _, html := range []string{adversarialWide(), adversarialDeep()} {
		start := time.Now()
		if err := redrawNew(html); err != nil {
			t.Fatal(err)
		}
		if d := time.Since(start); d > budgetMaxLag {
			t.Fatalf("redraw took %s", d)
		}
	}
	if n := testing.AllocsPerRun(3, func() { _ = redrawNew(adversarialWide()) }); n > budgetWide {
		t.Fatalf("wide allocs = %v", n)
	}
	if n := testing.AllocsPerRun(3, func() { _ = redrawNew(adversarialDeep()) }); n > budgetDeep {
		t.Fatalf("deep allocs = %v", n)
	}
	if n := testing.AllocsPerRun(20, func() { _ = checkDOMLimits(adversarialWide()) }); n > budgetGuard {
		t.Fatalf("guard allocs = %v", n)
	}
}

// BenchmarkAdversarialWide tracks the wide layout path.
func BenchmarkAdversarialWide(b *testing.B) {
	html := adversarialWide()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redrawNew(html)
	}
}

// BenchmarkDOMGuard tracks the depth/node guard.
func BenchmarkDOMGuard(b *testing.B) {
	html := adversarialDeep()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checkDOMLimits(html)
	}
}

// TestMain fails the binary when page tests leak goroutines (GEN-26).
func TestMain(m *testing.M) {
	os.Exit(leakcheck.Verify(m))
}

func TestPerfDisjointRegions(t *testing.T) {
	p := &Page{}
	d := &Display{PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3, Ops: []layout.DisplayOp{
		{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 150, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 75, Y: 0, W: 15, H: 15},
		{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 165, H: 15},
	}}
	rects := []image.Rectangle{image.Rect(0, 0, 20, 20), image.Rect(200, 0, 220, 20)}
	recordPartial(p, 2, rects, d)
	if p.stats.dirtyRegions != 2 || p.stats.dirtyArea != 800 || p.stats.dirtyOps != 3 {
		t.Fatalf("regions=%d area=%d ops=%d", p.stats.dirtyRegions, p.stats.dirtyArea, p.stats.dirtyOps)
	}
	recordPartial(p, 0, nil, d)
	if p.stats.dirtyRegions != 0 || p.stats.dirtyArea != 0 || p.stats.dirtyOps != 0 {
		t.Fatal("empty damage retained counters")
	}
}

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

// TestSec_PageFieldCaps pins GEN-28: the author maxlength is honored and
// the default cap bounds fields without one.
func TestSec_PageFieldCaps(t *testing.T) {
	t.Parallel()
	p, err := New(Config{HTML: `<input id="a" maxlength="3" value=""><input id="b" value="">`, Width: 10, Height: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.SetFormValue("a", "abcdef")
	if got := p.FormValue("a"); got != "abc" {
		t.Fatalf("maxlength: %q", got)
	}
	p.SetFormValue("b", strings.Repeat("x", DefaultMaxFieldLen+100))
	if got := p.FormValue("b"); len([]rune(got)) != DefaultMaxFieldLen {
		t.Fatalf("default cap: %d", len([]rune(got)))
	}
	var zero Page
	next, _, _, changed := insertValue(&zero, Control{MaxLength: 2}, 0, 0, "abcdef")
	if !changed || next.Value != "ab" {
		t.Fatalf("insert: %q %v", next.Value, changed)
	}
}

// TestSec_PageSrcAllowlist pins GEN-30: blocked schemes fail at resolve
// time, oversized data:image fails, and the allowlist passes.
func TestSec_PageSrcAllowlist(t *testing.T) {
	t.Parallel()
	for _, src := range []string{"file:///etc/passwd", "javascript:alert(1)", "data:text/plain,hi", "vbscript:x"} {
		if err := checkImageSrc(src); !errors.Is(err, ErrUnsupportedLinkScheme) {
			t.Fatalf("%q: %v", src, err)
		}
	}
	if err := checkImageSrc("data:image/png;base64," + strings.Repeat("A", MaxDataImageBytes)); !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("oversize: %v", err)
	}
	for _, src := range []string{"logo", "emoji/smile", "https://x/y.png", "http://x/y.png", "data:image/png;base64,AA=="} {
		if err := checkImageSrc(src); err != nil {
			t.Fatalf("%q: %v", src, err)
		}
	}
	if _, err := parseOutboundLink("javascript:alert(1)"); !errors.Is(err, ErrUnsupportedLinkScheme) {
		t.Fatalf("href: %v", err)
	}
}

// TestSec_PageSourceSizeCapped pins the template size cap.
func TestSec_PageSourceSizeCapped(t *testing.T) {
	t.Parallel()
	if _, err := New(Config{HTML: strings.Repeat("x", MaxSourceSize+1), Width: 10, Height: 10}); !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

// TestSec_PageDOMLimits pins the nesting and node caps at every parse
// entry: New, Load, and Route registration.
func TestSec_PageDOMLimits(t *testing.T) {
	t.Parallel()
	deep := strings.Repeat("<div>", MaxDOMDepth+1) + "x" + strings.Repeat("</div>", MaxDOMDepth+1)
	if _, err := New(Config{HTML: deep, Width: 10, Height: 10}); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatalf("deep: %v", err)
	}
	many := strings.Repeat("<i>x</i>", MaxDOMNodes+1)
	if _, err := New(Config{HTML: many, Width: 10, Height: 10}); !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("nodes: %v", err)
	}
	p, err := New(Config{HTML: "<p>x</p>", Width: 10, Height: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Load(context.Background(), deep); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatalf("load: %v", err)
	}
	if err := p.Route("a", many); !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("route: %v", err)
	}
}
