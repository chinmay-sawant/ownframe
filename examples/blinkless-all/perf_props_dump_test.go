package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	propDriftWindow     = 100
	propNoisePercentile = 95
	// propMeasurementVersion names the sampling rules a dump follows. A
	// reference dump must carry the same version, so a before and after
	// pair never mixes fixtures, controls, or iteration rules.
	propMeasurementVersion = 3
	propManifestName       = "manifest.json"
	propPprofCmdTimeout    = 15 * time.Second
)

type pprofAttribution struct {
	Phase    string `json:"phase"`
	Property string `json:"property"`
	Layer    string `json:"layer"`
	CPUFile  string `json:"cpu_file"`
	MemFile  string `json:"mem_file"`
	// MemBaseFile is the allocs snapshot taken before the workload. MemFrames
	// subtract it, so they cover only the profiled workload.
	MemBaseFile string   `json:"mem_base_file,omitempty"`
	CPUFrames   []string `json:"cpu_frames"`
	MemFrames   []string `json:"mem_frames"`
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
		// buildHoverDoc already adds the support declarations variantDecl
		// gives an animation longhand.
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
		// The profiled fixture is the measured one, including the support
		// declarations an animation longhand needs to run.
		doc := buildUniformScaleDoc(head, variantDecl(r.Name, r.Sample), propBenchScaleNodes)
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

	// The allocs profile is cumulative since the process started, so the
	// workload's allocations are the difference between a snapshot taken
	// before it and one taken after it. Each snapshot follows a GC because
	// the runtime publishes allocation records at the end of a cycle.
	memBasePath := filepath.Clean(filepath.Join(profDir, base+".mem-base.pprof"))
	baseOK := writeAllocsProfile(memBasePath)
	for range propPprofMemIters {
		if workload() != nil {
			break
		}
	}
	if baseOK && writeAllocsProfile(memPath) {
		attr.MemFile = filepath.Base(memPath)
		attr.MemBaseFile = filepath.Base(memBasePath)
	}

	if attr.CPUFile != "" {
		attr.CPUFrames = extractPprofTop(ctx, pprofTopArgs(cpuPath, "", ""))
	}
	if attr.MemFile != "" {
		attr.MemFrames = extractPprofTop(ctx, pprofTopArgs(memPath, "-alloc_space", memBasePath))
	}

	return attr
}

// writeAllocsProfile collects garbage and writes the cumulative allocs
// profile to path, reporting whether the file was written.
func writeAllocsProfile(path string) bool {
	runtime.GC()
	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, propFilePerm)
	if err != nil {
		return false
	}
	writeErr := pprof.Lookup("allocs").WriteTo(f, 0)

	return f.Close() == nil && writeErr == nil
}

// pprofTopArgs builds the go tool pprof -top invocation. A non-empty basePath
// subtracts that profile, which turns two cumulative allocs snapshots into the
// allocations made between them.
func pprofTopArgs(profPath, sampleFlag, basePath string) []string {
	args := []string{"tool", "pprof", "-top", fmt.Sprintf("-nodecount=%d", propPprofNodeCount*propBenchHalfDiv)}
	if sampleFlag != "" {
		args = append(args, sampleFlag)
	}
	if basePath != "" {
		args = append(args, "-base", filepath.Clean(basePath))
	}

	return append(args, filepath.Clean(profPath))
}

func extractPprofTop(ctx context.Context, args []string) []string {
	tctx, cancel := context.WithTimeout(ctx, propPprofCmdTimeout)
	defer cancel()

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

// writeBenchscoreTxt emits the refined measurements for scripts/cmd/benchscore. A
// sweep row is one sample and its noise exceeds the gate's budget, so only
// rows refined at opts.Iters (HighPrecision) are written, with the real
// iteration count of each phase. A clamped or non-positive result is skipped
// rather than gated. Compare runs made with --reference: both then refine the
// same rows, and benchscore fails a row missing from the candidate.
func writeBenchscoreTxt(path string, rows []propDumpRow, combos []comboEntry, opts propBenchOpts) error {
	var b strings.Builder
	fmt.Fprintf(&b, "goos: %s\ngoarch: %s\npkg: github.com/chinmay-sawant/ownframe/examples/blinkless-all\n", runtime.GOOS, runtime.GOARCH)
	fullIters := max(1, opts.Iters)
	halfIters := max(1, fullIters/propBenchHalfDiv)
	for _, r := range rows {
		if r.Err != "" || !r.HighPrecision {
			continue
		}
		id := sanitizeFilePart(r.Name)
		for _, phase := range []struct {
			name   string
			iters  int
			ns     int64
			bytes  uint64
			allocs uint64
		}{
			{name: "Cold", iters: fullIters, ns: r.ColdNS, bytes: r.ColdBytes, allocs: r.ColdAllocs},
			{name: "Hover", iters: fullIters, ns: r.HoverNS, bytes: r.HoverBytes, allocs: r.HoverAllocs},
			{name: "Scale64", iters: halfIters, ns: r.Scale64NS, bytes: r.Scale64Bytes, allocs: r.Scale64Allocs},
			{name: "Raster", iters: halfIters, ns: r.RasterNS, bytes: r.RasterBytes, allocs: r.RasterAllocs},
			{name: "RasterFull", iters: halfIters, ns: r.RasterFullNS},
		} {
			// A PNG-only time from a bitmap-fallback page measures encoding
			// alone, so it is gated only when control and variant share a path.
			if phase.ns <= 0 || (phase.name == "Raster" && !r.RasterDeltaValid) {
				continue
			}
			fmt.Fprintf(&b, "BenchmarkProp%s_%s-1\t%d\t%d ns/op\t%d B/op\t%d allocs/op\n",
				phase.name, id, phase.iters, phase.ns, phase.bytes, phase.allocs)
		}
	}
	for _, c := range combos {
		if c.Err != "" || !c.Refined || c.Scale64NS <= 0 {
			continue
		}
		id := sanitizeFilePart(c.Kind + "_" + c.Name)
		fmt.Fprintf(&b, "BenchmarkComboScale64_%s-1\t%d\t%d ns/op\t%d B/op\t%d allocs/op\n",
			id, max(1, c.Iters), c.Scale64NS, c.Scale64Bytes, c.Scale64Allocs)
	}

	return os.WriteFile(filepath.Clean(path), []byte(b.String()), propFilePerm)
}

// propDrift reports timing drift over one sweep two ways. The anchor
// measures one fixed page identically (warm-up, GC, timed loop) at the start
// and every propDriftWindow properties; AnchorMedianPct is the median
// deviation from the start and AnchorMaxPct the worst. The control drift
// reads the sweep's own paired controls: the median cold and scale64 control
// of each window of properties against the first window. The warning keys off
// the control drift, because those controls are what every delta subtracts.
type propDrift struct {
	AnchorMedianPct float64
	AnchorMaxPct    float64
	ColdControlPct  float64
	ScaleControlPct float64
	Windows         int
}

// controlPct is the drift the warning keys off. Only the scale64 control is
// the same document for every property; each cold control is that
// property's own fixture without the declaration, so its window medians
// move with the property groups rather than with the host.
func (d propDrift) controlPct() float64 { return d.ScaleControlPct }

// sweepControl is the pass-1 control timing of one property, kept before
// refinement replaces the row so the control drift compares like samples.
type sweepControl struct {
	cold, scale int64
	ok          bool
}

// controlDrift returns the largest deviation of a per-window median control
// from the first window's median, in percent, and the number of windows.
func controlDrift(controls []sweepControl, pick func(sweepControl) int64) (float64, int) {
	var medians []int64
	for start := 0; start < len(controls); start += propDriftWindow {
		var window []int64
		for _, c := range controls[start:min(len(controls), start+propDriftWindow)] {
			if v := pick(c); c.ok && v > 0 {
				window = append(window, v)
			}
		}
		if len(window) > 0 {
			medians = append(medians, medianInt64(window))
		}
	}
	if len(medians) == 0 || medians[0] <= 0 {
		return 0, len(medians)
	}
	worst := 0.0
	for _, m := range medians[1:] {
		worst = max(worst, float64(absInt64(m-medians[0]))*propPctScale/float64(medians[0]))
	}

	return worst, len(medians)
}

func medianInt64(values []int64) int64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	n := len(sorted)
	if n%propBenchHalfDiv == 0 {
		return (sorted[n/propBenchHalfDiv-1] + sorted[n/propBenchHalfDiv]) / propBenchHalfDiv
	}

	return sorted[n/propBenchHalfDiv]
}

func medianFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	n := len(sorted)
	if n%propBenchHalfDiv == 0 {
		return (sorted[n/propBenchHalfDiv-1] + sorted[n/propBenchHalfDiv]) / propBenchHalfDiv
	}

	return sorted[n/propBenchHalfDiv]
}

// noiseFloor is the propNoisePercentile of |delta| over successful rows that
// showed no effect. Those rows carry no real cost, so their spread is the
// timing noise of a delta measured the same way.
func noiseFloor(rows []propDumpRow, delta func(propDumpRow) int64) int64 {
	var spread []int64
	for _, r := range rows {
		if r.Err == "" && !r.Effective {
			spread = append(spread, absInt64(delta(r)))
		}
	}
	if len(spread) == 0 {
		return 0
	}
	slices.Sort(spread)
	idx := (len(spread)*propNoisePercentile - 1) / int(propPctScale)

	return spread[max(0, min(len(spread)-1, idx))]
}

func writeSummaryMarkdown(
	path string,
	rows []propDumpRow,
	combos []comboEntry,
	attrs []pprofAttribution,
	opts propBenchOpts,
	drift propDrift,
	elapsed time.Duration,
) error {
	var b strings.Builder
	effectiveCount, noopCount, warmLeaks, hoverFallbacks, bitmapFallbacks, errCount, refinedCount := 0, 0, 0, 0, 0, 0, 0
	tickAllocRows := 0
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
		if r.TickAllocs > 0 {
			tickAllocRows++
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
	fmt.Fprintf(&b, "- **Measurement version:** %d (interleaved property controls, delta-attributed layers, workload-only allocation profiles)\n", propMeasurementVersion)
	fmt.Fprintf(&b, "- **Properties audited:** **%d** (`%d` observed effects, `%d` effects unverified, `%d` refined at up to `%d` iters, `%d` failed rows, `%d` phase errors)\n", len(rows), effectiveCount, noopCount, refinedCount, opts.Iters, errCount, phaseErrors)
	fmt.Fprintf(&b, "- **Combinations audited:** **%d** (bucket pairs, top-offender pairs, UI recipes, seeded fuzz `seed=%d`)\n", len(combos), opts.Seed)
	fmt.Fprintf(&b, "- **Wall time:** `%s`\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(&b, "- **Control drift:** scale64 `%.1f%%` (worst per-%d-property window median of the uniform 64-node control against the first, %d windows); cold `%.1f%%` is informational, since each cold control is a different fixture\n",
		drift.ScaleControlPct, propDriftWindow, drift.Windows, drift.ColdControlPct)
	fmt.Fprintf(&b, "- **Anchor drift:** median `%.1f%%`, max `%.1f%%`\n", drift.AnchorMedianPct, drift.AnchorMaxPct)
	if drift.controlPct() > propDriftAlertPct {
		fmt.Fprintf(&b, "- **Warning:** paired controls drifted by `%.1f%%` (> %.0f%%); consider re-running on an idle machine.\n", drift.controlPct(), propDriftAlertPct)
	}
	fmt.Fprintf(&b, "- **Flags:** unchanged redraws with allocations=**%d**, ticks with allocations=**%d**, hover full-layout fallbacks=**%d**, bitmap fallbacks=**%d**\n",
		warmLeaks, tickAllocRows, hoverFallbacks, bitmapFallbacks)

	phaseDeltas := []struct {
		name  string
		delta func(propDumpRow) int64
	}{
		{"cold", func(r propDumpRow) int64 { return r.ColdDeltaNS }},
		{"hover", func(r propDumpRow) int64 { return r.HoverDeltaNS }},
		{"scale64", func(r propDumpRow) int64 { return r.Scale64DeltaNS }},
		{"raster", func(r propDumpRow) int64 { return r.RasterFullDeltaNS }},
		{"resize", func(r propDumpRow) int64 { return r.ResizeDeltaNS }},
	}
	floors := make(map[string]int64, len(phaseDeltas))
	floorText := make([]string, 0, len(phaseDeltas))
	for _, phase := range phaseDeltas {
		floors[phase.name] = noiseFloor(rows, phase.delta)
		floorText = append(floorText, fmt.Sprintf("%s `%.2f ms`", phase.name, float64(floors[phase.name])/propNSPerMS))
	}
	fmt.Fprintf(&b, "- **Noise floor** (p%d of |delta| over the %d rows with no observed effect): %s\n\n",
		propNoisePercentile, noopCount, strings.Join(floorText, ", "))

	b.WriteString("## Bottleneck Layer Distribution\n\n")
	b.WriteString("| Layer | Properties |\n|---|---:|\n")
	for _, layer := range []string{"blinkless/css", "blinkless/layout", "ownframe/page+motion", "ownframe/replay+bitmap", "noise"} {
		fmt.Fprintf(&b, "| **%s** | %d |\n", layer, layerCounts[layer])
	}
	b.WriteString("\n")

	// Only rows with an observed effect are ranked: a row with no effect
	// measures noise. Evidence says whether the row was refined, and marks a
	// ranked delta inside the phase's noise floor.
	writePropTable := func(title, phase string, sorted []propDumpRow, metric func(propDumpRow) string, delta func(propDumpRow) int64) {
		fmt.Fprintf(&b, "## %s\n\n", title)
		b.WriteString("| # | Property | Group | Metric | Evidence | Cold Delta | Hover Delta | Scale64 Delta | Raster Full Delta | Layer |\n")
		b.WriteString("|---:|---|---|---:|---|---:|---:|---:|---:|---|\n")
		n := min(propSummaryTopN, len(sorted))
		for i := range n {
			r := sorted[i]
			evidence := "1 sample"
			if r.HighPrecision {
				evidence = fmt.Sprintf("refined x%d", opts.Iters)
			}
			if absInt64(delta(r)) <= floors[phase] {
				evidence += ", ≤noise"
			}
			fmt.Fprintf(&b, "| %d | `%s` | %s | **%s** | %s | %.2f ms | %.2f ms | %.2f ms | %.2f ms | `%s` |\n",
				i+1, r.Name, r.Group, metric(r), evidence,
				float64(r.ColdDeltaNS)/propNSPerMS,
				float64(r.HoverDeltaNS)/propNSPerMS,
				float64(r.Scale64DeltaNS)/propNSPerMS,
				float64(r.RasterFullDeltaNS)/propNSPerMS,
				r.Layer,
			)
		}
		b.WriteString("\n")
	}

	valid := make([]propDumpRow, 0, len(rows))
	effective := make([]propDumpRow, 0, len(rows))
	for _, r := range rows {
		if r.Err == "" {
			valid = append(valid, r)
			if r.Effective {
				effective = append(effective, r)
			}
		}
	}
	ranked := func(delta func(propDumpRow) int64) []propDumpRow {
		sorted := slices.Clone(effective)
		slices.SortFunc(sorted, func(a, b propDumpRow) int { return cmpDescInt64(delta(a), delta(b)) })

		return sorted
	}

	scaleDelta := func(r propDumpRow) int64 { return r.Scale64DeltaNS }
	writePropTable("Top 25 by Scale-64 Paired Delta (Uniform 64-Node Grid)", "scale64", ranked(scaleDelta), func(r propDumpRow) string {
		ratio := "n/a"
		if r.ScalingRatio > 0 {
			ratio = fmt.Sprintf("%.1fx", r.ScalingRatio)
		}
		return fmt.Sprintf("%.2f ms (%.1f KB, ratio %s)", float64(r.Scale64DeltaNS)/propNSPerMS, float64(r.Scale64Bytes)/propBytesPerKB, ratio)
	}, scaleDelta)

	hoverDelta := func(r propDumpRow) int64 { return r.HoverDeltaNS }
	writePropTable("Top 25 by Hover Paired Delta (64-Node Retained Container)", "hover", ranked(hoverDelta), func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%d allocs, fb=%d)", float64(r.HoverDeltaNS)/propNSPerMS, r.HoverAllocs, r.HoverFallbacks)
	}, hoverDelta)

	// Raster ranks Redraw plus PNG on fresh pages, which is comparable on
	// every row; the PNG-only delta is shown only where control and variant
	// paint on the same path.
	rasterDelta := func(r propDumpRow) int64 { return r.RasterFullDeltaNS }
	writePropTable("Top 25 by Software Raster Paired Delta (Redraw + PNG)", "raster", ranked(rasterDelta), func(r propDumpRow) string {
		pngOnly := "n/a, fallback differs"
		if r.RasterDeltaValid {
			pngOnly = fmt.Sprintf("%.2f ms", float64(r.RasterDeltaNS)/propNSPerMS)
		}
		return fmt.Sprintf("%.2f ms (PNG only %s, %.1f KB)", float64(r.RasterFullDeltaNS)/propNSPerMS, pngOnly, float64(r.RasterBytes)/propBytesPerKB)
	}, rasterDelta)

	resizeDelta := func(r propDumpRow) int64 { return r.ResizeDeltaNS }
	writePropTable("Top 25 by Resize Relayout Paired Delta", "resize", ranked(resizeDelta), func(r propDumpRow) string {
		return fmt.Sprintf("%.2f ms (%d allocs)", float64(r.ResizeDeltaNS)/propNSPerMS, r.ResizeAllocs)
	}, resizeDelta)

	b.WriteString("## Failed Rows and Phase Errors\n\n")
	engineErrors := 0
	for _, r := range rows {
		if r.EngineError {
			engineErrors++
		}
	}
	if errCount == 0 && phaseErrors == 0 && engineErrors == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("| Property | Group | Phase | Error |\n|---|---|---|---|\n")
		for _, r := range rows {
			if r.Err != "" {
				fmt.Fprintf(&b, "| `%s` | %s | **row** | `%s` |\n", r.Name, r.Group, r.Err)
			}
			if r.EngineError {
				fmt.Fprintf(&b, "| `%s` | %s | **engine** | the page refused the fixture; phases that build their own pages still ran |\n", r.Name, r.Group)
			}
			phases := make([]string, 0, len(r.PhaseErrors))
			for phase := range r.PhaseErrors {
				phases = append(phases, phase)
			}
			slices.Sort(phases)
			for _, phase := range phases {
				fmt.Fprintf(&b, "| `%s` | %s | %s | `%s` |\n", r.Name, r.Group, phase, r.PhaseErrors[phase])
			}
		}
		b.WriteString("\n")
	}

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

	b.WriteString("## Unchanged Redraw and Tick Allocations, Bitmap Fallbacks\n\n")
	if warmLeaks == 0 && tickAllocRows == 0 && bitmapFallbacks == 0 {
		b.WriteString("None.\n\n")
	} else {
		b.WriteString("| Property | Group | Unchanged Redraw Allocs | Tick Allocs | Tick Time | Bitmap Fallback | Sample |\n|---|---|---:|---:|---:|---|---|\n")
		for _, r := range valid {
			if r.WarmAllocs > 0 || r.TickAllocs > 0 || r.BitmapFallback {
				fmt.Fprintf(&b, "| `%s` | %s | **%d** | **%d** | %.3f ms | %t | `%s` |\n",
					r.Name, r.Group, r.WarmAllocs, r.TickAllocs, float64(r.TickNS)/propNSPerMS, r.BitmapFallback, r.Sample)
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
				b.WriteString("- **Alloc top frames, workload only** (`profiles/" + a.MemFile + "` minus `profiles/" + a.MemBaseFile + "`):\n")
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

	comboPlanRows, comboPlan, err := loadPropReference(opts.Reference, filtered, newPropManifest(opts))
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
	// Every anchor sample runs the same way: measureTimedNS with gcIsolate
	// warms the page once and collects garbage before the timed loop.
	anchorStart, err := measureTimedNS(opts.Iters, true, anchorRun)
	if err != nil {
		return 0, 0, fmt.Errorf("anchor warmup: %w", err)
	}
	var anchorPcts []float64

	// Pass 1: Fast paired-control sweep over every property.
	selected := make([]propDumpRow, len(filtered))
	controls := make([]sweepControl, len(filtered))
	slowScale, slowHover, slowRaster := propDumpRow{}, propDumpRow{}, propDumpRow{}
	for i, r := range filtered {
		if i > 0 && i%propBenchDriftEvery == 0 {
			if cur, dErr := measureTimedNS(opts.Iters, true, anchorRun); dErr == nil && anchorStart.MedianNS > 0 {
				anchorPcts = append(anchorPcts,
					float64(absInt64(cur.MedianNS-anchorStart.MedianNS))*propPctScale/float64(anchorStart.MedianNS))
			}
		}
		row := measurePropertyFull(ctx, r, 1, false)
		selected[i] = row
		controls[i] = sweepControl{cold: row.ColdControlNS, scale: row.Scale64ControlNS, ok: row.Err == ""}
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

	drift := propDrift{AnchorMedianPct: medianFloat(anchorPcts)}
	for _, pct := range anchorPcts {
		drift.AnchorMaxPct = max(drift.AnchorMaxPct, pct)
	}
	drift.ColdControlPct, drift.Windows = controlDrift(controls, func(c sweepControl) int64 { return c.cold })
	drift.ScaleControlPct, _ = controlDrift(controls, func(c sweepControl) int64 { return c.scale })

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
		// Redraw plus PNG is comparable on every row; the PNG-only delta is
		// not when only one side is a bitmap fallback.
		pickTop("raster", func(r propDumpRow) int64 { return r.RasterFullDeltaNS }, max(1, opts.PprofTop-propBenchHalfDiv*perPhase))
	}

	if err := writeBenchscoreTxt(filepath.Join(opts.DumpDir, "props-bench.txt"), selected, combos, opts); err != nil {
		return 0, 0, err
	}
	if err := writeSummaryMarkdown(filepath.Join(opts.DumpDir, "summary.md"), selected, combos, attrs, opts, drift, time.Since(tStart)); err != nil {
		return 0, 0, err
	}
	if err := writePropManifest(opts.DumpDir, newPropManifest(opts)); err != nil {
		return 0, 0, err
	}

	return len(selected), len(combos), nil
}

// propManifest records the measurement version and the sampling flags of a
// dump. --reference requires an identical manifest, so a before and after
// pair always shares fixtures, controls, and iteration rules.
type propManifest struct {
	MeasurementVersion int    `json:"measurement_version"`
	Iters              int    `json:"iters"`
	RefineTop          int    `json:"refine_top"`
	TopPairs           int    `json:"top_pairs"`
	FuzzCount          int    `json:"fuzz"`
	Seed               uint64 `json:"seed"`
	Filter             string `json:"filter,omitempty"`
	Limit              int    `json:"limit,omitempty"`
}

func newPropManifest(opts propBenchOpts) propManifest {
	m := propManifest{
		MeasurementVersion: propMeasurementVersion,
		Iters:              opts.Iters,
		RefineTop:          opts.RefineTop,
		TopPairs:           opts.TopPairs,
		FuzzCount:          opts.FuzzCount,
		Seed:               opts.Seed,
		Limit:              opts.Limit,
	}
	if opts.Filter != nil {
		m.Filter = opts.Filter.String()
	}

	return m
}

func writePropManifest(dir string, m propManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(filepath.Join(dir, propManifestName)), append(data, '\n'), propFilePerm)
}

// checkPropReferenceManifest rejects a reference dump from another
// measurement version or with other sampling flags. A dump without a
// manifest predates version 3.
func checkPropReferenceManifest(dir string, want propManifest) error {
	data, err := os.ReadFile(filepath.Clean(filepath.Join(dir, propManifestName)))
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reference %s has no %s, so it predates measurement version %d; regenerate it with this harness",
			dir, propManifestName, propMeasurementVersion)
	}
	if err != nil {
		return fmt.Errorf("read reference manifest: %w", err)
	}
	var got propManifest
	if err := json.Unmarshal(data, &got); err != nil {
		return fmt.Errorf("decode reference manifest: %w", err)
	}
	if got.MeasurementVersion != want.MeasurementVersion {
		return fmt.Errorf("reference %s uses measurement version %d, this harness uses %d; regenerate it with this harness",
			dir, got.MeasurementVersion, want.MeasurementVersion)
	}
	if got != want {
		return fmt.Errorf("reference %s sampling flags %+v differ from this run's %+v", dir, got, want)
	}

	return nil
}

func loadPropReference(dir string, props []genProp, want propManifest) ([]propDumpRow, []comboEntry, error) {
	if dir == "" {
		return nil, nil, nil
	}
	if err := checkPropReferenceManifest(dir, want); err != nil {
		return nil, nil, err
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

func TestPropPprofTopArgsSubtractBase(t *testing.T) {
	args := pprofTopArgs("p/after.pprof", "-alloc_space", "p/before.pprof")
	base := slices.Index(args, "-base")
	if base < 0 || base+1 >= len(args) || args[base+1] != filepath.Clean("p/before.pprof") {
		t.Fatalf("args %v do not subtract the before snapshot", args)
	}
	if args[len(args)-1] != filepath.Clean("p/after.pprof") || !slices.Contains(args, "-alloc_space") {
		t.Fatalf("args %v lost the profile or sample flag", args)
	}
	if slices.Contains(pprofTopArgs("cpu.pprof", "", ""), "-base") {
		t.Fatal("a CPU profile must not subtract a base")
	}
}

func TestPropControlDriftComparesWindowMedians(t *testing.T) {
	controls := make([]sweepControl, 0, propDriftWindow*3)
	for window, ns := range []int64{100, 110, 90} {
		for i := range propDriftWindow {
			v := ns
			if i == 0 && window == 1 {
				v = 10_000 // one outlier must not move the window median
			}
			controls = append(controls, sweepControl{cold: v, ok: true})
		}
	}
	pct, windows := controlDrift(controls, func(c sweepControl) int64 { return c.cold })
	if windows != 3 || pct != 10 {
		t.Fatalf("control drift = %.1f%% over %d windows, want 10%% over 3", pct, windows)
	}
}

func TestPropNoiseFloorUsesRowsWithoutEffect(t *testing.T) {
	rows := []propDumpRow{{Effective: true, ColdDeltaNS: 9_000_000}, {Err: "x", ColdDeltaNS: 8_000_000}}
	for i := range 20 {
		rows = append(rows, propDumpRow{ColdDeltaNS: int64(i - 10)})
	}
	if got := noiseFloor(rows, func(r propDumpRow) int64 { return r.ColdDeltaNS }); got != 9 {
		t.Fatalf("noise floor = %d, want 9 (p95 of |delta| over rows with no effect)", got)
	}
}

func TestPropReferenceRejectsOtherVersion(t *testing.T) {
	want := newPropManifest(propBenchOpts{Iters: propBenchDefaultIters, Seed: propBenchDefaultSeed})
	dir := t.TempDir()
	if err := checkPropReferenceManifest(dir, want); err == nil || !strings.Contains(err.Error(), "predates") {
		t.Fatalf("missing manifest: %v", err)
	}
	old := want
	old.MeasurementVersion = 2
	if err := writePropManifest(dir, old); err != nil {
		t.Fatal(err)
	}
	if err := checkPropReferenceManifest(dir, want); err == nil || !strings.Contains(err.Error(), "version 2") {
		t.Fatalf("version 2 manifest: %v", err)
	}
	if err := writePropManifest(dir, want); err != nil {
		t.Fatal(err)
	}
	if err := checkPropReferenceManifest(dir, want); err != nil {
		t.Fatal(err)
	}
	other := want
	other.Iters++
	if err := checkPropReferenceManifest(dir, other); err == nil {
		t.Fatal("accepted a reference with different sampling flags")
	}
}

func TestPropBenchscoreEmitsRefinedRowsOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.txt")
	rows := []propDumpRow{
		{Name: "refined", HighPrecision: true, ColdNS: 100, Scale64NS: 200},
		{Name: "sweep", ColdNS: 100},
	}
	combos := []comboEntry{{Kind: "k", Name: "kept", Refined: true, Iters: 3, Scale64NS: 5}, {Kind: "k", Name: "zero", Refined: true, Iters: 3}}
	if err := writeBenchscoreTxt(path, rows, combos, propBenchOpts{Iters: 6}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, want := range []string{"BenchmarkPropCold_refined-1\t6\t", "BenchmarkPropScale64_refined-1\t3\t", "BenchmarkComboScale64_k_kept-1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("bench output lacks %q:\n%s", want, out)
		}
	}
	for _, reject := range []string{"_sweep-", "_zero-"} {
		if strings.Contains(out, reject) {
			t.Fatalf("bench output kept %q:\n%s", reject, out)
		}
	}
}
