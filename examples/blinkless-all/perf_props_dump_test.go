package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	propDirPerm         = 0o750
	propFilePerm        = 0o600
	propSummaryTopN     = 25
	propPprofMinCPU     = 120 * time.Millisecond
	propPprofMaxIters   = 40
	propPprofMinIters   = 2
	propPprofMemIters   = 10
	propPprofNodeCount  = 6
	propPprofFieldsMin  = 6
	propPprofFuncCol    = 5
	propPprofCumCol     = 3
	propPprofMaxFrames  = 5
	propPprofPhaseCount = 3
	propRefineLogEvery  = 10
	propSmokePropCount  = 3
	propSmokeBuckets    = 2
	propSmokeRecipes    = 1
	propNSPerMS         = 1e6
	propBytesPerKB      = 1024.0
	propPctScale        = 100.0
	propDriftAlertPct   = 15.0
	propPprofCmdTimeout = 15 * time.Second
)

type pprofAttribution struct {
	Phase     string   `json:"phase"`
	Property  string   `json:"property"`
	Layer     string   `json:"layer"`
	CPUFile   string   `json:"cpu_file"`
	MemFile   string   `json:"mem_file"`
	CPUFrames []string `json:"cpu_frames"`
	MemFrames []string `json:"mem_frames"`
}

func sanitizeFilePart(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			b.WriteRune(ch)
		} else {
			b.WriteByte('_')
		}
	}

	return b.String()
}

func captureOffenderProfile(ctx context.Context, profDir, phase string, r propDumpRow) pprofAttribution {
	attr := pprofAttribution{
		Phase:    phase,
		Property: r.Name,
		Layer:    r.Layer,
	}
	if err := os.MkdirAll(profDir, propDirPerm); err != nil {
		return attr
	}

	head := propBenchHead()
	prop := genProp{Name: r.Name, Sample: r.Sample, Group: r.Group}
	spec := specFor(prop)

	var workload func() error
	switch phase {
	case "hover":
		work, err := newHoverWorkload(ctx, buildHoverDoc(head, r.Name, r.Sample, true))
		if err != nil {
			return attr
		}
		workload = func() error { return work.cycle(ctx) }
	case "raster":
		doc := head + spec.expand(prop, "p", true)
		workload = func() error {
			p, err := newBenchPage(doc, propBenchWidth, propBenchHeight, false)
			if err != nil {
				return err
			}
			if err := p.Redraw(ctx); err != nil {
				return err
			}
			_ = p.Probe().PNGWithContext(ctx)

			return nil
		}
	default:
		doc := buildUniformScaleDoc(head, r.Name+":"+r.Sample+";", propBenchScaleNodes)
		workload = func() error {
			p, err := newBenchPage(doc, propBenchScaleWidth, propBenchScaleHeight, false)
			if err != nil {
				return err
			}

			return p.Redraw(ctx)
		}
	}

	base := fmt.Sprintf("%s-%s", sanitizeFilePart(phase), sanitizeFilePart(r.Name))
	cpuPath := filepath.Clean(filepath.Join(profDir, base+".cpu.pprof"))
	memPath := filepath.Clean(filepath.Join(profDir, base+".mem.pprof"))

	if f, err := os.OpenFile(cpuPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, propFilePerm); err == nil {
		if startErr := pprof.StartCPUProfile(f); startErr == nil {
			deadline := time.Now().Add(propPprofMinCPU)
			for i := 0; i < propPprofMaxIters; i++ {
				if workload() != nil {
					break
				}
				if i >= propPprofMinIters && time.Now().After(deadline) {
					break
				}
			}
			pprof.StopCPUProfile()
			attr.CPUFile = filepath.Base(cpuPath)
		}
		_ = f.Close()
	}

	runtime.GC()
	for range propPprofMemIters {
		if workload() != nil {
			break
		}
	}
	if f, err := os.OpenFile(memPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, propFilePerm); err == nil {
		if pprof.Lookup("allocs").WriteTo(f, 0) == nil {
			attr.MemFile = filepath.Base(memPath)
		}
		_ = f.Close()
	}

	if attr.CPUFile != "" {
		attr.CPUFrames = extractPprofTop(ctx, cpuPath, "")
	}
	if attr.MemFile != "" {
		attr.MemFrames = extractPprofTop(ctx, memPath, "-alloc_space")
	}

	return attr
}

func extractPprofTop(ctx context.Context, profPath, sampleFlag string) []string {
	tctx, cancel := context.WithTimeout(ctx, propPprofCmdTimeout)
	defer cancel()

	args := []string{"tool", "pprof", "-top", fmt.Sprintf("-nodecount=%d", propPprofNodeCount*propBenchHalfDiv)}
	if sampleFlag != "" {
		args = append(args, sampleFlag)
	}
	args = append(args, filepath.Clean(profPath))

	cmd := exec.CommandContext(tctx, "go", args...) //nolint:gosec // reason: fixed go tool pprof invocation on local profile path
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var frames []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	inTable := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "flat") && strings.Contains(line, "cum") {
			inTable = true
			continue
		}
		if !inTable || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < propPprofFieldsMin {
			continue
		}
		fnName := strings.Join(fields[propPprofFuncCol:], " ")
		if strings.HasPrefix(fnName, "runtime.") || strings.HasPrefix(fnName, "testing.") ||
			strings.Contains(fnName, "examples/blinkless-all") {
			continue
		}
		frames = append(frames, fmt.Sprintf("%s (flat %s, cum %s)", fnName, fields[0], fields[propPprofCumCol]))
		if len(frames) >= propPprofMaxFrames {
			break
		}
	}

	return frames
}

func writeJSONL[T any](path string, items []T) error {
	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, propFilePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	bw := bufio.NewWriter(f)
	enc := json.NewEncoder(bw)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}

	return bw.Flush()
}

func readPropJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var rows []T
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var row T
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("decode %s row %d: %w", path, len(rows)+1, err)
		}
		rows = append(rows, row)
	}

	return rows, scanner.Err()
}

func writeBenchscoreTxt(path string, rows []propDumpRow, combos []comboEntry) error {
	var b strings.Builder
	fmt.Fprintf(&b, "goos: %s\ngoarch: %s\npkg: github.com/chinmay-sawant/ownframe/examples/blinkless-all\n", runtime.GOOS, runtime.GOARCH)
	for _, r := range rows {
		if r.Err != "" {
			continue
		}
		id := sanitizeFilePart(r.Name)
		for _, phase := range []struct {
			name   string
			ns     int64
			bytes  uint64
			allocs uint64
		}{
			{name: "Cold", ns: r.ColdNS, bytes: r.ColdBytes, allocs: r.ColdAllocs},
			{name: "Hover", ns: r.HoverNS, bytes: r.HoverBytes, allocs: r.HoverAllocs},
			{name: "Scale64", ns: r.Scale64NS, bytes: r.Scale64Bytes, allocs: r.Scale64Allocs},
			{name: "Raster", ns: r.RasterNS, bytes: r.RasterBytes, allocs: r.RasterAllocs},
		} {
			if phase.ns <= 0 {
				continue
			}
			fmt.Fprintf(&b, "BenchmarkProp%s_%s-1\t1\t%d ns/op\t%d B/op\t%d allocs/op\n",
				phase.name, id, phase.ns, phase.bytes, phase.allocs)
		}
	}
	for _, c := range combos {
		if c.Err != "" {
			continue
		}
		id := sanitizeFilePart(c.Kind + "_" + c.Name)
		fmt.Fprintf(&b, "BenchmarkComboScale64_%s-1\t1\t%d ns/op\t%d B/op\t%d allocs/op\n", id, max(1, c.Scale64NS), c.Scale64Bytes, c.Scale64Allocs)
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), propFilePerm)
}

func writeSummaryMarkdown(
	path string,
	rows []propDumpRow,
	combos []comboEntry,
	attrs []pprofAttribution,
	opts propBenchOpts,
	driftPct float64,
	elapsed time.Duration,
) error {
	var b strings.Builder
	effectiveCount, noopCount, warmLeaks, hoverFallbacks, bitmapFallbacks, errCount, refinedCount := 0, 0, 0, 0, 0, 0, 0
	layerCounts := map[string]int{}
	phaseErrors := 0
	for _, r := range rows {
		if r.Err != "" {
			errCount++
			continue
		}
		phaseErrors += len(r.PhaseErrors)
		if r.Effective {
			effectiveCount++
		} else {
			noopCount++
		}
		if r.HighPrecision {
			refinedCount++
		}
		if r.WarmAllocs > 0 {
			warmLeaks++
		}
		if r.HoverFallbacks > 0 {
			hoverFallbacks++
		}
		if r.BitmapFallback {
			bitmapFallbacks++
		}
		layerCounts[r.Layer]++
	}

	b.WriteString("# CSS Property Performance Dump Summary\n\n")
	fmt.Fprintf(&b, "- **Generated:** %s (`%s/%s`, Go `%s`, `%d` CPUs)\n", time.Now().Format(time.RFC3339), runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.NumCPU())
	fmt.Fprintf(&b, "- **Measurement version:** 2 (interleaved property controls with untimed effect and PNG validation)\n")
	fmt.Fprintf(&b, "- **Properties audited:** **%d** (`%d` observed effects, `%d` effects unverified, `%d` refined at up to `%d` iters, `%d` failed rows, `%d` phase errors)\n", len(rows), effectiveCount, noopCount, refinedCount, opts.Iters, errCount, phaseErrors)
	fmt.Fprintf(&b, "- **Combinations audited:** **%d** (bucket pairs, top-offender pairs, UI recipes, seeded fuzz `seed=%d`)\n", len(combos), opts.Seed)
	fmt.Fprintf(&b, "- **Wall time:** `%s` | **Control drift:** `%.1f%%`\n", elapsed.Round(time.Millisecond), driftPct)
	if driftPct > propDriftAlertPct {
		fmt.Fprintf(&b, "- **Warning:** Control anchor drifted by `%.1f%%` (> %.0f%%); consider re-running on an idle machine.\n", driftPct, propDriftAlertPct)
	}
	fmt.Fprintf(&b, "- **Flags:** warm redraws with allocations=**%d**, hover full-layout fallbacks=**%d**, bitmap fallbacks=**%d**\n\n", warmLeaks, hoverFallbacks, bitmapFallbacks)

	b.WriteString("## Bottleneck Layer Distribution\n\n")
	b.WriteString("| Layer | Properties |\n|---|---:|\n")
	for _, layer := range []string{"blinkless/css", "blinkless/layout", "ownframe/page+motion", "ownframe/replay+bitmap", "balanced"} {
		fmt.Fprintf(&b, "| **%s** | %d |\n", layer, layerCounts[layer])
	}
	b.WriteString("\n")

	writePropTable := func(title string, sorted []propDumpRow, metric func(propDumpRow) string) {
		fmt.Fprintf(&b, "## %s\n\n", title)
		b.WriteString("| # | Property | Group | Metric | Cold Delta | Hover Delta | Scale64 Delta | Raster Delta | Layer |\n")
		b.WriteString("|---:|---|---|---:|---:|---:|---:|---:|---|\n")
		n := min(propSummaryTopN, len(sorted))
		for i := range n {
			r := sorted[i]
			fmt.Fprintf(&b, "| %d | `%s` | %s | **%s** | %.2f ms | %.2f ms | %.2f ms | %.2f ms | `%s` |\n",
				i+1, r.Name, r.Group, metric(r),
				float64(r.ColdDeltaNS)/propNSPerMS,
				float64(r.HoverDeltaNS)/propNSPerMS,
				float64(r.Scale64DeltaNS)/propNSPerMS,
				float64(r.RasterDeltaNS)/propNSPerMS,
				r.Layer,
			)
		}
		b.WriteString("\n")
	}

	valid := make([]propDumpRow, 0, len(rows))
	for _, r := range rows {
		if r.Err == "" {
			valid = append(valid, r)
		}
	}

	byScale := slices.Clone(valid)
	slices.SortFunc(byScale, func(a, b propDumpRow) int { return cmpDescInt64(a.Scale64DeltaNS, b.Scale64DeltaNS) })
	writePropTable("Top 25 by Scale-64 Paired Delta (Uniform 64-Node Grid)", byScale, func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%.1f KB, ratio %.1fx)", float64(r.Scale64DeltaNS)/propNSPerMS, float64(r.Scale64Bytes)/propBytesPerKB, r.ScalingRatio)
	})

	byHover := slices.Clone(valid)
	slices.SortFunc(byHover, func(a, b propDumpRow) int { return cmpDescInt64(a.HoverDeltaNS, b.HoverDeltaNS) })
	writePropTable("Top 25 by Hover Paired Delta (64-Node Retained Container)", byHover, func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%d allocs, fb=%d)", float64(r.HoverDeltaNS)/propNSPerMS, r.HoverAllocs, r.HoverFallbacks)
	})

	byRaster := slices.Clone(valid)
	slices.SortFunc(byRaster, func(a, b propDumpRow) int { return cmpDescInt64(a.RasterDeltaNS, b.RasterDeltaNS) })
	writePropTable("Top 25 by Software Raster Paired Delta", byRaster, func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%.1f KB)", float64(r.RasterDeltaNS)/propNSPerMS, float64(r.RasterBytes)/propBytesPerKB)
	})

	byResize := slices.Clone(valid)
	slices.SortFunc(byResize, func(a, b propDumpRow) int { return cmpDescInt64(a.ResizeDeltaNS, b.ResizeDeltaNS) })
	writePropTable("Top 25 by Resize Relayout Paired Delta", byResize, func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%d allocs)", float64(r.ResizeDeltaNS)/propNSPerMS, r.ResizeAllocs)
	})

	// Hover fallbacks, warm leaks, bitmap fallbacks.
	b.WriteString("## Hover Full-Layout Fallbacks\n\n")
	if hoverFallbacks == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("| Property | Group | Fallbacks | Hover Time | Failure Reasons |\n|---|---|---:|---:|---|\n")
		for _, r := range valid {
			if r.HoverFallbacks > 0 {
				fmt.Fprintf(&b, "| `%s` | %s | **%d** | %.2f ms | `%s` |\n",
					r.Name, r.Group, r.HoverFallbacks, float64(r.HoverNS)/propNSPerMS, strings.Join(r.HoverFailureReasons, ", "))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("## Warm Redraw Allocations and Bitmap Fallbacks\n\n")
	if warmLeaks == 0 && bitmapFallbacks == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("| Property | Group | Warm Allocs | Bitmap Fallback | Sample |\n|---|---|---:|---|---|\n")
		for _, r := range valid {
			if r.WarmAllocs > 0 || r.BitmapFallback {
				fmt.Fprintf(&b, "| `%s` | %s | **%d** | %t | `%s` |\n", r.Name, r.Group, r.WarmAllocs, r.BitmapFallback, r.Sample)
			}
		}
		b.WriteString("\n")
	}

	// Tier 2 Combinations: super-linear multipliers + UI recipes.
	b.WriteString("## Tier 2: Super-Linear Combinations (Multiplier >= 1.5x)\n\n")
	super := make([]comboEntry, 0)
	for _, c := range combos {
		if c.SuperLinear && c.Err == "" {
			super = append(super, c)
		}
	}
	slices.SortFunc(super, func(a, b comboEntry) int {
		if a.MaxMultiplier > b.MaxMultiplier {
			return -1
		}
		if a.MaxMultiplier < b.MaxMultiplier {
			return 1
		}

		return strings.Compare(a.Name, b.Name)
	})
	if len(super) == 0 {
		b.WriteString("No super-linear combinations above 1.5x detected.\n\n")
	} else {
		b.WriteString("| # | Kind | Combination | Max Mult | Scale64 | Hover | Raster | Hover FB |\n|---:|---|---|---:|---:|---:|---:|---:|\n")
		for i := range min(propSummaryTopN, len(super)) {
			c := super[i]
			fmt.Fprintf(&b, "| %d | %s | `%s` | **%.2fx** | %.2f ms | %.2f ms | %.2f ms | %d |\n",
				i+1, c.Kind, c.Name, c.MaxMultiplier,
				float64(c.Scale64NS)/propNSPerMS, float64(c.HoverNS)/propNSPerMS, float64(c.RasterNS)/propNSPerMS, c.HoverFallbacks)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Tier 2: Real-World UI Recipes (64 Nodes)\n\n")
	b.WriteString("| Recipe | Scale64 | Scale64 KB | Hover | Hover FB | Raster | Ops | Bitmap FB |\n|---|---:|---:|---:|---:|---:|---:|---|\n")
	for _, c := range combos {
		if c.Kind == "ui-recipe" {
			fmt.Fprintf(&b, "| `%s` | **%.2f ms** | %.1f KB | %.2f ms | %d | %.2f ms | %d | %t |\n",
				c.Name, float64(c.Scale64NS)/propNSPerMS, float64(c.Scale64Bytes)/propBytesPerKB,
				float64(c.HoverNS)/propNSPerMS, c.HoverFallbacks, float64(c.RasterNS)/propNSPerMS, c.Ops, c.BitmapFallback)
		}
	}
	b.WriteString("\n")

	// Automated pprof stack attribution.
	b.WriteString("## Top Offender `pprof` Stack Attribution\n\n")
	if len(attrs) == 0 {
		b.WriteString("No profiles captured.\n\n")
	} else {
		for _, a := range attrs {
			fmt.Fprintf(&b, "### `[%s]` `%s` (`%s`)\n\n", a.Phase, a.Property, a.Layer)
			if len(a.CPUFrames) > 0 {
				b.WriteString("- **CPU top frames** (`profiles/" + a.CPUFile + "`):\n")
				for _, fr := range a.CPUFrames {
					fmt.Fprintf(&b, "  - `%s`\n", fr)
				}
			}
			if len(a.MemFrames) > 0 {
				b.WriteString("- **Alloc top frames** (`profiles/" + a.MemFile + "`):\n")
				for _, fr := range a.MemFrames {
					fmt.Fprintf(&b, "  - `%s`\n", fr)
				}
			}
			b.WriteString("\n")
		}
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), propFilePerm)
}

func cmpDescInt64(a, b int64) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}

	return 0
}

func selectRefinementIndices(rows []propDumpRow, limit int) []int {
	if limit <= 0 || len(rows) == 0 {
		return nil
	}

	// Each phase ranks its own candidates. Choose across rankings before
	// returning to any phase, so a long list of warm flags cannot starve raster.
	scores := []func(propDumpRow) int64{
		func(r propDumpRow) int64 { return r.HoverDeltaNS },
		func(r propDumpRow) int64 { return r.Scale64DeltaNS },
		func(r propDumpRow) int64 { return r.RasterDeltaNS },
		func(r propDumpRow) int64 { return r.ResizeDeltaNS },
		func(r propDumpRow) int64 { return r.ColdDeltaNS },
		func(r propDumpRow) int64 { return r.WarmNS },
		func(r propDumpRow) int64 { return r.TickNS },
	}
	rankings := make([][]int, len(scores))
	for phase, score := range scores {
		for idx, row := range rows {
			if row.Err == "" && row.Effective && score(row) > 0 {
				rankings[phase] = append(rankings[phase], idx)
			}
		}
		slices.SortStableFunc(rankings[phase], func(a, b int) int {
			return cmpDescInt64(score(rows[a]), score(rows[b]))
		})
	}

	chosen := make(map[int]bool, limit)
	order := make([]int, 0, limit)
	for rank := 0; rank < len(rows) && len(order) < limit; rank++ {
		for _, phase := range rankings {
			if rank >= len(phase) || chosen[phase[rank]] {
				continue
			}
			idx := phase[rank]
			chosen[idx] = true
			order = append(order, idx)
			if len(order) == limit {
				return order
			}
		}
	}

	return order
}

func runPropPerfDump(ctx context.Context, opts propBenchOpts, logf func(string, ...any)) (int, int, error) {
	if err := os.MkdirAll(filepath.Clean(opts.DumpDir), propDirPerm); err != nil {
		return 0, 0, fmt.Errorf("mkdir dump dir: %w", err)
	}
	log := func(format string, args ...any) {
		if logf != nil {
			logf(format, args...)
		}
	}

	tStart := time.Now()
	allProps := genProps()
	filtered := make([]genProp, 0, len(allProps))
	for _, r := range allProps {
		if opts.Filter != nil && !opts.Filter.MatchString(r.Name) {
			continue
		}
		filtered = append(filtered, r)
		if opts.Limit > 0 && len(filtered) >= opts.Limit {
			break
		}
	}

	comboPlanRows, comboPlan, err := loadPropReference(opts.Reference, filtered)
	if err != nil {
		return 0, 0, err
	}

	// Prime process caches & measure shared uniform baselines once.
	base, err := measureSharedBaselines(ctx, max(1, opts.Iters))
	if err != nil {
		return 0, 0, fmt.Errorf("measure shared baselines: %w", err)
	}
	log("[init] baselines: cold1=%.2fms hover64=%.2fms scale64=%.2fms raster1=%.2fms (auditing %d properties)",
		float64(base.Cold1NS)/propNSPerMS, float64(base.Hover64NS)/propNSPerMS,
		float64(base.Scale64NS)/propNSPerMS, float64(base.Raster1NS)/propNSPerMS, len(filtered))

	anchorDoc := propBenchHead() + `<div id="p" class="gbox">anchor</div>`
	anchorRun := func() error {
		p, pErr := newBenchPage(anchorDoc, propBenchWidth, propBenchHeight, false)
		if pErr != nil {
			return pErr
		}

		return p.Redraw(ctx)
	}
	anchorStart, err := measureTimedNS(opts.Iters, true, anchorRun)
	if err != nil {
		return 0, 0, fmt.Errorf("anchor warmup: %w", err)
	}
	maxDriftPct := 0.0

	// Pass 1: Fast paired-control sweep over every property.
	selected := make([]propDumpRow, len(filtered))
	slowScale, slowHover, slowRaster := propDumpRow{}, propDumpRow{}, propDumpRow{}
	for i, r := range filtered {
		if i > 0 && i%propBenchDriftEvery == 0 {
			runtime.GC()
			if cur, dErr := measureTimedNS(opts.Iters, false, anchorRun); dErr == nil && anchorStart.MedianNS > 0 {
				pct := float64(absInt64(cur.MedianNS-anchorStart.MedianNS)) * propPctScale / float64(anchorStart.MedianNS)
				if pct > maxDriftPct {
					maxDriftPct = pct
				}
			}
		}
		row := measurePropertyFull(ctx, r, 1, false)
		selected[i] = row
		if row.Scale64DeltaNS > slowScale.Scale64DeltaNS {
			slowScale = row
		}
		if row.HoverDeltaNS > slowHover.HoverDeltaNS {
			slowHover = row
		}
		if row.RasterDeltaNS > slowRaster.RasterDeltaNS {
			slowRaster = row
		}
		if (i+1)%propBenchLogEvery == 0 || i+1 == len(filtered) {
			log("[tier1 sweep %d/%d] elapsed=%s top_scale=%s(%.2fms) top_hover=%s(%.2fms) top_raster=%s(%.2fms)",
				i+1, len(filtered), time.Since(tStart).Round(time.Millisecond),
				slowScale.Name, float64(slowScale.Scale64DeltaNS)/propNSPerMS,
				slowHover.Name, float64(slowHover.HoverDeltaNS)/propNSPerMS,
				slowRaster.Name, float64(slowRaster.RasterDeltaNS)/propNSPerMS)
		}
	}

	// Pass 2: High-precision GC-isolated refinement on Top-K outliers.
	if opts.Iters > 1 && (opts.RefineTop > 0 || opts.Reference != "") {
		refineIdxs := selectRefinementIndices(selected, opts.RefineTop)
		if opts.Reference != "" {
			refineIdxs = refineIdxs[:0]
			for idx, row := range comboPlanRows {
				if row.HighPrecision {
					refineIdxs = append(refineIdxs, idx)
				}
			}
		}
		log("[tier1 refine] re-measuring top %d outliers at %d GC-isolated iters...", len(refineIdxs), opts.Iters)
		for k, idx := range refineIdxs {
			selected[idx] = measurePropertyFull(ctx, filtered[idx], opts.Iters, true)
			if (k+1)%propRefineLogEvery == 0 || k+1 == len(refineIdxs) {
				log("[tier1 refine %d/%d] elapsed=%s last=%s", k+1, len(refineIdxs), time.Since(tStart).Round(time.Millisecond), filtered[idx].Name)
			}
		}
	}

	// Write Tier-1 JSONL immediately so partial results are already on disk.
	if err := writeJSONL(filepath.Join(opts.DumpDir, "props-individual.jsonl"), selected); err != nil {
		return 0, 0, err
	}

	// Tier 2: Combination matrix.
	if opts.Reference == "" {
		comboPlanRows = selected
	}
	combos, err := runComboMatrix(ctx, selected, comboPlanRows, comboPlan, opts, base, func(msg string) {
		log("%s elapsed=%s", msg, time.Since(tStart).Round(time.Millisecond))
	})
	if err != nil {
		return 0, 0, err
	}
	if err := writeJSONL(filepath.Join(opts.DumpDir, "props-combos.jsonl"), combos); err != nil {
		return 0, 0, err
	}

	// Pass 3: Capture pprof profiles for top offenders across scale64, hover, and raster.
	var attrs []pprofAttribution
	if opts.PprofTop > 0 {
		profDir := filepath.Join(opts.DumpDir, "profiles")
		seen := map[string]bool{}
		pickTop := func(phase string, score func(propDumpRow) int64, limit int) {
			cands := make([]propDumpRow, 0, len(selected))
			for _, r := range selected {
				if r.Err == "" && r.Effective {
					cands = append(cands, r)
				}
			}
			slices.SortFunc(cands, func(a, b propDumpRow) int { return cmpDescInt64(score(a), score(b)) })
			added := 0
			for _, r := range cands {
				if added >= limit {
					break
				}
				key := phase + ":" + r.Name
				if seen[key] {
					continue
				}
				seen[key] = true
				log("[pprof] capturing %s for %s...", phase, r.Name)
				attrs = append(attrs, captureOffenderProfile(ctx, profDir, phase, r))
				added++
			}
		}
		perPhase := max(1, opts.PprofTop/propPprofPhaseCount)
		pickTop("scale64", func(r propDumpRow) int64 { return r.Scale64DeltaNS }, perPhase)
		pickTop("hover", func(r propDumpRow) int64 { return r.HoverDeltaNS }, perPhase)
		pickTop("raster", func(r propDumpRow) int64 { return r.RasterDeltaNS }, max(1, opts.PprofTop-propBenchHalfDiv*perPhase))
	}

	if err := writeBenchscoreTxt(filepath.Join(opts.DumpDir, "props-bench.txt"), selected, combos); err != nil {
		return 0, 0, err
	}
	if err := writeSummaryMarkdown(filepath.Join(opts.DumpDir, "summary.md"), selected, combos, attrs, opts, maxDriftPct, time.Since(tStart)); err != nil {
		return 0, 0, err
	}

	return len(selected), len(combos), nil
}

func loadPropReference(dir string, props []genProp) ([]propDumpRow, []comboEntry, error) {
	if dir == "" {
		return nil, nil, nil
	}
	rows, err := readPropJSONL[propDumpRow](filepath.Join(dir, "props-individual.jsonl"))
	if err != nil {
		return nil, nil, fmt.Errorf("read reference properties: %w", err)
	}
	if len(rows) != len(props) {
		return nil, nil, fmt.Errorf("reference has %d properties, selected %d", len(rows), len(props))
	}
	for idx, row := range rows {
		sample, _ := activeSample(props[idx])
		if row.Name != props[idx].Name || row.Sample != sample {
			return nil, nil, fmt.Errorf("reference fixture mismatch at property %q", props[idx].Name)
		}
	}
	combos, err := readPropJSONL[comboEntry](filepath.Join(dir, "props-combos.jsonl"))
	if err != nil {
		return nil, nil, fmt.Errorf("read reference combinations: %w", err)
	}

	return rows, combos, nil
}

func absInt64(v int64) int64 {
	if v < 0 {
		return -v
	}

	return v
}

// TestPropPerfDump executes the full 688-property paired-control benchmark,
// Tier-2 combination matrix, and top-offender pprof attribution when
// OWNFRAME_PROP_DUMP_DIR is set (see scripts/perf-props-dump.sh).
func TestPropPerfDump(t *testing.T) {
	opts, err := loadPropBenchOpts()
	if err != nil {
		t.Fatal(err)
	}
	if opts.DumpDir == "" {
		t.Skip("set OWNFRAME_PROP_DUMP_DIR or run scripts/perf-props-dump.sh to generate a property performance dump")
	}

	nProps, nCombos, err := runPropPerfDump(context.Background(), opts, func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Wrote %d properties + %d combinations to %s", nProps, nCombos, opts.DumpDir)
}

// TestPropPerfHarnessSmoke exercises the paired-control property harness,
// combination matrix, benchscore output, and summary writer on a tiny subset
// during normal `make test` so CI always guards the profiling tooling.
func TestPropPerfHarnessSmoke(t *testing.T) {
	dir := t.TempDir()
	opts := propBenchOpts{
		DumpDir:    dir,
		Iters:      propBenchHalfDiv,
		RefineTop:  propBenchHalfDiv,
		TopPairs:   propBenchHalfDiv,
		FuzzCount:  0,
		Seed:       propBenchDefaultSeed,
		PprofTop:   0,
		Filter:     regexp.MustCompile(`^(border-radius|box-shadow|backdrop-filter)$`),
		Limit:      propSmokePropCount,
		MaxBuckets: propSmokeBuckets,
		MaxRecipes: propSmokeRecipes,
	}

	nProps, nCombos, err := runPropPerfDump(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("runPropPerfDump smoke: %v", err)
	}
	if nProps != propSmokePropCount {
		t.Fatalf("audited %d properties, want %d", nProps, propSmokePropCount)
	}
	if nCombos == 0 {
		t.Fatal("expected non-zero combination entries")
	}

	for _, name := range []string{"props-individual.jsonl", "props-combos.jsonl", "props-bench.txt", "summary.md"} {
		info, statErr := os.Stat(filepath.Join(dir, name))
		if statErr != nil || info.Size() == 0 {
			t.Fatalf("expected non-empty %s (err=%v)", name, statErr)
		}
	}

	rows, err := readPropJSONL[propDumpRow](filepath.Join(dir, "props-individual.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Err != "" || len(row.PhaseErrors) > 0 {
			t.Fatalf("%s: row error %q, phase errors %v", row.Name, row.Err, row.PhaseErrors)
		}
		if row.HoverControlNS <= 0 || row.Scale64ControlNS <= 0 || row.RasterControlNS <= 0 {
			t.Fatalf("%s: missing a paired control", row.Name)
		}
	}

	opts.DumpDir, opts.Reference = t.TempDir(), dir
	matchedProps, matchedCombos, err := runPropPerfDump(context.Background(), opts, nil)
	if err != nil {
		t.Fatal(err)
	}
	if matchedProps != nProps || matchedCombos != nCombos {
		t.Fatalf("matched dump has %d properties and %d combinations, want %d and %d",
			matchedProps, matchedCombos, nProps, nCombos)
	}
}

func TestPropPerfEffectEvidence(t *testing.T) {
	ctx := context.Background()
	inactive := measurePropertyFull(ctx, genProp{Name: "--unreferenced-bench-value", Sample: "1"}, 1, false)
	if inactive.Err != "" || len(inactive.PhaseErrors) > 0 {
		t.Fatalf("inactive fixture: %q, %v", inactive.Err, inactive.PhaseErrors)
	}
	if inactive.Effective || len(inactive.EffectEvidence) != 0 {
		t.Fatalf("inactive fixture reported effects: %v", inactive.EffectEvidence)
	}

	active := measurePropertyFull(ctx, genProp{Name: "color", Sample: "rgb(0, 128, 255)"}, 1, false)
	if active.Err != "" || len(active.PhaseErrors) > 0 {
		t.Fatalf("active fixture: %q, %v", active.Err, active.PhaseErrors)
	}
	if !active.Effective || !slices.Contains(active.EffectEvidence, "hover-display") {
		t.Fatalf("active hover fixture failed to report its effect: %v", active.EffectEvidence)
	}
}

func TestPropPerfRejectsInvalidPNG(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("invalid PNG")} {
		if err := validateRasterPNG(data); err == nil {
			t.Fatalf("accepted invalid PNG %q", data)
		}
	}
}
