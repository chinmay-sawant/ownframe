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
	Reference  string
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
		Reference: strings.TrimSpace(os.Getenv("OWNFRAME_PROP_REFERENCE_DIR")),
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
	Name           string   `json:"name"`
	Group          string   `json:"group"`
	Sample         string   `json:"sample"`
	OverrideUsed   bool     `json:"override_used"`
	Declined       bool     `json:"declined"`
	Effective      bool     `json:"effective"`
	EffectEvidence []string `json:"effect_evidence,omitempty"`
	HighPrecision  bool     `json:"high_precision"`
	Layer          string   `json:"bottleneck_layer"`

	// Phase 1: Cold build (New + Redraw), paired against same DOM scaffold without property.
	ColdNS          int64  `json:"cold_ns"`
	ColdP95NS       int64  `json:"cold_p95_ns"`
	ColdIQRNS       int64  `json:"cold_iqr_ns"`
	ColdControlNS   int64  `json:"cold_control_ns"`
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
	HoverControlNS      int64    `json:"hover_control_ns"`
	HoverIQRNS          int64    `json:"hover_iqr_ns"`
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
	ResizeNS        int64  `json:"resize_ns"`
	ResizeControlNS int64  `json:"resize_control_ns"`
	ResizeIQRNS     int64  `json:"resize_iqr_ns"`
	ResizeDeltaNS   int64  `json:"resize_delta_ns"`
	ResizeBytes     uint64 `json:"resize_bytes"`
	ResizeAllocs    uint64 `json:"resize_allocs"`
	ResizeStyleNS   int64  `json:"resize_style_ns"`
	ResizeLayoutNS  int64  `json:"resize_layout_ns"`

	// Phase 5: Scaled uniform 64-node container build (exposes O(N^2) scaling).
	Scale64NS        int64   `json:"scale64_ns"`
	Scale64ControlNS int64   `json:"scale64_control_ns"`
	Scale64IQRNS     int64   `json:"scale64_iqr_ns"`
	Scale64DeltaNS   int64   `json:"scale64_delta_ns"`
	Scale64Bytes     uint64  `json:"scale64_bytes"`
	Scale64Allocs    uint64  `json:"scale64_allocs"`
	Scale64Ops       int     `json:"scale64_ops"`
	ScalingRatio     float64 `json:"scaling_ratio"`

	// Phase 6: Tick + software raster (PNGWithContext) + display-list signals.
	TickNS          int64  `json:"tick_ns"`
	TickAllocs      uint64 `json:"tick_allocs"`
	MotionBound     int    `json:"motion_bound"`
	MotionActive    int    `json:"motion_active"`
	RasterNS        int64  `json:"raster_ns"`
	RasterControlNS int64  `json:"raster_control_ns"`
	RasterIQRNS     int64  `json:"raster_iqr_ns"`
	RasterDeltaNS   int64  `json:"raster_delta_ns"`
	RasterBytes     uint64 `json:"raster_bytes"`
	RasterAllocs    uint64 `json:"raster_allocs"`
	Ops             int    `json:"ops"`
	ExtraOps        int    `json:"extra_ops"`
	BitmapFallback  bool   `json:"bitmap_fallback"`

	PhaseErrors map[string]string `json:"phase_errors,omitempty"`
	Err         string            `json:"error,omitempty"`
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
	result, err := measureWorkloadsNS(iters, gcIsolate, fn)
	if err != nil {
		return timingStats{}, err
	}

	return result[0], nil
}

// measureWorkloadsNS rotates the first workload on each iteration. Controls
// and variants share the GC boundary and sample the same period of host load.
func measureWorkloadsNS(iters int, gcIsolate bool, workloads ...func() error) ([]timingStats, error) {
	iters = max(1, iters)
	if gcIsolate {
		for _, run := range workloads {
			if err := run(); err != nil {
				return nil, err
			}
		}
		runtime.GC()
	}
	prevGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGC)

	samples := make([][]int64, len(workloads))
	for i := range samples {
		samples[i] = make([]int64, iters)
	}
	for i := range iters {
		for j := range workloads {
			k := (i + j) % len(workloads)
			start := time.Now()
			if err := workloads[k](); err != nil {
				return nil, err
			}
			samples[k][i] = time.Since(start).Nanoseconds()
		}
	}

	result := make([]timingStats, len(samples))
	for i := range result {
		result[i] = summarizeTimings(samples[i])
	}

	return result, nil
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
			writeF64(op.Width)
			writeF64(float64(op.RotateDeg))
			writeU64(uint64(op.StrokeMask))
			if op.XformSet {
				m := op.Transform()
				for _, value := range [...]float64{m.A, m.B, m.C, m.D, m.E, m.F} {
					writeF64(value)
				}
			}
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
