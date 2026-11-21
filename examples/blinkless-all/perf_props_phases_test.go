package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"runtime"

	"github.com/chinmay-sawant/ownframe"
)

type coldWorkload struct {
	doc string
	w   int
	h   int
	p   *ownframe.Page
}

func (work *coldWorkload) run(ctx context.Context) error {
	p, err := newBenchPage(work.doc, work.w, work.h, false)
	if err != nil {
		return err
	}
	if err := p.Redraw(ctx); err != nil {
		return err
	}
	work.p = p

	return nil
}

func measureColdSingle(ctx context.Context, doc string, w, h, iters int, gcIsolate, withAllocs bool) (
	timingStats, allocStats, ownframe.Stats, *ownframe.Page, uint64, error,
) {
	work := coldWorkload{doc: doc, w: w, h: h}
	run := func() error { return work.run(ctx) }
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

	return tm, al, work.p.Probe().Stats(), work.p, displaySignature(work.p), nil
}

type coldMeasurement struct {
	timing timingStats
	allocs allocStats
	page   *ownframe.Page
}

func measureColdPair(
	ctx context.Context, offDoc, onDoc string, w, h, iters int, gcIsolate bool,
) (coldMeasurement, coldMeasurement, error) {
	off := coldWorkload{doc: offDoc, w: w, h: h}
	on := coldWorkload{doc: onDoc, w: w, h: h}
	runOff := func() error {
		if err := off.run(ctx); err != nil {
			return fmt.Errorf("control: %w", err)
		}

		return nil
	}
	runOn := func() error {
		if err := on.run(ctx); err != nil {
			return fmt.Errorf("variant: %w", err)
		}

		return nil
	}
	timing, err := measureWorkloadsNS(iters, gcIsolate, runOff, runOn)
	if err != nil {
		return coldMeasurement{}, coldMeasurement{}, err
	}

	allocIters := max(1, iters/propBenchHalfDiv)
	offAllocs, err := measureAllocAvg(allocIters, runOff)
	if err != nil {
		return coldMeasurement{}, coldMeasurement{}, err
	}
	onAllocs, err := measureAllocAvg(allocIters, runOn)
	if err != nil {
		return coldMeasurement{}, coldMeasurement{}, err
	}

	return coldMeasurement{timing: timing[0], allocs: offAllocs, page: off.p},
		coldMeasurement{timing: timing[1], allocs: onAllocs, page: on.p}, nil
}

type hoverWorkload struct {
	p      *ownframe.Page
	hx, hy float64
}

func newHoverWorkload(ctx context.Context, doc string) (hoverWorkload, error) {
	p, err := newBenchPage(doc, propBenchScaleWidth, propBenchScaleHeight, true)
	if err != nil {
		return hoverWorkload{}, err
	}
	if err := p.Redraw(ctx); err != nil {
		return hoverWorkload{}, err
	}
	hx, hy, ok := findHoverTarget(p, "h0")
	if !ok {
		return hoverWorkload{}, fmt.Errorf("missing hover target h0")
	}

	return hoverWorkload{p: p, hx: hx, hy: hy}, nil
}

func (work hoverWorkload) enter(ctx context.Context) error {
	if err := work.p.Edit().Hover(ctx, work.hx, work.hy); err != nil {
		return err
	}

	return work.p.Testing().Tick(ctx)
}

func (work hoverWorkload) cycle(ctx context.Context) error {
	if err := work.enter(ctx); err != nil {
		return err
	}
	if err := work.p.Edit().Hover(ctx, propBenchHoverOffX, propBenchHoverOffY); err != nil {
		return err
	}

	return work.p.Testing().Tick(ctx)
}

func hoverFailureReasons(stats ownframe.Stats) []string {
	reasons := make([]string, 0, len(stats.HoverPatchFailures))
	for _, f := range stats.HoverPatchFailures {
		reasons = append(reasons, fmt.Sprintf("%s:%s(%d)", f.Stage, f.Reason, f.Count))
	}

	return reasons
}

func measureSingleHover(ctx context.Context, doc string, iters int, gcIsolate bool) (
	timingStats, allocStats, ownframe.Stats, int, int, []string, error,
) {
	work, err := newHoverWorkload(ctx, doc)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	cycle := func() error { return work.cycle(ctx) }
	tm, err := measureTimedNS(iters, gcIsolate, cycle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	al, err := measureAllocAvg(1, cycle)
	if err != nil {
		return timingStats{}, allocStats{}, ownframe.Stats{}, 0, 0, nil, err
	}
	st := work.p.Probe().Stats()

	return tm, al, st, st.DirtyRegions, st.DirtyOps, hoverFailureReasons(st), nil
}

func measurePropertyHover(ctx context.Context, row *propDumpRow, head string, iters int, gcIsolate bool) error {
	off, err := newHoverWorkload(ctx, buildHoverDoc(head, row.Name, row.Sample, false))
	if err != nil {
		return fmt.Errorf("control: %w", err)
	}
	on, err := newHoverWorkload(ctx, buildHoverDoc(head, row.Name, row.Sample, true))
	if err != nil {
		return fmt.Errorf("variant: %w", err)
	}
	runOff := func() error { return off.cycle(ctx) }
	runOn := func() error { return on.cycle(ctx) }
	timing, err := measureWorkloadsNS(iters, gcIsolate, runOff, runOn)
	if err != nil {
		return err
	}
	allocs, err := measureAllocAvg(1, runOn)
	if err != nil {
		return err
	}

	stats := on.p.Probe().Stats()
	row.HoverNS = timing[1].MedianNS
	row.HoverControlNS = timing[0].MedianNS
	row.HoverIQRNS = timing[1].IQRNS
	row.HoverDeltaNS = row.HoverNS - row.HoverControlNS
	row.HoverBytes = allocs.Bytes
	row.HoverAllocs = allocs.Allocs
	row.HoverPatches = stats.HoverPatches
	row.HoverFallbacks = stats.HoverFallbacks
	row.HoverPatchErrors = stats.HoverPatchErrors
	row.HoverFailureReasons = hoverFailureReasons(stats)
	row.HoverDirtyRegions = stats.DirtyRegions
	row.HoverDirtyOps = stats.DirtyOps

	// Inspect the hovered state after timing. Dirty counts also rise for the
	// empty control and therefore cannot establish a property effect.
	if err := off.enter(ctx); err != nil {
		return fmt.Errorf("inspect control: %w", err)
	}
	if err := on.enter(ctx); err != nil {
		return fmt.Errorf("inspect variant: %w", err)
	}
	if displaySignature(off.p) != displaySignature(on.p) {
		row.EffectEvidence = append(row.EffectEvidence, "hover-display")
	}

	return nil
}

func newResizeWorkload(ctx context.Context, doc string) (*ownframe.Page, func() error, error) {
	p, err := newBenchPage(doc, propBenchWidth, propBenchHeight, true)
	if err != nil {
		return nil, nil, err
	}
	if err := p.Redraw(ctx); err != nil {
		return nil, nil, err
	}
	cycle := func() error {
		p.SetSize(propBenchResizeWidth, propBenchResizeHeight)
		if err := p.Redraw(ctx); err != nil {
			return err
		}
		p.SetSize(propBenchWidth, propBenchHeight)

		return p.Redraw(ctx)
	}

	return p, cycle, nil
}

func measurePropertyResize(ctx context.Context, row *propDumpRow, offDoc, onDoc string, iters int, gcIsolate bool) error {
	_, off, err := newResizeWorkload(ctx, offDoc)
	if err != nil {
		return fmt.Errorf("control: %w", err)
	}
	p, on, err := newResizeWorkload(ctx, onDoc)
	if err != nil {
		return fmt.Errorf("variant: %w", err)
	}
	timing, err := measureWorkloadsNS(iters, gcIsolate, off, on)
	if err != nil {
		return err
	}
	allocs, err := measureAllocAvg(1, on)
	if err != nil {
		return err
	}
	stats := p.Probe().Stats()
	row.ResizeNS = timing[1].MedianNS
	row.ResizeControlNS = timing[0].MedianNS
	row.ResizeIQRNS = timing[1].IQRNS
	row.ResizeDeltaNS = row.ResizeNS - row.ResizeControlNS
	row.ResizeBytes = allocs.Bytes
	row.ResizeAllocs = allocs.Allocs
	row.ResizeStyleNS = stats.StyleTime.Nanoseconds()
	row.ResizeLayoutNS = stats.LayoutTime.Nanoseconds()

	return nil
}

type rasterWorkload struct {
	pages []*ownframe.Page
	pngs  [][]byte
	index int
}

func newRasterWorkload(ctx context.Context, doc string, w, h, iters int) (rasterWorkload, error) {
	work := rasterWorkload{
		pages: make([]*ownframe.Page, iters),
		pngs:  make([][]byte, iters),
	}
	for i := range work.pages {
		p, err := newBenchPage(doc, w, h, false)
		if err != nil {
			return rasterWorkload{}, err
		}
		if err := p.Redraw(ctx); err != nil {
			return rasterWorkload{}, err
		}
		work.pages[i] = p
	}

	return work, nil
}

func (work *rasterWorkload) run(ctx context.Context) error {
	work.pngs[work.index] = work.pages[work.index].Probe().PNGWithContext(ctx)
	work.index++

	return nil
}

func validateRasterPNG(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty PNG")
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("decode PNG: %w", err)
	}

	return nil
}

type rasterMeasurement struct {
	timing timingStats
	allocs allocStats
	png    []byte
}

func measureRasterWorkloads(
	ctx context.Context, w, h, iters int, gcIsolate bool, docs ...string,
) ([]rasterMeasurement, error) {
	// Keep one fresh page per timed sample and one for allocations. Redraw
	// happens before timing and each page is encoded exactly once.
	iters = max(1, iters)
	work := make([]rasterWorkload, len(docs))
	timed := make([]func() error, len(docs))
	for i, doc := range docs {
		var err error
		work[i], err = newRasterWorkload(ctx, doc, w, h, iters+1)
		if err != nil {
			return nil, fmt.Errorf("prepare workload %d: %w", i+1, err)
		}
		timed[i] = func() error { return work[i].run(ctx) }
	}
	if gcIsolate {
		runtime.GC()
	}
	timing, err := measureWorkloadsNS(iters, false, timed...)
	if err != nil {
		return nil, err
	}

	results := make([]rasterMeasurement, len(work))
	for i := range work {
		allocs, err := measureAllocAvg(1, timed[i])
		if err != nil {
			return nil, err
		}
		for j, data := range work[i].pngs {
			if err := validateRasterPNG(data); err != nil {
				return nil, fmt.Errorf("workload %d sample %d: %w", i+1, j+1, err)
			}
		}
		results[i] = rasterMeasurement{timing: timing[i], allocs: allocs, png: work[i].pngs[0]}
	}

	return results, nil
}

func measureSingleRaster(ctx context.Context, doc string, w, h, iters int, gcIsolate bool) (
	timingStats, allocStats, error,
) {
	results, err := measureRasterWorkloads(ctx, w, h, iters, gcIsolate, doc)
	if err != nil {
		return timingStats{}, allocStats{}, err
	}

	return results[0].timing, results[0].allocs, nil
}

func measurePropertyScale(ctx context.Context, row *propDumpRow, head string, iters int, gcIsolate bool) error {
	offDoc := buildUniformScaleDoc(head, "", propBenchScaleNodes)
	onDoc := buildUniformScaleDoc(head, row.Name+":"+row.Sample+";", propBenchScaleNodes)
	off, on, err := measureColdPair(ctx, offDoc, onDoc, propBenchScaleWidth, propBenchScaleHeight, iters, gcIsolate)
	if err != nil {
		return err
	}
	row.Scale64NS = on.timing.MedianNS
	row.Scale64ControlNS = off.timing.MedianNS
	row.Scale64IQRNS = on.timing.IQRNS
	row.Scale64DeltaNS = row.Scale64NS - row.Scale64ControlNS
	row.Scale64Bytes = on.allocs.Bytes
	row.Scale64Allocs = on.allocs.Allocs
	if d := on.page.Probe().Display(); d != nil {
		row.Scale64Ops = len(d.Ops)
	}
	if row.Scale64DeltaNS > 0 {
		row.ScalingRatio = float64(row.Scale64DeltaNS) / float64(max(propBenchMinBaseNS, row.ColdDeltaNS))
	}
	if displaySignature(off.page) != displaySignature(on.page) {
		row.EffectEvidence = append(row.EffectEvidence, "scale64-display")
	}

	return nil
}

func measurePropertyRaster(ctx context.Context, row *propDumpRow, offDoc, onDoc string, iters int, gcIsolate bool) error {
	results, err := measureRasterWorkloads(ctx, propBenchWidth, propBenchHeight, iters, gcIsolate, offDoc, onDoc)
	if err != nil {
		return err
	}
	off, on := results[0], results[1]
	row.RasterNS = on.timing.MedianNS
	row.RasterControlNS = off.timing.MedianNS
	row.RasterIQRNS = on.timing.IQRNS
	row.RasterDeltaNS = row.RasterNS - row.RasterControlNS
	row.RasterBytes = on.allocs.Bytes
	row.RasterAllocs = on.allocs.Allocs
	// The same PNG encoder emits both controls in one run.
	if !bytes.Equal(off.png, on.png) {
		row.EffectEvidence = append(row.EffectEvidence, "raster-png")
	}

	return nil
}

func (row *propDumpRow) notePhaseError(phase string, err error) {
	if err == nil {
		return
	}
	if row.PhaseErrors == nil {
		row.PhaseErrors = make(map[string]string)
	}
	row.PhaseErrors[phase] = err.Error()
}

// measurePropertyFull keeps the sweep at one sample per workload and uses
// repeated, GC-isolated samples for the rows selected for refinement.
func measurePropertyFull(ctx context.Context, r genProp, iters int, highPrecision bool) propDumpRow {
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

	off, on, err := measureColdPair(ctx, offDoc, onDoc, propBenchWidth, propBenchHeight, phaseIters, highPrecision)
	if err != nil {
		row.Err = fmt.Sprintf("cold: %v", err)

		return row
	}
	onPage := on.page
	onStats := onPage.Probe().Stats()
	if displaySignature(off.page) != displaySignature(onPage) {
		row.EffectEvidence = append(row.EffectEvidence, "cold-display")
	}
	row.ColdNS = on.timing.MedianNS
	row.ColdControlNS = off.timing.MedianNS
	row.ColdP95NS = on.timing.P95NS
	row.ColdIQRNS = on.timing.IQRNS
	row.ColdDeltaNS = row.ColdNS - row.ColdControlNS
	row.ColdBytes = on.allocs.Bytes
	row.ColdDeltaBytes = safeSubUint64(on.allocs.Bytes, off.allocs.Bytes)
	row.ColdAllocs = on.allocs.Allocs
	row.ColdDeltaAllocs = safeSubUint64(on.allocs.Allocs, off.allocs.Allocs)
	row.StyleNS = onStats.StyleTime.Nanoseconds()
	row.LayoutNS = onStats.LayoutTime.Nanoseconds()
	row.PaintNS = onStats.PaintTime.Nanoseconds()
	row.OtherNS = onStats.OtherTime.Nanoseconds()
	if d := onPage.Probe().Display(); d != nil {
		row.Ops = len(d.Ops)
		row.ExtraOps = countExtraOps(d)
	}
	row.BitmapFallback = onPage.Probe().Image() != nil && onPage.Probe().Display() == nil

	tick := func() error { return onPage.Testing().Tick(ctx) }
	row.notePhaseError("tick-settle", tick())
	motionStats := onPage.Probe().Motion()
	offMotion := off.page.Probe().Motion()
	row.MotionBound = motionStats.Bound + motionStats.TransitionsBound
	row.MotionActive = motionStats.Active
	if row.MotionBound != offMotion.Bound+offMotion.TransitionsBound || row.MotionActive != offMotion.Active {
		row.EffectEvidence = append(row.EffectEvidence, "motion")
	}

	redraw := func() error { return onPage.Redraw(ctx) }
	if warmT, err := measureTimedNS(phaseIters, false, redraw); err == nil {
		row.WarmNS = warmT.MedianNS
	} else {
		row.notePhaseError("warm", err)
	}
	if warmA, err := measureAllocAvg(1, redraw); err == nil {
		row.WarmBytes = warmA.Bytes
		row.WarmAllocs = warmA.Allocs
	} else {
		row.notePhaseError("warm-allocs", err)
	}
	if tickT, err := measureTimedNS(phaseIters, false, tick); err == nil {
		row.TickNS = tickT.MedianNS
	} else {
		row.notePhaseError("tick", err)
	}
	if tickA, err := measureAllocAvg(1, tick); err == nil {
		row.TickAllocs = tickA.Allocs
	} else {
		row.notePhaseError("tick-allocs", err)
	}

	row.notePhaseError("hover", measurePropertyHover(ctx, &row, head, phaseIters, highPrecision))
	row.notePhaseError("resize", measurePropertyResize(ctx, &row, offDoc, onDoc, phaseIters, highPrecision))
	reducedIters := max(1, phaseIters/propBenchHalfDiv)
	row.notePhaseError("scale64", measurePropertyScale(ctx, &row, head, reducedIters, highPrecision))
	row.notePhaseError("raster", measurePropertyRaster(ctx, &row, offDoc, onDoc, reducedIters, highPrecision))

	row.Effective = len(row.EffectEvidence) > 0
	row.Layer = classifyBottleneckLayer(row)

	return row
}
