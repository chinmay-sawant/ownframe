package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

type comboEntry struct {
	Kind                 string   `json:"kind"`
	Name                 string   `json:"name"`
	Props                []string `json:"props"`
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
	for i := range propBenchHoverNodes {
		fmt.Fprintf(&b, `<div id="h%d" class="pitem" style="%s">item %d</div>`, i, baseStyle, i)
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
	SumScale   int64
	SumHover   int64
	SumRaster  int64
}

func measureCombo(ctx context.Context, spec comboSpec, iters int, highPrecision bool, base sharedBaselines) comboEntry {
	entry := comboEntry{
		Kind:          spec.Kind,
		Name:          spec.Name,
		Props:         spec.Props,
		HighPrecision: highPrecision,
	}

	head := propBenchHead()
	on1 := buildUniformScaleDoc(head, spec.BaseStyle, 1)
	on64 := buildUniformScaleDoc(head, spec.BaseStyle, propBenchScaleNodes)
	hOn := buildComboHoverDoc(head, spec.BaseStyle, spec.HoverStyle)

	// Cold 1-box vs shared baseline.
	cOnT, _, _, _, _, err := measureColdSingle(ctx, on1, propBenchWidth, propBenchHeight, iters, highPrecision, false)
	if err != nil {
		entry.Err = fmt.Sprintf("cold: %v", err)

		return entry
	}
	entry.ColdNS = cOnT.MedianNS
	entry.ColdDeltaNS = cOnT.MedianNS - base.Cold1NS

	// Scaled 64-node vs shared baseline.
	sOnT, sOnA, _, sOnPage, _, err := measureColdSingle(ctx, on64, propBenchScaleWidth, propBenchScaleHeight, iters, highPrecision, true)
	if err != nil {
		entry.Err = fmt.Sprintf("scale64: %v", err)

		return entry
	}
	entry.Scale64NS = sOnT.MedianNS
	entry.Scale64DeltaNS = sOnT.MedianNS - base.Scale64NS
	entry.Scale64Bytes = sOnA.Bytes
	entry.Scale64Allocs = sOnA.Allocs
	if sOnPage != nil {
		if d := sOnPage.Probe().Display(); d != nil {
			entry.Ops = len(d.Ops)
		}
		entry.BitmapFallback = sOnPage.Probe().Image() != nil && sOnPage.Probe().Display() == nil
	}

	// Hover vs shared baseline.
	if hT, _, hStats, _, _, _, hErr := measureSingleHover(ctx, hOn, iters, highPrecision); hErr == nil {
		entry.HoverNS = hT.MedianNS
		entry.HoverDeltaNS = hT.MedianNS - base.Hover64NS
		entry.HoverPatches = hStats.HoverPatches
		entry.HoverFallbacks = hStats.HoverFallbacks
	}

	// Uncached software raster vs shared baseline.
	if rT, _, rErr := measureSingleRaster(ctx, on1, propBenchWidth, propBenchHeight, iters, highPrecision); rErr == nil {
		entry.RasterNS = rT.MedianNS
		entry.RasterDeltaNS = rT.MedianNS - base.Raster1NS
	}

	if spec.SumScale > propBenchMinBaseNS*propBenchComboMinMul && entry.Scale64DeltaNS > propBenchMinBaseNS*propBenchComboMinOut {
		entry.ExpectedSumScale64NS = spec.SumScale
		entry.Scale64Multiplier = float64(entry.Scale64DeltaNS) / float64(spec.SumScale)
	}
	if spec.SumHover > propBenchMinBaseNS*propBenchComboMinMul && entry.HoverDeltaNS > propBenchMinBaseNS*propBenchComboMinOut {
		entry.HoverMultiplier = float64(entry.HoverDeltaNS) / float64(spec.SumHover)
	}
	if spec.SumRaster > propBenchMinBaseNS*propBenchComboMinMul && entry.RasterDeltaNS > propBenchMinBaseNS*propBenchComboMinOut {
		entry.RasterMultiplier = float64(entry.RasterDeltaNS) / float64(spec.SumRaster)
	}
	entry.MaxMultiplier = max(entry.Scale64Multiplier, max(entry.HoverMultiplier, entry.RasterMultiplier))
	entry.SuperLinear = entry.MaxMultiplier >= propBenchSuperLinear

	return entry
}

func runComboMatrix(ctx context.Context, rows []propDumpRow, opts propBenchOpts, base sharedBaselines, progress func(string)) []comboEntry {
	byName := make(map[string]propDumpRow, len(rows))
	for _, r := range rows {
		byName[r.Name] = r
	}

	var specs []comboSpec

	// Layer A: Subsystem-bucket pairwise covering array.
	reps := subsystemBucketReps()
	if opts.MaxBuckets > 0 && opts.MaxBuckets < len(reps) {
		reps = reps[:opts.MaxBuckets]
	}
	singleCache := make(map[string]comboEntry, len(reps))
	for _, rep := range reps {
		decl := rep.Prop.Name + ":" + rep.Prop.Sample + ";"
		singleCache[rep.Bucket] = measureCombo(ctx, comboSpec{
			Kind:       "bucket-single",
			Name:       rep.Bucket,
			Props:      []string{rep.Prop.Name},
			BaseStyle:  decl,
			HoverStyle: decl,
		}, 1, false, base)
	}
	for i := range reps {
		for j := i + 1; j < len(reps); j++ {
			a, b := reps[i], reps[j]
			if a.Prop.Name == b.Prop.Name {
				continue
			}
			sa, sb := singleCache[a.Bucket], singleCache[b.Bucket]
			decl := a.Prop.Name + ":" + a.Prop.Sample + ";" + b.Prop.Name + ":" + b.Prop.Sample + ";"
			specs = append(specs, comboSpec{
				Kind:       "bucket-pair",
				Name:       a.Bucket + "+" + b.Bucket,
				Props:      []string{a.Prop.Name, b.Prop.Name},
				BaseStyle:  decl,
				HoverStyle: decl,
				SumScale:   max(0, sa.Scale64DeltaNS) + max(0, sb.Scale64DeltaNS),
				SumHover:   max(0, sa.HoverDeltaNS) + max(0, sb.HoverDeltaNS),
				SumRaster:  max(0, sa.RasterDeltaNS) + max(0, sb.RasterDeltaNS),
			})
		}
	}

	// Layer B: Auto-selected Top-M offender pairwise from Tier 1.
	if opts.TopPairs >= propBenchHalfDiv {
		candidates := make([]propDumpRow, 0, len(rows))
		for _, r := range rows {
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
				decl := a.Name + ":" + a.Sample + ";" + b.Name + ":" + b.Sample + ";"
				specs = append(specs, comboSpec{
					Kind:       "top-offender-pair",
					Name:       a.Name + "+" + b.Name,
					Props:      []string{a.Name, b.Name},
					BaseStyle:  decl,
					HoverStyle: decl,
					SumScale:   max(0, a.Scale64DeltaNS) + max(0, b.Scale64DeltaNS),
					SumHover:   max(0, a.HoverDeltaNS) + max(0, b.HoverDeltaNS),
					SumRaster:  max(0, a.RasterDeltaNS) + max(0, b.RasterDeltaNS),
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
		for _, r := range rows {
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
				var sumScale, sumHover, sumRaster int64
				for _, p := range picked {
					sb.WriteString(p.Name)
					sb.WriteString(":")
					sb.WriteString(p.Sample)
					sb.WriteString(";")
					names = append(names, p.Name)
					sumScale += max(0, byName[p.Name].Scale64DeltaNS)
					sumHover += max(0, byName[p.Name].HoverDeltaNS)
					sumRaster += max(0, byName[p.Name].RasterDeltaNS)
				}
				decl := sb.String()
				specs = append(specs, comboSpec{
					Kind:       "seeded-fuzz",
					Name:       fmt.Sprintf("fuzz-%02d", idx+1),
					Props:      names,
					BaseStyle:  decl,
					HoverStyle: decl,
					SumScale:   sumScale,
					SumHover:   sumHover,
					SumRaster:  sumRaster,
				})
			}
		}
	}

	// Pass 1: Fast sweep across all combinations.
	out := make([]comboEntry, len(specs))
	for i, sp := range specs {
		out[i] = measureCombo(ctx, sp, 1, false, base)
		if progress != nil && ((i+1)%propBenchLogEvery == 0 || i+1 == len(specs)) {
			progress(fmt.Sprintf("[tier2 sweep %d/%d] last=%s:%s (scale64=%.2fms hover=%.2fms)",
				i+1, len(specs), sp.Kind, sp.Name, float64(out[i].Scale64NS)/propNSPerMS, float64(out[i].HoverNS)/propNSPerMS))
		}
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
			out[idx] = measureCombo(ctx, specs[idx], min(propBenchComboRefineIt, opts.Iters), true, base)
		}
	}

	return out
}
