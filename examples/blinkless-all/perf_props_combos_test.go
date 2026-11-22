package main

import (
	"context"
	"fmt"
	"html"
	"math/rand/v2"
	"slices"
	"strings"
)

type comboEntry struct {
	Kind                 string   `json:"kind"`
	Name                 string   `json:"name"`
	Props                []string `json:"props"`
	BaseStyle            string   `json:"base_style"`
	HoverStyle           string   `json:"hover_style"`
	HighPrecision        bool     `json:"high_precision"`
	ColdNS               int64    `json:"cold_ns"`
	ColdDeltaNS          int64    `json:"cold_delta_ns"`
	HoverNS              int64    `json:"hover_ns"`
	HoverDeltaNS         int64    `json:"hover_delta_ns"`
	HoverPatches         uint64   `json:"hover_patches"`
	HoverFallbacks       uint64   `json:"hover_fallbacks"`
	Scale64NS            int64    `json:"scale64_ns"`
	Scale64DeltaNS       int64    `json:"scale64_delta_ns"`
	Scale64Bytes         uint64   `json:"scale64_bytes"`
	Scale64Allocs        uint64   `json:"scale64_allocs"`
	RasterNS             int64    `json:"raster_ns"`
	RasterDeltaNS        int64    `json:"raster_delta_ns"`
	Ops                  int      `json:"ops"`
	BitmapFallback       bool     `json:"bitmap_fallback"`
	ExpectedSumScale64NS int64    `json:"expected_sum_scale64_ns,omitempty"`
	Scale64Multiplier    float64  `json:"scale64_multiplier,omitempty"`
	HoverMultiplier      float64  `json:"hover_multiplier,omitempty"`
	RasterMultiplier     float64  `json:"raster_multiplier,omitempty"`
	MaxMultiplier        float64  `json:"max_multiplier,omitempty"`
	SuperLinear          bool     `json:"super_linear"`
	Err                  string   `json:"error,omitempty"`

	// Iters is the timed samples per workload behind every delta above, and
	// Refined marks an entry the high-precision pass re-measured. Only refined
	// entries carry enough samples for a regression gate.
	Iters   int  `json:"iters"`
	Refined bool `json:"refined"`
	// The controls are measured interleaved with this combination, on the
	// same fixture with the declarations removed; each delta is against them.
	ColdControlNS    int64 `json:"cold_control_ns"`
	Scale64ControlNS int64 `json:"scale64_control_ns"`
	HoverControlNS   int64 `json:"hover_control_ns"`
	RasterControlNS  int64 `json:"raster_control_ns"`
	// The expected sums are the positive single-declaration deltas on the
	// same fixtures, at the same sample count, that each multiplier divides.
	ExpectedSumHoverNS  int64 `json:"expected_sum_hover_ns,omitempty"`
	ExpectedSumRasterNS int64 `json:"expected_sum_raster_ns,omitempty"`
}

type propDecl struct {
	Name   string
	Sample string
}

// subsystemBucketReps picks one representative property per rendering/layout
// mechanism so pairwise covering tests all cross-subsystem interactions.
func subsystemBucketReps() []struct {
	Bucket string
	Prop   propDecl
} {
	return []struct {
		Bucket string
		Prop   propDecl
	}{
		{"flex", propDecl{"display", "flex"}},
		{"grid", propDecl{"display", "grid"}},
		{"position", propDecl{"position", "relative"}},
		{"overflow_clip", propDecl{"overflow", "hidden"}},
		{"border_radius", propDecl{"border-radius", "12px"}},
		{"bg_gradient", propDecl{"background-image", "linear-gradient(#c0392b, #2980b9)"}},
		{"shadow", propDecl{"box-shadow", "0 8px 24px rgba(0,0,0,0.35)"}},
		{"filter_backdrop", propDecl{"backdrop-filter", "blur(12px)"}},
		{"filter_fx", propDecl{"filter", "blur(4px)"}},
		{"blend_opacity", propDecl{"mix-blend-mode", "multiply"}},
		{"opacity", propDecl{"opacity", "0.75"}},
		{"transform", propDecl{"transform", "translateY(-2px) scale(1.03)"}},
		{"typography", propDecl{"letter-spacing", "1.5px"}},
		{"clip_shape", propDecl{"clip-path", "inset(4px round 8px)"}},
		{"transition", propDecl{"transition", "background-color 160ms ease, transform 160ms ease"}},
	}
}

type uiRecipe struct {
	Name       string
	BaseStyle  string
	HoverStyle string
	Props      []string
}

func realWorldUIRecipes() []uiRecipe {
	return []uiRecipe{
		{
			Name:       "glassmorphic-card",
			BaseStyle:  "display:flex;flex-direction:column;gap:4px;border-radius:12px;background:rgba(255,255,255,0.14);border:1px solid rgba(255,255,255,0.25);backdrop-filter:blur(12px);box-shadow:0 8px 24px rgba(0,0,0,0.35)",
			HoverStyle: "transform:translateY(-2px);background:rgba(255,255,255,0.22)",
			Props:      []string{"display", "border-radius", "backdrop-filter", "box-shadow", "transform"},
		},
		{
			Name:       "sticky-table-hover",
			BaseStyle:  "display:flex;align-items:center;justify-content:space-between;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;border-bottom:1px solid #333;transition:background-color 120ms ease",
			HoverStyle: "background-color:#262626;color:#ffffff",
			Props:      []string{"display", "overflow", "text-overflow", "border-bottom", "transition"},
		},
		{
			Name:       "dashboard-grid-shadow",
			BaseStyle:  "display:grid;grid-template-columns:1fr 1fr;gap:4px;border-radius:10px;box-shadow:0 4px 12px rgba(0,0,0,0.25);padding:6px",
			HoverStyle: "border-color:#2563eb;box-shadow:0 8px 20px rgba(37,99,235,0.25)",
			Props:      []string{"display", "grid-template-columns", "border-radius", "box-shadow", "border-color"},
		},
		{
			Name:       "scrollable-feed-clip",
			BaseStyle:  "overflow:hidden;clip-path:inset(2px round 8px);opacity:0.92;background:#1e293b;color:#f8fafc",
			HoverStyle: "opacity:1;background:#334155",
			Props:      []string{"overflow", "clip-path", "opacity", "background"},
		},
		{
			Name:       "fixed-modal-backdrop",
			BaseStyle:  "position:relative;z-index:10;border-radius:14px;background:rgba(15,23,42,0.85);backdrop-filter:blur(8px);box-shadow:0 12px 32px rgba(0,0,0,0.45)",
			HoverStyle: "border-color:#38bdf8",
			Props:      []string{"position", "z-index", "border-radius", "backdrop-filter", "box-shadow"},
		},
		{
			Name:       "code-block-typography",
			BaseStyle:  "font-family:monospace;font-size:12px;line-height:1.5;letter-spacing:0.4px;white-space:pre-wrap;text-shadow:0 1px 1px rgba(0,0,0,0.2);border-left:3px solid #22c55e",
			HoverStyle: "background-color:#0f172a;color:#e2e8f0",
			Props:      []string{"font-family", "line-height", "letter-spacing", "white-space", "text-shadow", "border-left"},
		},
		{
			Name:       "animated-hover-cards",
			BaseStyle:  "border-radius:10px;background:#181818;color:#e5e7eb;transition:color 160ms ease, background-color 160ms ease, transform 160ms ease;will-change:transform, opacity",
			HoverStyle: "background-color:#303030;transform:translateY(-2px)",
			Props:      []string{"border-radius", "transition", "will-change", "transform"},
		},
		{
			Name:       "gradient-badge-strip",
			BaseStyle:  "display:inline-flex;align-items:center;border-radius:999px;background-image:linear-gradient(90deg, #ef4444, #3b82f6);color:#fff;font-weight:700;letter-spacing:0.04em",
			HoverStyle: "opacity:0.85;transform:scale(1.04)",
			Props:      []string{"display", "border-radius", "background-image", "font-weight", "letter-spacing", "transform"},
		},
		{
			Name:       "blend-mode-hero",
			BaseStyle:  "position:relative;border-radius:12px;background-image:radial-gradient(#f59e0b, #7c3aed);mix-blend-mode:multiply;opacity:0.85",
			HoverStyle: "opacity:1;filter:blur(2px)",
			Props:      []string{"position", "border-radius", "background-image", "mix-blend-mode", "opacity", "filter"},
		},
		{
			Name:       "outline-focus-form",
			BaseStyle:  "border:1px solid #94a3b8;border-radius:6px;outline:2px solid #2563eb;outline-offset:2px;box-shadow:0 1px 2px rgba(0,0,0,0.1)",
			HoverStyle: "border-color:#2563eb;background-color:#eff6ff",
			Props:      []string{"border", "border-radius", "outline", "outline-offset", "box-shadow"},
		},
		{
			Name:       "mask-gradient-fade",
			BaseStyle:  "overflow:hidden;mask-image:linear-gradient(black, transparent);border-radius:8px;background:#0f172a;color:#f8fafc",
			HoverStyle: "background:#1e293b",
			Props:      []string{"overflow", "mask-image", "border-radius"},
		},
		{
			Name:       "transform-perspective-deck",
			BaseStyle:  "perspective:400px;transform:rotate(3deg) scale(0.98);border-radius:10px;box-shadow:0 6px 16px rgba(0,0,0,0.3);transition:transform 160ms ease",
			HoverStyle: "transform:rotate(0deg) scale(1.03)",
			Props:      []string{"perspective", "transform", "border-radius", "box-shadow", "transition"},
		},
	}
}

func buildComboHoverDoc(head, baseStyle, hoverStyle string) string {
	var b strings.Builder
	b.WriteString(head)
	b.WriteString(`<style>#h0:hover{`)
	b.WriteString(hoverStyle)
	b.WriteString(`}</style><div class="pgrid">`)
	escaped := html.EscapeString(baseStyle)
	for i := range propBenchHoverNodes {
		fmt.Fprintf(&b, `<div id="h%d" class="pitem" style="%s">item %d</div>`, i, escaped, i)
	}
	b.WriteString(`</div>`)

	return b.String()
}

type comboSpec struct {
	Kind       string
	Name       string
	Props      []string
	BaseStyle  string
	HoverStyle string
	// Singles lists the component declarations ("name:value;") whose summed
	// single deltas form the interaction multiplier's denominator. Empty for
	// a recipe, which has no additive expectation.
	Singles []string
}

// comboMultiplierMinNS is the smallest expected sum a multiplier divides by.
// Below it the single deltas sit inside timer noise and the ratio is
// meaningless, so the multiplier stays 0 and cannot set super_linear.
const comboMultiplierMinNS = 50_000

// comboMultiplier returns delta/sum when sum clears comboMultiplierMinNS and
// delta is positive, and 0 otherwise.
func comboMultiplier(delta, sum int64) float64 {
	if sum <= comboMultiplierMinNS || delta <= 0 {
		return 0
	}

	return float64(delta) / float64(sum)
}

// comboSums are the expected additive deltas of one combination's singles.
type comboSums struct {
	scale, hover, raster int64
}

// applyComboMultipliers fills the multipliers and super_linear from sums.
func applyComboMultipliers(entry *comboEntry, sums comboSums) {
	entry.ExpectedSumScale64NS = sums.scale
	entry.ExpectedSumHoverNS = sums.hover
	entry.ExpectedSumRasterNS = sums.raster
	entry.Scale64Multiplier = comboMultiplier(entry.Scale64DeltaNS, sums.scale)
	entry.HoverMultiplier = comboMultiplier(entry.HoverDeltaNS, sums.hover)
	entry.RasterMultiplier = comboMultiplier(entry.RasterDeltaNS, sums.raster)
	entry.MaxMultiplier = max(entry.Scale64Multiplier, max(entry.HoverMultiplier, entry.RasterMultiplier))
	entry.SuperLinear = entry.MaxMultiplier >= propBenchSuperLinear
}

type singleKey struct {
	decl          string
	iters         int
	highPrecision bool
}

// comboSingleCache measures each single declaration once per sample count
// on the combination fixtures, so a pair and its singles share the fixture,
// the sample count, and the GC isolation.
type comboSingleCache struct {
	measure func(spec comboSpec, iters int, highPrecision bool) comboEntry
	entries map[singleKey]comboEntry
}

func newComboSingleCache(measure func(comboSpec, int, bool) comboEntry) *comboSingleCache {
	return &comboSingleCache{measure: measure, entries: map[singleKey]comboEntry{}}
}

func (c *comboSingleCache) get(decl string, iters int, highPrecision bool) comboEntry {
	key := singleKey{decl: decl, iters: iters, highPrecision: highPrecision}
	if e, ok := c.entries[key]; ok {
		return e
	}
	e := c.measure(comboSpec{Kind: "single", Name: decl, BaseStyle: decl, HoverStyle: decl}, iters, highPrecision)
	c.entries[key] = e

	return e
}

// sums adds the positive deltas of spec's singles. A failed single counts as
// zero, which only lowers the sum below the multiplier floor.
func (c *comboSingleCache) sums(spec comboSpec, iters int, highPrecision bool) comboSums {
	var s comboSums
	for _, decl := range spec.Singles {
		e := c.get(decl, iters, highPrecision)
		if e.Err != "" {
			continue
		}
		s.scale += max(0, e.Scale64DeltaNS)
		s.hover += max(0, e.HoverDeltaNS)
		s.raster += max(0, e.RasterDeltaNS)
	}

	return s
}

// measureComboWithSingles measures spec and divides its deltas by its
// singles measured the same way.
func measureComboWithSingles(ctx context.Context, singles *comboSingleCache, spec comboSpec, iters int, highPrecision bool) comboEntry {
	entry := measureCombo(ctx, spec, iters, highPrecision)
	if entry.Err == "" && len(spec.Singles) > 0 {
		applyComboMultipliers(&entry, singles.sums(spec, iters, highPrecision))
	}

	return entry
}

// measureCombo measures one combination against controls timed in the same
// loop: the same fixture with no declarations, so host load and heap state
// hit both sides alike.
func measureCombo(ctx context.Context, spec comboSpec, iters int, highPrecision bool) comboEntry {
	iters = max(1, iters)
	entry := comboEntry{
		Kind:          spec.Kind,
		Name:          spec.Name,
		Props:         spec.Props,
		BaseStyle:     spec.BaseStyle,
		HoverStyle:    spec.HoverStyle,
		HighPrecision: highPrecision,
		Refined:       highPrecision,
		Iters:         iters,
	}

	head := propBenchHead()
	off1 := buildUniformScaleDoc(head, "", 1)
	on1 := buildUniformScaleDoc(head, spec.BaseStyle, 1)

	cOff, cOn, err := measureColdPair(ctx, off1, on1, propBenchWidth, propBenchHeight, iters, highPrecision)
	if err != nil {
		entry.Err = fmt.Sprintf("cold: %v", err)

		return entry
	}
	entry.ColdNS = cOn.timing.MedianNS
	entry.ColdControlNS = cOff.timing.MedianNS
	entry.ColdDeltaNS = entry.ColdNS - entry.ColdControlNS

	sOff, sOn, err := measureColdPair(ctx,
		buildUniformScaleDoc(head, "", propBenchScaleNodes),
		buildUniformScaleDoc(head, spec.BaseStyle, propBenchScaleNodes),
		propBenchScaleWidth, propBenchScaleHeight, iters, highPrecision)
	if err != nil {
		entry.Err = fmt.Sprintf("scale64: %v", err)

		return entry
	}
	entry.Scale64NS = sOn.timing.MedianNS
	entry.Scale64ControlNS = sOff.timing.MedianNS
	entry.Scale64DeltaNS = entry.Scale64NS - entry.Scale64ControlNS
	entry.Scale64Bytes = sOn.allocs.Bytes
	entry.Scale64Allocs = sOn.allocs.Allocs
	if sOn.page != nil {
		if d := sOn.page.Probe().Display(); d != nil {
			entry.Ops = len(d.Ops)
		}
		entry.BitmapFallback = sOn.page.Probe().Image() != nil && sOn.page.Probe().Display() == nil
	}

	measureComboHover(ctx, &entry, head, spec, iters, highPrecision)

	// Uncached software raster, control and variant in one rotation.
	if r, rErr := measureRasterWorkloads(ctx, propBenchWidth, propBenchHeight, iters, highPrecision, off1, on1); rErr == nil {
		entry.RasterNS = r[1].timing.MedianNS
		entry.RasterControlNS = r[0].timing.MedianNS
		entry.RasterDeltaNS = entry.RasterNS - entry.RasterControlNS
	}

	return entry
}

// measureComboHover times hover cycles on the 64-node fixture with and
// without the declarations. A failure leaves the hover fields zero, as the
// shared-baseline version did.
func measureComboHover(ctx context.Context, entry *comboEntry, head string, spec comboSpec, iters int, highPrecision bool) {
	off, err := newHoverWorkload(ctx, buildComboHoverDoc(head, "", ""))
	if err != nil {
		return
	}
	on, err := newHoverWorkload(ctx, buildComboHoverDoc(head, spec.BaseStyle, spec.HoverStyle))
	if err != nil {
		return
	}
	timing, err := measureWorkloadsNS(iters, highPrecision,
		func() error { return off.cycle(ctx) },
		func() error { return on.cycle(ctx) })
	if err != nil {
		return
	}
	stats := on.p.Probe().Stats()
	entry.HoverNS = timing[1].MedianNS
	entry.HoverControlNS = timing[0].MedianNS
	entry.HoverDeltaNS = entry.HoverNS - entry.HoverControlNS
	entry.HoverPatches = stats.HoverPatches
	entry.HoverFallbacks = stats.HoverFallbacks
}

func runComboMatrix(
	ctx context.Context,
	rows, selectionRows []propDumpRow,
	reference []comboEntry,
	opts propBenchOpts,
	_ sharedBaselines,
	progress func(string),
) ([]comboEntry, error) {
	// Every combination now times its own interleaved control, so the
	// process-start shared baselines are no longer a denominator.
	singles := newComboSingleCache(func(spec comboSpec, iters int, highPrecision bool) comboEntry {
		return measureCombo(ctx, spec, iters, highPrecision)
	})

	var specs []comboSpec

	// Layer A: Subsystem-bucket pairwise covering array.
	reps := subsystemBucketReps()
	if opts.MaxBuckets > 0 && opts.MaxBuckets < len(reps) {
		reps = reps[:opts.MaxBuckets]
	}
	for i := range reps {
		for j := i + 1; j < len(reps); j++ {
			a, b := reps[i], reps[j]
			if a.Prop.Name == b.Prop.Name {
				continue
			}
			da, db := a.Prop.Name+":"+a.Prop.Sample+";", b.Prop.Name+":"+b.Prop.Sample+";"
			specs = append(specs, comboSpec{
				Kind:       "bucket-pair",
				Name:       a.Bucket + "+" + b.Bucket,
				Props:      []string{a.Prop.Name, b.Prop.Name},
				BaseStyle:  da + db,
				HoverStyle: da + db,
				Singles:    []string{da, db},
			})
		}
	}

	// Layer B: Auto-selected Top-M offender pairwise from Tier 1.
	if opts.TopPairs >= propBenchHalfDiv {
		candidates := make([]propDumpRow, 0, len(rows))
		for _, r := range selectionRows {
			if r.Err == "" && r.Effective {
				candidates = append(candidates, r)
			}
		}
		slices.SortFunc(candidates, func(a, b propDumpRow) int {
			scoreA := max(0, a.Scale64DeltaNS) + max(0, a.HoverDeltaNS)*propBenchScoreWeight + max(0, a.RasterDeltaNS)*propBenchScoreWeight
			scoreB := max(0, b.Scale64DeltaNS) + max(0, b.HoverDeltaNS)*propBenchScoreWeight + max(0, b.RasterDeltaNS)*propBenchScoreWeight
			if scoreA > scoreB {
				return -1
			}
			if scoreA < scoreB {
				return 1
			}

			return strings.Compare(a.Name, b.Name)
		})
		topN := min(opts.TopPairs, len(candidates))
		top := candidates[:topN]
		for i := range top {
			for j := i + 1; j < len(top); j++ {
				a, b := top[i], top[j]
				da, db := a.Name+":"+a.Sample+";", b.Name+":"+b.Sample+";"
				specs = append(specs, comboSpec{
					Kind:       "top-offender-pair",
					Name:       a.Name + "+" + b.Name,
					Props:      []string{a.Name, b.Name},
					BaseStyle:  da + db,
					HoverStyle: da + db,
					Singles:    []string{da, db},
				})
			}
		}
	}

	// Layer C: Real-world UI archetype recipes.
	recipes := realWorldUIRecipes()
	if opts.MaxRecipes > 0 && opts.MaxRecipes < len(recipes) {
		recipes = recipes[:opts.MaxRecipes]
	}
	for _, rec := range recipes {
		specs = append(specs, comboSpec{
			Kind:       "ui-recipe",
			Name:       rec.Name,
			Props:      rec.Props,
			BaseStyle:  rec.BaseStyle,
			HoverStyle: rec.HoverStyle,
		})
	}

	// Layer D: Deterministic seeded multi-property fuzz stacks (math/rand/v2).
	if opts.FuzzCount > 0 {
		pool := make([]propDumpRow, 0, len(rows))
		for _, r := range selectionRows {
			if r.Err == "" && r.Effective && !strings.HasPrefix(r.Name, "-webkit-") {
				pool = append(pool, r)
			}
		}
		if len(pool) >= propBenchFuzzSize {
			rng := rand.New(rand.NewPCG(opts.Seed, opts.Seed^propBenchPCGStreamXOR)) //nolint:gosec // reason: deterministic non-cryptographic benchmark seed
			for idx := range opts.FuzzCount {
				picked := make([]propDumpRow, 0, propBenchFuzzSize)
				used := make(map[string]bool, propBenchFuzzSize)
				for len(picked) < propBenchFuzzSize {
					cand := pool[rng.IntN(len(pool))]
					if used[cand.Name] {
						continue
					}
					used[cand.Name] = true
					picked = append(picked, cand)
				}
				var sb strings.Builder
				names := make([]string, 0, len(picked))
				decls := make([]string, 0, len(picked))
				for _, p := range picked {
					d := p.Name + ":" + p.Sample + ";"
					sb.WriteString(d)
					names = append(names, p.Name)
					decls = append(decls, d)
				}
				decl := sb.String()
				specs = append(specs, comboSpec{
					Kind:       "seeded-fuzz",
					Name:       fmt.Sprintf("fuzz-%02d", idx+1),
					Props:      names,
					BaseStyle:  decl,
					HoverStyle: decl,
					Singles:    decls,
				})
			}
		}
	}

	if opts.Reference != "" && len(reference) != len(specs) {
		return nil, fmt.Errorf("reference has %d combinations, generated %d", len(reference), len(specs))
	}

	// Pass 1: Fast sweep across all combinations.
	out := make([]comboEntry, len(specs))
	for i, sp := range specs {
		if opts.Reference != "" {
			prior := reference[i]
			if prior.Kind != sp.Kind || prior.Name != sp.Name || !slices.Equal(prior.Props, sp.Props) ||
				prior.BaseStyle != sp.BaseStyle || prior.HoverStyle != sp.HoverStyle {
				return nil, fmt.Errorf("reference fixture mismatch at %s %q", sp.Kind, sp.Name)
			}
		}
		out[i] = measureComboWithSingles(ctx, singles, sp, 1, false)
		if progress != nil && ((i+1)%propBenchLogEvery == 0 || i+1 == len(specs)) {
			progress(fmt.Sprintf("[tier2 sweep %d/%d] last=%s:%s (scale64=%.2fms hover=%.2fms)",
				i+1, len(specs), sp.Kind, sp.Name, float64(out[i].Scale64NS)/propNSPerMS, float64(out[i].HoverNS)/propNSPerMS))
		}
	}

	if opts.Reference != "" {
		for i, prior := range reference {
			if prior.HighPrecision {
				out[i] = measureComboWithSingles(ctx, singles, specs[i], min(propBenchComboRefineIt, opts.Iters), true)
			}
		}

		return out, nil
	}

	// Pass 2: High-precision confirmation on top super-linear / heavy combos.
	if opts.Iters > 1 && len(out) > 0 {
		type idxScore struct {
			idx   int
			score float64
		}
		ranked := make([]idxScore, 0, len(out))
		for i, c := range out {
			if c.Err == "" && (c.SuperLinear || c.Kind == "ui-recipe") {
				ranked = append(ranked, idxScore{idx: i, score: c.MaxMultiplier*propPctScale + float64(c.Scale64NS)/propNSPerMS})
			}
		}
		slices.SortFunc(ranked, func(a, b idxScore) int {
			if a.score > b.score {
				return -1
			}
			if a.score < b.score {
				return 1
			}

			return 0
		})
		refineN := min(propBenchComboRefineN, len(ranked))
		for i := range refineN {
			idx := ranked[i].idx
			out[idx] = measureComboWithSingles(ctx, singles, specs[idx], min(propBenchComboRefineIt, opts.Iters), true)
		}
	}

	return out, nil
}
