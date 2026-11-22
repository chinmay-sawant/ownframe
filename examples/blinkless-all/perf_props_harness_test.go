package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"html"
	"math"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"testing"
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
	// propBenchNoiseNS is the smallest per-stage or per-phase delta the
	// classifier and the scaling ratio treat as a property cost. The
	// no-effect rows of a full sweep spread about ±0.2 ms on single-sample
	// phases, so stage deltas below this read as noise, not as a layer.
	propBenchNoiseNS = 20_000
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
	// SupportDecl is the declaration list the variant adds before the
	// property so an otherwise inert longhand runs (see animSupportDecl).
	SupportDecl string `json:"support_decl,omitempty"`
	// EngineError marks Err as a product failure (page build or Redraw
	// refused the variant or control), not a harness failure.
	EngineError bool `json:"engine_error,omitempty"`

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
	// Control* are the control page's stage times from the same cold pair.
	// The layer classifier uses variant minus control per stage.
	ControlStyleNS  int64 `json:"control_style_ns"`
	ControlLayoutNS int64 `json:"control_layout_ns"`
	ControlPaintNS  int64 `json:"control_paint_ns"`
	ControlOtherNS  int64 `json:"control_other_ns"`

	// Phase 2: Unchanged redraw after one settle tick. An identical Redraw
	// normally skips, so this measures the skip check plus any rebuild the
	// property forces. WarmControlNS is the same redraw on the control page.
	WarmNS        int64  `json:"warm_ns"`
	WarmControlNS int64  `json:"warm_control_ns"`
	WarmDeltaNS   int64  `json:"warm_delta_ns"`
	WarmBytes     uint64 `json:"warm_bytes"`
	WarmAllocs    uint64 `json:"warm_allocs"`

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
	Scale64NS        int64  `json:"scale64_ns"`
	Scale64ControlNS int64  `json:"scale64_control_ns"`
	Scale64IQRNS     int64  `json:"scale64_iqr_ns"`
	Scale64DeltaNS   int64  `json:"scale64_delta_ns"`
	Scale64Bytes     uint64 `json:"scale64_bytes"`
	Scale64Allocs    uint64 `json:"scale64_allocs"`
	Scale64Ops       int    `json:"scale64_ops"`
	// Scale1DeltaNS is the paired 1-node uniform delta with the same
	// declaration. ScalingRatio is Scale64DeltaNS / Scale1DeltaNS, left 0
	// when either delta is within propBenchNoiseNS.
	Scale1DeltaNS int64   `json:"scale1_delta_ns"`
	ScalingRatio  float64 `json:"scaling_ratio"`

	// Phase 6: Tick + software raster (PNGWithContext) + display-list signals.
	TickNS          int64  `json:"tick_ns"`
	TickControlNS   int64  `json:"tick_control_ns"`
	TickAllocs      uint64 `json:"tick_allocs"`
	MotionBound     int    `json:"motion_bound"`
	MotionActive    int    `json:"motion_active"`
	RasterNS        int64  `json:"raster_ns"`
	RasterControlNS int64  `json:"raster_control_ns"`
	RasterIQRNS     int64  `json:"raster_iqr_ns"`
	RasterDeltaNS   int64  `json:"raster_delta_ns"`
	RasterBytes     uint64 `json:"raster_bytes"`
	RasterAllocs    uint64 `json:"raster_allocs"`
	// RasterDeltaValid is false when only one side is a bitmap fallback:
	// that side painted during Redraw, before the PNG-only timer, so
	// RasterDeltaNS compares encode against paint plus encode.
	RasterDeltaValid bool `json:"raster_delta_valid"`
	// RasterFull* time Redraw plus PNGWithContext on fresh pages, so paint
	// lands inside the timer on both sides whichever path paints it.
	RasterFullNS        int64 `json:"raster_full_ns"`
	RasterFullControlNS int64 `json:"raster_full_control_ns"`
	RasterFullDeltaNS   int64 `json:"raster_full_delta_ns"`
	// ControlBitmapFallback reports the control page's paint path.
	ControlBitmapFallback bool `json:"control_bitmap_fallback"`
	Ops                   int  `json:"ops"`
	ExtraOps              int  `json:"extra_ops"`
	BitmapFallback        bool `json:"bitmap_fallback"`

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

// animRunDecl gives a uniform or hover fixture a running animation. The
// property declaration follows it, so a longhand under test overrides its
// own component.
const animRunDecl = "animation-name:slide;animation-duration:1s;animation-timing-function:linear;animation-iteration-count:infinite;"

// animNameSupportDecl is everything an animation needs except its name, so
// the animation-name row is what starts it.
const animNameSupportDecl = "animation-duration:1s;animation-timing-function:linear;animation-iteration-count:infinite;"

// animSupportDecl returns the declarations an animation longhand needs to
// run on a uniform or hover fixture. Alone, animation-duration or
// animation-direction declares no animation, and animation-name alone runs
// a zero-second animation, so the row measured nothing. Scroll-driven
// longhands (timeline, range, trigger) are left alone.
func animSupportDecl(name string) string {
	switch strings.TrimPrefix(name, "-webkit-") {
	case "animation-name":
		return animNameSupportDecl
	case "animation-delay", "animation-direction", "animation-duration", "animation-fill-mode",
		"animation-iteration-count", "animation-play-state", "animation-timing-function", "animation-composition":
		return animRunDecl
	}

	return ""
}

// variantDecl is the declaration list a uniform or hover variant applies for
// one property: its support declarations, then the property itself.
func variantDecl(name, sample string) string {
	return animSupportDecl(name) + name + ":" + sample + ";"
}

// coldSpecFor returns the cold-phase demo for a property. The catalog's
// animation-name demo already runs the animation on both sides, so the
// harness uses a base that lacks only the name.
func coldSpecFor(r genProp) DemoSpec {
	if strings.TrimPrefix(r.Name, "-webkit-") == "animation-name" {
		return DemoSpec{HTML: animBox(animNameSupportDecl)}
	}

	return specFor(r)
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
	// writeInt folds a signed value as a varint, so no sign conversion is
	// needed for indices, z-indices, or sticky ids.
	writeInt := func(v int) {
		var buf [binary.MaxVarintLen64]byte
		_, _ = h.Write(binary.AppendVarint(buf[:0], int64(v)))
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
		// Paint order carries z-index and stacking changes that leave
		// every op's own fields alone.
		writeU64(uint64(len(d.Order)))
		for _, idx := range d.Order {
			writeInt(idx)
		}
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
			for _, value := range [...]float64{
				op.RadiusY, op.RadiusTopLeft, op.RadiusTopRight, op.RadiusBottomRight, op.RadiusBottomLeft,
				op.RadiusTopLeftY, op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY,
				op.Opacity(),
			} {
				writeF64(value)
			}
			writeInt(op.ZIndex)
			writeInt(op.StickyID)
			writeU64(boolBits(op.ZIndexSet, op.Fixed, op.Bold, op.Positioned))
			_, _ = h.Write([]byte(op.BlendModeName()))
			writeGroupChain(h, op.Group())
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

// boolBits packs flags into one word for the signature.
func boolBits(flags ...bool) uint64 {
	var bits uint64
	for i, f := range flags {
		if f {
			bits |= 1 << i
		}
	}

	return bits
}

// writeGroupChain hashes the compositing metadata of an op's group chain:
// blend mode, backdrop filter, isolation, and mask presence. Opacity,
// filters, blending, and masks live on the group rather than the op, so an
// op-only signature misses them. Group IDs are left out: they number groups
// in creation order, which is not paint.
func writeGroupChain(h hash.Hash64, g *layout.DisplayGroup) {
	depth := uint64(0)
	for ; g != nil; g = g.Parent {
		depth++
		_, _ = h.Write([]byte(g.Mode))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(g.BackdropFilter))
		var flags [propBenchByteBits]byte
		binary.LittleEndian.PutUint64(flags[:], boolBits(g.Isolate, g.Mask != nil, g.MaskBorder != nil))
		_, _ = h.Write(flags[:])
	}
	var buf [propBenchByteBits]byte
	binary.LittleEndian.PutUint64(buf[:], depth)
	_, _ = h.Write(buf[:])
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
	// A value such as `"liga" 0` carries double quotes that would end the
	// attribute and drop the declaration, so the attribute value is escaped.
	escaped := html.EscapeString(styleDecl)
	for i := range count {
		if styleDecl != "" {
			fmt.Fprintf(&b, `<div id="s%d" class="pitem" style="%s">cell %d</div>`, i, escaped, i)
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
		b.WriteString(variantDecl(propName, sample))
		b.WriteString(`}`)
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

// classifyBottleneckLayer names the layer that pays most of the property's
// added cost. Stage times are variant minus control from the same cold pair,
// so a layer that dominates every page (layout, usually) does not win by
// default. Explicit path changes take precedence: a bitmap fallback is the
// software painter, and a hover full-layout fallback is the page's patch
// scope. A row whose largest delta stays within propBenchNoiseNS is "noise".
func classifyBottleneckLayer(row propDumpRow) string {
	if row.BitmapFallback {
		return "ownframe/replay+bitmap"
	}
	if row.HoverFallbacks > 0 {
		return "ownframe/page+motion"
	}
	raster := int64(0)
	if row.RasterDeltaValid {
		raster = row.RasterDeltaNS
	}
	stages := [...]struct {
		layer string
		delta int64
	}{
		{"blinkless/css", row.StyleNS - row.ControlStyleNS},
		{"blinkless/layout", row.LayoutNS - row.ControlLayoutNS},
		{"ownframe/replay+bitmap", row.PaintNS - row.ControlPaintNS},
		{"ownframe/page+motion", row.OtherNS - row.ControlOtherNS},
		{"ownframe/replay+bitmap", raster},
	}
	best, bestDelta := "noise", int64(propBenchNoiseNS)
	for _, stage := range stages {
		if stage.delta > bestDelta {
			best, bestDelta = stage.layer, stage.delta
		}
	}

	return best
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

func TestClassifyBottleneckLayerUsesStageDeltas(t *testing.T) {
	t.Parallel()
	const ms = 1_000_000
	cases := []struct {
		name string
		row  propDumpRow
		want string
	}{
		{"layout heavy page, no added cost", propDumpRow{LayoutNS: 3 * ms, ControlLayoutNS: 3 * ms}, "noise"},
		{"added style cost", propDumpRow{StyleNS: ms, ControlStyleNS: ms / 10, LayoutNS: 3 * ms, ControlLayoutNS: 3 * ms}, "blinkless/css"},
		{"added layout cost", propDumpRow{LayoutNS: 4 * ms, ControlLayoutNS: 3 * ms}, "blinkless/layout"},
		{"added paint cost", propDumpRow{PaintNS: ms, ControlPaintNS: 0}, "ownframe/replay+bitmap"},
		{"valid raster cost", propDumpRow{RasterDeltaNS: ms, RasterDeltaValid: true}, "ownframe/replay+bitmap"},
		{"invalid raster delta ignored", propDumpRow{RasterDeltaNS: ms}, "noise"},
		{"bitmap fallback wins", propDumpRow{BitmapFallback: true, LayoutNS: 4 * ms}, "ownframe/replay+bitmap"},
		{"hover fallback wins", propDumpRow{HoverFallbacks: 1, LayoutNS: 4 * ms}, "ownframe/page+motion"},
		{"warm allocs no longer force motion", propDumpRow{WarmAllocs: 200}, "noise"},
	}
	for _, c := range cases {
		if got := classifyBottleneckLayer(c.row); got != c.want {
			t.Errorf("%s: layer = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestPropPerfAnimationLonghandsRun checks that animation-name, alone or
// prefixed, now starts an animation the cold fixture's control lacks, and
// that a longhand gets the support declarations on the uniform fixture.
func TestPropPerfAnimationLonghandsRun(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"animation-name", "-webkit-animation-name"} {
		row := measurePropertyFull(ctx, genProp{Name: name, Sample: "slide"}, 1, false)
		if row.Err != "" {
			t.Fatalf("%s: %s", name, row.Err)
		}
		if !slices.Contains(row.EffectEvidence, "motion") {
			t.Fatalf("%s: evidence = %v, want motion", name, row.EffectEvidence)
		}
	}
	if decl := variantDecl("animation-duration", "2s"); !strings.HasPrefix(decl, animRunDecl) ||
		!strings.HasSuffix(decl, "animation-duration:2s;") {
		t.Fatalf("animation-duration decl = %q", decl)
	}
	if decl := variantDecl("color", "red"); decl != "color:red;" {
		t.Fatalf("color decl = %q", decl)
	}
}

// TestDisplaySignatureSeesGroupAndOrder checks the signature reacts to
// effects that live on the compositing group or the paint order.
func TestDisplaySignatureSeesGroupAndOrder(t *testing.T) {
	ctx := context.Background()
	page := func(style string) *ownframe.Page {
		p, err := newBenchPage(propBenchHead()+`<div style="position:relative">`+
			`<div id="a" class="pitem" style="position:relative;`+style+`">a</div>`+
			`<div id="b" class="pitem" style="position:relative;margin-top:-10px">b</div></div>`,
			propBenchWidth, propBenchHeight, false)
		if err != nil {
			t.Fatal(err)
		}
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		return p
	}
	base := displaySignature(page(""))
	for _, style := range []string{"z-index:2", "mix-blend-mode:multiply", "isolation:isolate"} {
		if displaySignature(page(style)) == base {
			t.Errorf("%s left the display signature unchanged", style)
		}
	}
}

// TestPropPerfEngineErrorKeepsPhases checks that a fixture the page refuses
// is labelled an engine error and still runs the phases that build their
// own pages.
func TestPropPerfEngineErrorKeepsPhases(t *testing.T) {
	row := measurePropertyFull(context.Background(), genProp{Name: "font-variant", Sample: "small-caps"}, 1, false)
	if row.Err == "" {
		t.Skip("font-variant now renders; the engine-error path has no fixture")
	}
	if !row.EngineError {
		t.Fatalf("cold failure %q not marked as an engine error", row.Err)
	}
	if row.HoverNS == 0 && row.PhaseErrors["hover"] == "" {
		t.Fatal("hover phase neither ran nor recorded an error")
	}
	if row.Scale64NS == 0 && row.PhaseErrors["scale64"] == "" {
		t.Fatal("scale64 phase neither ran nor recorded an error")
	}
}

// TestUniformScaleDocKeepsQuotedValues pins the attribute escaping: a feature
// list with double quotes must reach the engine on every box instead of ending
// the style attribute.
func TestUniformScaleDocKeepsQuotedValues(t *testing.T) {
	t.Parallel()
	doc := buildUniformScaleDoc(propBenchHead(), `font-feature-settings:"liga" 0;`, 2)
	p, err := newBenchPage(doc, propBenchWidth, propBenchHeight, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, op := range p.Probe().Display().Ops {
		if op.Text != "" && op.FontFeatures() != "" {
			found++
		}
	}
	if found < 2 {
		t.Fatalf("%d text runs carry the feature list, want one per box", found)
	}
}
