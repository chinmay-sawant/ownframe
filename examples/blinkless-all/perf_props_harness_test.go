package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
)

const (
	propBenchWidth         = 360
	propBenchHeight        = 160
	propBenchResizeWidth   = 480
	propBenchResizeHeight  = 200
	propBenchScaleWidth    = 760
	propBenchScaleHeight   = 520
	propBenchScaleNodes    = 64
	propBenchHoverNodes    = 64
	propBenchDefaultIters  = 5
	propBenchTopPairs      = 10
	propBenchFuzzCount     = 25
	propBenchFuzzSize      = 5
	propBenchDefaultSeed   = 42
	propBenchPprofTop      = 9
	propBenchRefineTop     = 40
	propBenchDriftEvery    = 100
	propBenchLogEvery      = 50
	propBenchP95Numerator  = 95
	propBenchPercentBase   = 100
	propBenchQuartileDiv   = 4
	propBenchQ3Mul         = 3
	propBenchHalfDiv       = 2
	propBenchMinBaseNS     = 1000
	propBenchRasterMinMul  = 10
	propBenchComboMinMul   = 5
	propBenchComboMinOut   = 10
	propBenchScoreWeight   = 4
	propBenchSigScale      = 100
	propBenchByteBits      = 8
	propBenchBitmapTag     = 0xB17A9
	propBenchPCGStreamXOR  = 0x9e3779b97f4a7c15
	propBenchFNVPrime      = 1099511628211
	propBenchSuperLinear   = 1.5
	propBenchHoverOffX     = 4.0
	propBenchHoverOffY     = 4.0
	propBenchComboRefineN  = 10
	propBenchComboRefineIt = 3
)

type propBenchOpts struct {
	DumpDir    string
	Iters      int
	RefineTop  int
	TopPairs   int
	FuzzCount  int
	Seed       uint64
	PprofTop   int
	Filter     *regexp.Regexp
	Limit      int
	MaxBuckets int
	MaxRecipes int
}

func loadPropBenchOpts() (propBenchOpts, error) {
	opts := propBenchOpts{
		DumpDir:   strings.TrimSpace(os.Getenv("OWNFRAME_PROP_DUMP_DIR")),
		Iters:     envInt("OWNFRAME_PROP_ITERS", propBenchDefaultIters),
		RefineTop: envInt("OWNFRAME_PROP_REFINE_TOP", propBenchRefineTop),
		TopPairs:  envInt("OWNFRAME_PROP_TOP_PAIRS", propBenchTopPairs),
		FuzzCount: envInt("OWNFRAME_PROP_FUZZ", propBenchFuzzCount),
		Seed:      envUint64("OWNFRAME_PROP_SEED", propBenchDefaultSeed),
		PprofTop:  envInt("OWNFRAME_PROP_PPROF_TOP", propBenchPprofTop),
		Limit:     envInt("OWNFRAME_PROP_LIMIT", 0),
	}
	if opts.Iters < 1 {
		opts.Iters = 1
	}
	if pat := strings.TrimSpace(os.Getenv("OWNFRAME_PROP_FILTER")); pat != "" {
		re, err := regexp.Compile(pat)
		if err != nil {
			return opts, fmt.Errorf("compile OWNFRAME_PROP_FILTER %q: %w", pat, err)
		}
		opts.Filter = re
	}

	return opts, nil
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return n
}

func envUint64(key string, fallback uint64) uint64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return fallback
	}

	return n
}

type timingStats struct {
	MedianNS int64 `json:"median_ns"`
	P95NS    int64 `json:"p95_ns"`
	IQRNS    int64 `json:"iqr_ns"`
}

type allocStats struct {
	Bytes  uint64 `json:"bytes"`
	Allocs uint64 `json:"allocs"`
}

// sharedBaselines holds the invariant "off" measurements for uniform 1-node and
// 64-node grids so we do not re-benchmark identical controls 688+ times.
type sharedBaselines struct {
	Cold1NS   int64
	Hover64NS int64
	Scale64NS int64
	Raster1NS int64
}

// propDumpRow is one Tier-1 paired-control record emitted to props-individual.jsonl.
type propDumpRow struct {
	Name          string `json:"name"`
	Group         string `json:"group"`
	Sample        string `json:"sample"`
	OverrideUsed  bool   `json:"override_used"`
	Declined      bool   `json:"declined"`
	Effective     bool   `json:"effective"`
	HighPrecision bool   `json:"high_precision"`
	Layer         string `json:"bottleneck_layer"`

	// Phase 1: Cold build (New + Redraw), paired against same DOM scaffold without property.
	ColdNS          int64  `json:"cold_ns"`
	ColdP95NS       int64  `json:"cold_p95_ns"`
	ColdIQRNS       int64  `json:"cold_iqr_ns"`
	ColdDeltaNS     int64  `json:"cold_delta_ns"`
	ColdBytes       uint64 `json:"cold_bytes"`
	ColdDeltaBytes  int64  `json:"cold_delta_bytes"`
	ColdAllocs      uint64 `json:"cold_allocs"`
	ColdDeltaAllocs int64  `json:"cold_delta_allocs"`
	StyleNS         int64  `json:"style_ns"`
	LayoutNS        int64  `json:"layout_ns"`
	PaintNS         int64  `json:"paint_ns"`
	OtherNS         int64  `json:"other_ns"`

	// Phase 2: Warm unchanged redraw after one settle tick.
	WarmNS     int64  `json:"warm_ns"`
	WarmBytes  uint64 `json:"warm_bytes"`
	WarmAllocs uint64 `json:"warm_allocs"`

	// Phase 3: Hover toggle inside a 64-node RetainState container.
	HoverNS             int64    `json:"hover_ns"`
	HoverDeltaNS        int64    `json:"hover_delta_ns"`
	HoverBytes          uint64   `json:"hover_bytes"`
	HoverAllocs         uint64   `json:"hover_allocs"`
	HoverPatches        uint64   `json:"hover_patches"`
	HoverFallbacks      uint64   `json:"hover_fallbacks"`
	HoverPatchErrors    uint64   `json:"hover_patch_errors"`
	HoverFailureReasons []string `json:"hover_failure_reasons,omitempty"`
	HoverDirtyRegions   int      `json:"hover_dirty_regions"`
	HoverDirtyOps       int      `json:"hover_dirty_ops"`

	// Phase 4: Viewport resize relayout (360x160 -> 480x200 -> 360x160).
	ResizeNS       int64  `json:"resize_ns"`
	ResizeDeltaNS  int64  `json:"resize_delta_ns"`
	ResizeBytes    uint64 `json:"resize_bytes"`
	ResizeAllocs   uint64 `json:"resize_allocs"`
	ResizeStyleNS  int64  `json:"resize_style_ns"`
	ResizeLayoutNS int64  `json:"resize_layout_ns"`

	// Phase 5: Scaled uniform 64-node container build (exposes O(N^2) scaling).
	Scale64NS      int64   `json:"scale64_ns"`
	Scale64DeltaNS int64   `json:"scale64_delta_ns"`
	Scale64Bytes   uint64  `json:"scale64_bytes"`
	Scale64Allocs  uint64  `json:"scale64_allocs"`
	Scale64Ops     int     `json:"scale64_ops"`
	ScalingRatio   float64 `json:"scaling_ratio"`

	// Phase 6: Tick + software raster (PNGWithContext) + display-list signals.
	TickNS         int64  `json:"tick_ns"`
	TickAllocs     uint64 `json:"tick_allocs"`
	MotionBound    int    `json:"motion_bound"`
	MotionActive   int    `json:"motion_active"`
	RasterNS       int64  `json:"raster_ns"`
	RasterDeltaNS  int64  `json:"raster_delta_ns"`
	RasterBytes    uint64 `json:"raster_bytes"`
	RasterAllocs   uint64 `json:"raster_allocs"`
	Ops            int    `json:"ops"`
	ExtraOps       int    `json:"extra_ops"`
	BitmapFallback bool   `json:"bitmap_fallback"`

	Err string `json:"error,omitempty"`
}

// activeSampleOverrides replaces passive ("none") or syntax-incomplete demo
// values from props_gen.go with active CSS values so real engine paths run.
func activeSampleOverrides() map[string]string {
	return map[string]string{
		"backdrop-filter":             "blur(12px)",
		"filter":                      "blur(4px)",
		"box-shadow":                  "0 8px 24px rgba(0,0,0,0.35)",
		"text-shadow":                 "1px 2px 4px rgba(0,0,0,0.4)",
		"mask-image":                  "linear-gradient(black, transparent)",
		"-webkit-mask-image":          "linear-gradient(black, transparent)",
		"mask":                        "linear-gradient(black, transparent)",
		"-webkit-mask":                "linear-gradient(black, transparent)",
		"transition-property":         "opacity, transform",
		"-webkit-transition-property": "opacity, transform",
		"animation-name":              "slide",
		"-webkit-animation-name":      "slide",
		"will-change":                 "transform, opacity",
		"clip-path":                   "inset(4px round 8px)",
	}
}

func activeSample(r genProp) (string, bool) {
	if v, ok := activeSampleOverrides()[r.Name]; ok {
		return v, true
	}

	return r.Sample, false
}

func propBenchHead() string {
	return genHead() +
		`<style>` +
		`@keyframes slide{from{opacity:0.4;transform:translateX(0px)}to{opacity:1;transform:translateX(12px)}}` +
		`.pgrid{display:flex;flex-wrap:wrap;gap:4px;padding:4px}` +
		`.pitem{width:84px;height:22px;padding:2px 4px;background:#f1efe7;border:1px solid #bbb;font-size:11px;box-sizing:border-box}` +
		`</style>`
}

// measureTimedNS isolates timing from Go GC pauses and ReadMemStats STW pauses.
// When gcIsolate is true (used in Pass-2 high-precision runs), it forces a GC
// before disabling GC for the tight sampling loop; Pass-1 fast sweeps disable
// GC once per property without per-phase GC storms.
func measureTimedNS(iters int, gcIsolate bool, fn func() error) (timingStats, error) {
	if iters < 1 {
		iters = 1
	}
	if gcIsolate {
		if err := fn(); err != nil {
			return timingStats{}, err
		}
		runtime.GC()
	}
	prevGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGC)

	samples := make([]int64, iters)
	for i := range iters {
		t0 := time.Now()
		if err := fn(); err != nil {
			return timingStats{}, err
		}
		samples[i] = time.Since(t0).Nanoseconds()
	}

	return summarizeTimings(samples), nil
}

func summarizeTimings(samples []int64) timingStats {
	if len(samples) == 0 {
		return timingStats{}
	}
	slices.Sort(samples)
	n := len(samples)
	med := samples[n/propBenchHalfDiv]
	if n%propBenchHalfDiv == 0 && n >= propBenchHalfDiv {
		med = (samples[n/propBenchHalfDiv-1] + samples[n/propBenchHalfDiv]) / propBenchHalfDiv
	}
	p95Idx := (n*propBenchP95Numerator - 1) / propBenchPercentBase
	if p95Idx >= n {
		p95Idx = n - 1
	}
	if p95Idx < 0 {
		p95Idx = 0
	}
	q1 := samples[n/propBenchQuartileDiv]
	q3 := samples[(n*propBenchQ3Mul)/propBenchQuartileDiv]

	return timingStats{
		MedianNS: med,
		P95NS:    samples[p95Idx],
		IQRNS:    q3 - q1,
	}
}

// measureAllocAvg measures steady-state allocations in a separate pass from
// timing so ReadMemStats STW pauses never contaminate latency samples.
func measureAllocAvg(iters int, fn func() error) (allocStats, error) {
	if iters < 1 {
		iters = 1
	}
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	for range iters {
		if err := fn(); err != nil {
			return allocStats{}, err
		}
	}
	runtime.ReadMemStats(&m1)
	denom := uint64(iters)

	return allocStats{
		Bytes:  (m1.TotalAlloc - m0.TotalAlloc) / denom,
		Allocs: (m1.Mallocs - m0.Mallocs) / denom,
	}, nil
}

func newBenchPage(htmlDoc string, w, h int, retain bool) (*ownframe.Page, error) {
	return ownframe.NewWithOptions(ownframe.Config{
		Title:       "prop-bench",
		HTML:        htmlDoc,
		Width:       w,
		Height:      h,
		RetainState: retain,
	}, ownframe.WithPerf(true))
}

func displaySignature(p *ownframe.Page) uint64 {
	h := fnv.New64a()
	writeU64 := func(v uint64) {
		var buf [propBenchByteBits]byte
		binary.LittleEndian.PutUint64(buf[:], v)
		_, _ = h.Write(buf[:])
	}
	writeF64 := func(v float64) {
		writeU64(math.Float64bits(math.Round(v*propBenchSigScale) / propBenchSigScale))
	}

	for _, b := range p.Boxes() {
		_, _ = h.Write([]byte(b.ID))
		writeF64(b.X)
		writeF64(b.Y)
		writeF64(b.W)
		writeF64(b.H)
	}
	if d := p.Probe().Display(); d != nil {
		writeU64(uint64(len(d.Ops)))
		for i := range d.Ops {
			op := &d.Ops[i]
			writeU64(uint64(op.Kind))
			writeF64(op.X)
			writeF64(op.Y)
			writeF64(op.W)
			writeF64(op.H)
			writeF64(op.R)
			writeF64(op.G)
			writeF64(op.B)
			writeF64(op.Alpha)
			writeF64(op.Size)
			writeF64(op.Radius)
			writeF64(op.LetterSpacing)
			_, _ = h.Write([]byte(op.Text))
			if op.HasExtra() {
				extraH, _ := op.MixExtraStamp(h.Sum64(), propBenchFNVPrime)
				writeU64(extraH)
			}
		}
	}
	if p.Probe().Image() != nil {
		writeU64(propBenchBitmapTag)
	}

	return h.Sum64()
}

func countExtraOps(d *layout.Display) int {
	if d == nil {
		return 0
	}
	n := 0
	for i := range d.Ops {
		if d.Ops[i].HasExtra() {
			n++
		}
	}

	return n
}

// buildUniformScaleDoc applies styleDecl across 64 identical .pitem boxes so
// Scale64 is apples-to-apples across every property (never multiplying a
// 20-node table/SVG demo into 1,280 nodes).
func buildUniformScaleDoc(head, styleDecl string, count int) string {
	var b strings.Builder
	b.WriteString(head)
	b.WriteString(`<div class="pgrid">`)
	for i := range count {
		if styleDecl != "" {
			fmt.Fprintf(&b, `<div id="s%d" class="pitem" style="%s">cell %d</div>`, i, styleDecl, i)
		} else {
			fmt.Fprintf(&b, `<div id="s%d" class="pitem">cell %d</div>`, i, i)
		}
	}
	b.WriteString(`</div>`)

	return b.String()
}

func buildHoverDoc(head, propName, sample string, withStyle bool) string {
	var b strings.Builder
	b.WriteString(head)
	b.WriteString(`<style>`)
	if withStyle {
		b.WriteString(`#h0:hover{`)
		b.WriteString(propName)
		b.WriteString(`:`)
		b.WriteString(sample)
		b.WriteString(`;}`)
	} else {
		b.WriteString(`#h0:hover{}`)
	}
	b.WriteString(`</style><div class="pgrid">`)
	for i := range propBenchHoverNodes {
		fmt.Fprintf(&b, `<div id="h%d" class="pitem">item %d</div>`, i, i)
	}
	b.WriteString(`</div>`)

	return b.String()
}

func findHoverTarget(p *ownframe.Page, id string) (float64, float64, bool) {
	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.X + b.W/propBenchHalfDiv, b.Y + b.H/propBenchHalfDiv, true
		}
	}

	return 0, 0, false
}

func classifyBottleneckLayer(row propDumpRow) string {
	if row.BitmapFallback {
		return "ownframe/replay+bitmap"
	}
	if row.HoverFallbacks > 0 || row.WarmAllocs > 0 {
		return "ownframe/page+motion"
	}
	if row.RasterDeltaNS > row.ColdDeltaNS*propBenchHalfDiv && row.RasterDeltaNS > propBenchMinBaseNS*propBenchRasterMinMul {
		return "ownframe/replay+bitmap"
	}
	if row.StyleNS > row.LayoutNS && row.StyleNS > row.OtherNS && row.StyleNS > row.PaintNS {
		return "blinkless/css"
	}
	if row.LayoutNS >= row.StyleNS && row.LayoutNS > row.OtherNS && row.LayoutNS > row.PaintNS {
		return "blinkless/layout"
	}
	if row.PaintNS > row.OtherNS {
		return "ownframe/replay+bitmap"
	}
	if row.OtherNS > 0 {
		return "ownframe/page+motion"
	}

	return "balanced"
}

func measureSharedBaselines(ctx context.Context, iters int) (sharedBaselines, error) {
	head := propBenchHead()
	off1 := buildUniformScaleDoc(head, "", 1)
	off64 := buildUniformScaleDoc(head, "", propBenchScaleNodes)
	hOff := buildHoverDoc(head, "", "", false)

	c1, err := measureTimedNS(iters, true, func() error {
		p, pErr := newBenchPage(off1, propBenchWidth, propBenchHeight, false)
		if pErr != nil {
			return pErr
		}

		return p.Redraw(ctx)
	})
	if err != nil {
		return sharedBaselines{}, err
	}

	s64, err := measureTimedNS(iters, true, func() error {
		p, pErr := newBenchPage(off64, propBenchScaleWidth, propBenchScaleHeight, false)
		if pErr != nil {
			return pErr
		}

		return p.Redraw(ctx)
	})
	if err != nil {
		return sharedBaselines{}, err
	}

	hT, _, _, _, _, _, err := measureSingleHover(ctx, hOff, iters, true)
	if err != nil {
		return sharedBaselines{}, err
	}

	rT, _, err := measureSingleRaster(ctx, off1, propBenchWidth, propBenchHeight, iters, true)
	if err != nil {
		return sharedBaselines{}, err
	}

	return sharedBaselines{
		Cold1NS:   c1.MedianNS,
		Hover64NS: hT.MedianNS,
		Scale64NS: s64.MedianNS,
		Raster1NS: rT.MedianNS,
	}, nil
}

func measureColdSingle(ctx context.Context, doc string, w, h, iters int, gcIsolate, withAllocs bool) (
	timingStats, allocStats, ownframe.Stats, *ownframe.Page, uint64, error,
) {
	var kept *ownframe.Page
	var sig uint64
	run := func() error {
		p, err := newBenchPage(doc, w, h, false)
		if err != nil {
			return err
		}
		if err := p.Redraw(ctx); err != nil {
			return err
		}
		kept = p
		if sig == 0 {
			sig = displaySignature(p)
		}

		return nil
	}
	tm, err := measureTimedNS(iters, gcIsolate, run)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, nil, 0, err
	}
	var al allocStats
	if withAllocs {
		al, err = measureAllocAvg(max(1, iters/propBenchHalfDiv), run)
		if err != nil {
			return timingStats{}, allocStats{}, ownframe.Stats{}, nil, 0, err
		}
	}
	var st ownframe.Stats
	if kept != nil {
		st = kept.Probe().Stats()
	}

	return tm, al, st, kept, sig, nil
}

func measureSingleHover(ctx context.Context, doc string, iters int, gcIsolate bool) (
	timingStats, allocStats, ownframe.Stats, int, int, []string, error,
) {
	p, err := newBenchPage(doc, propBenchScaleWidth, propBenchScaleHeight, true)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	if err := p.Redraw(ctx); err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	hx, hy, ok := findHoverTarget(p, "h0")
	if !ok {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, fmt.Errorf("missing hover target h0")
	}
	toggle := func() error {
		if hErr := p.Edit().Hover(ctx, hx, hy); hErr != nil {
			return hErr
		}
		if tErr := p.Testing().Tick(ctx); tErr != nil {
			return tErr
		}
		if hErr := p.Edit().Hover(ctx, propBenchHoverOffX, propBenchHoverOffY); hErr != nil {
			return hErr
		}

		return p.Testing().Tick(ctx)
	}
	tm, err := measureTimedNS(iters, gcIsolate, toggle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	al, err := measureAllocAvg(1, toggle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	st := p.Probe().Stats()
	var rs []string
	for _, f := range st.HoverPatchFailures {
		rs = append(rs, fmt.Sprintf("%s:%s(%d)", f.Stage, f.Reason, f.Count))
	}

	return tm, al, st, st.DirtyRegions, st.DirtyOps, rs, nil
}

func measureSingleResize(ctx context.Context, doc string, iters int, gcIsolate bool) (
	timingStats, allocStats, ownframe.Stats, error,
) {
	p, err := newBenchPage(doc, propBenchWidth, propBenchHeight, true)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, err
	}
	if err := p.Redraw(ctx); err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, err
	}
	cycle := func() error {
		p.SetSize(propBenchResizeWidth, propBenchResizeHeight)
		if rErr := p.Redraw(ctx); rErr != nil {
			return rErr
		}
		p.SetSize(propBenchWidth, propBenchHeight)

		return p.Redraw(ctx)
	}
	tm, err := measureTimedNS(iters, gcIsolate, cycle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, err
	}
	al, err := measureAllocAvg(1, cycle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, err
	}

	return tm, al, p.Probe().Stats(), nil
}

func measureSingleRaster(ctx context.Context, doc string, w, h, iters int, gcIsolate bool) (
	timingStats, allocStats, error,
) {
	pages := make([]*ownframe.Page, iters)
	for i := range iters {
		p, err := newBenchPage(doc, w, h, false)
		if err != nil {
			return timingStats{}, allocStats{}, err
		}
		if err := p.Redraw(ctx); err != nil {
			return timingStats{}, allocStats{}, err
		}
		pages[i] = p
	}
	if gcIsolate {
		runtime.GC()
	}
	prevGC := debug.SetGCPercent(-1)
	samples := make([]int64, iters)
	for i := range iters {
		t0 := time.Now()
		_ = pages[i].Probe().PNGWithContext(ctx)
		samples[i] = time.Since(t0).Nanoseconds()
	}
	debug.SetGCPercent(prevGC)

	allocPage, err := newBenchPage(doc, w, h, false)
	if err != nil {
		return timingStats{}, allocStats{}, err
	}
	if err := allocPage.Redraw(ctx); err != nil {
		return timingStats{}, allocStats{}, err
	}
	al, err := measureAllocAvg(1, func() error {
		_ = allocPage.Probe().PNGWithContext(ctx)

		return nil
	})
	if err != nil {
		return timingStats{}, allocStats{}, err
	}

	return summarizeTimings(samples), al, nil
}

// measurePropertyFull runs all 6 paired-control phases for one property.
// When highPrecision is false (Pass-1 fast sweep), it uses iters=1 without
// per-phase GC storms and reuses shared uniform baselines for Hover and Scale64.
// When highPrecision is true (Pass-2 outlier refinement), it runs full iters
// with GC isolation and paired off controls for every phase.
func measurePropertyFull(ctx context.Context, r genProp, iters int, highPrecision bool, base sharedBaselines) propDumpRow {
	sample, overrideUsed := activeSample(r)
	prop := r
	prop.Sample = sample

	row := propDumpRow{
		Name:          r.Name,
		Group:         r.Group,
		Sample:        sample,
		OverrideUsed:  overrideUsed,
		Declined:      declines(r.Name),
		HighPrecision: highPrecision,
	}

	phaseIters := 1
	if highPrecision {
		phaseIters = max(1, iters)
	}

	head := propBenchHead()
	spec := specFor(prop)
	offDoc := head + spec.expand(prop, "p", false)
	onDoc := head + spec.expand(prop, "p", true)

	// Phase 1: Paired cold build (always paired against the property's own spec scaffold).
	offT, offA, _, _, offSig, err := measureColdSingle(ctx, offDoc, propBenchWidth, propBenchHeight, phaseIters, highPrecision, true)
	if err != nil {
		row.Err = fmt.Sprintf("cold-off: %v", err)

		return row
	}
	onT, onA, onStats, onPage, onSig, err := measureColdSingle(ctx, onDoc, propBenchWidth, propBenchHeight, phaseIters, highPrecision, true)
	if err != nil {
		row.Err = fmt.Sprintf("cold-on: %v", err)

		return row
	}
	row.Effective = offSig != onSig
	row.ColdNS = onT.MedianNS
	row.ColdP95NS = onT.P95NS
	row.ColdIQRNS = onT.IQRNS
	row.ColdDeltaNS = onT.MedianNS - offT.MedianNS
	row.ColdBytes = onA.Bytes
	row.ColdDeltaBytes = safeSubUint64(onA.Bytes, offA.Bytes)
	row.ColdAllocs = onA.Allocs
	row.ColdDeltaAllocs = safeSubUint64(onA.Allocs, offA.Allocs)
	row.StyleNS = onStats.StyleTime.Nanoseconds()
	row.LayoutNS = onStats.LayoutTime.Nanoseconds()
	row.PaintNS = onStats.PaintTime.Nanoseconds()
	row.OtherNS = onStats.OtherTime.Nanoseconds()
	if d := onPage.Probe().Display(); d != nil {
		row.Ops = len(d.Ops)
		row.ExtraOps = countExtraOps(d)
	}
	row.BitmapFallback = onPage.Probe().Image() != nil && onPage.Probe().Display() == nil

	// Phase 2: Warm unchanged redraw + Phase 6a: Tick & motion counters.
	_ = onPage.Testing().Tick(ctx)
	motionStats := onPage.Probe().Motion()
	row.MotionBound = motionStats.Bound + motionStats.TransitionsBound
	row.MotionActive = motionStats.Active
	if row.MotionBound > 0 {
		row.Effective = true
	}

	if warmT, wErr := measureTimedNS(phaseIters, false, func() error { return onPage.Redraw(ctx) }); wErr == nil {
		row.WarmNS = warmT.MedianNS
	}
	if warmA, waErr := measureAllocAvg(1, func() error { return onPage.Redraw(ctx) }); waErr == nil {
		row.WarmBytes = warmA.Bytes
		row.WarmAllocs = warmA.Allocs
	}
	if tickT, tErr := measureTimedNS(phaseIters, false, func() error { return onPage.Testing().Tick(ctx) }); tErr == nil {
		row.TickNS = tickT.MedianNS
	}
	if tickA, taErr := measureAllocAvg(1, func() error { return onPage.Testing().Tick(ctx) }); taErr == nil {
		row.TickAllocs = tickA.Allocs
	}

	// Phase 3: Hover toggle in a 64-node retained container (reuses shared off baseline).
	hOn := buildHoverDoc(head, r.Name, sample, true)
	if hT, hA, hStats, hDirtyReg, hDirtyOps, hReasons, hErr := measureSingleHover(ctx, hOn, phaseIters, highPrecision); hErr == nil {
		row.HoverNS = hT.MedianNS
		row.HoverDeltaNS = hT.MedianNS - base.Hover64NS
		row.HoverBytes = hA.Bytes
		row.HoverAllocs = hA.Allocs
		row.HoverPatches = hStats.HoverPatches
		row.HoverFallbacks = hStats.HoverFallbacks
		row.HoverPatchErrors = hStats.HoverPatchErrors
		row.HoverFailureReasons = hReasons
		row.HoverDirtyRegions = hDirtyReg
		row.HoverDirtyOps = hDirtyOps
		if hDirtyOps > 0 {
			row.Effective = true
		}
	}

	// Phase 4: Viewport resize relayout.
	if rOffT, _, _, rOffErr := measureSingleResize(ctx, offDoc, phaseIters, false); rOffErr == nil {
		if rOnT, rOnA, rOnSt, rOnErr := measureSingleResize(ctx, onDoc, phaseIters, highPrecision); rOnErr == nil {
			row.ResizeNS = rOnT.MedianNS
			row.ResizeDeltaNS = rOnT.MedianNS - rOffT.MedianNS
			row.ResizeBytes = rOnA.Bytes
			row.ResizeAllocs = rOnA.Allocs
			row.ResizeStyleNS = rOnSt.StyleTime.Nanoseconds()
			row.ResizeLayoutNS = rOnSt.LayoutTime.Nanoseconds()
		}
	}

	// Phase 5: Uniform 64-node container build (reuses shared off64 baseline).
	scaleDecl := r.Name + ":" + sample + ";"
	sOn := buildUniformScaleDoc(head, scaleDecl, propBenchScaleNodes)
	sIters := 1
	if highPrecision {
		sIters = max(1, phaseIters/propBenchHalfDiv)
	}
	if sOnT, sOnA, _, sOnPage, _, sErr := measureColdSingle(ctx, sOn, propBenchScaleWidth, propBenchScaleHeight, sIters, highPrecision, true); sErr == nil {
		row.Scale64NS = sOnT.MedianNS
		row.Scale64DeltaNS = sOnT.MedianNS - base.Scale64NS
		row.Scale64Bytes = sOnA.Bytes
		row.Scale64Allocs = sOnA.Allocs
		if sOnPage != nil && sOnPage.Probe().Display() != nil {
			row.Scale64Ops = len(sOnPage.Probe().Display().Ops)
		}
		denom := float64(max(propBenchMinBaseNS, row.ColdDeltaNS))
		if row.Scale64DeltaNS > 0 {
			row.ScalingRatio = float64(row.Scale64DeltaNS) / denom
		}
	}

	// Phase 6b: Uncached software raster paired delta.
	rIters := 1
	if highPrecision {
		rIters = max(1, phaseIters/propBenchHalfDiv)
	}
	if rasOffT, _, rOffErr := measureSingleRaster(ctx, offDoc, propBenchWidth, propBenchHeight, rIters, false); rOffErr == nil {
		if rasOnT, rasOnA, rOnErr := measureSingleRaster(ctx, onDoc, propBenchWidth, propBenchHeight, rIters, highPrecision); rOnErr == nil {
			row.RasterNS = rasOnT.MedianNS
			row.RasterDeltaNS = rasOnT.MedianNS - rasOffT.MedianNS
			row.RasterBytes = rasOnA.Bytes
			row.RasterAllocs = rasOnA.Allocs
		}
	}

	row.Layer = classifyBottleneckLayer(row)

	return row
}

func safeSubUint64(a, b uint64) int64 {
	if a >= b {
		diff := a - b
		if diff > math.MaxInt64 {
			return math.MaxInt64
		}

		return int64(diff)
	}
	diff := b - a
	if diff > math.MaxInt64 {
		return math.MinInt64
	}

	return -int64(diff)
}
