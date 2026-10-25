//go:build !js

package main

import (
	"context"
	"log"
	"math"
	"runtime"
	"sort"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

// profileWarmup skips the first frames after the window starts drawing, so
// the page's first layout and the dock's first build do not land in the
// samples.
const (
	profileWarmup = 2 * time.Second
	profilePoll   = 50 * time.Millisecond
	profileDrain  = 200 * time.Millisecond
	profileP95    = 0.95
)

// profileSample is one per-second reading of the window counters.
type profileSample struct {
	fps float64
	tps float64
}

// profileResult is the samples and measured span delivered after the run.
type profileResult struct {
	samples    []profileSample
	elapsed    time.Duration
	winW, winH int
}

// startProfile waits for the first drawn frame, samples the frame counters
// once a second for d, then cancels the run and sends the samples. It runs off
// the page loop, so it does not register a tick callback: a page tick snapshots
// and diffs every operation, which would change the frame path being measured.
func startProfile(d time.Duration, cancel context.CancelFunc) <-chan profileResult {
	out := make(chan profileResult, 1)

	go func() {
		for ebiten.ActualFPS() <= 0 {
			time.Sleep(profilePoll)
		}
		time.Sleep(profileWarmup)

		start := time.Now()
		samples := make([]profileSample, 0, int(d/time.Second)+1)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for now := range ticker.C {
			samples = append(samples, profileSample{fps: ebiten.ActualFPS(), tps: ebiten.ActualTPS()})
			if now.Sub(start) >= d {
				break
			}
		}
		winW, winH := ebiten.WindowSize()
		cancel()
		out <- profileResult{samples: samples, elapsed: time.Since(start), winW: winW, winH: winH}
	}()

	return out
}

// logProfileSummary prints one greppable block with the frame rate, the page
// pipeline, the motion counters, and memory. The run has stopped by the time
// this is called, so reading the page counters does not race the page loop.
func logProfileSummary(p *ownframe.Page, samples []profileSample, elapsed time.Duration, winW, winH int) {
	fps := make([]float64, len(samples))
	tps := make([]float64, len(samples))
	for i := range samples {
		fps[i] = samples[i].fps
		tps[i] = samples[i].tps
	}

	series := make([]int, len(samples))
	for i := range samples {
		series[i] = int(math.Round(samples[i].fps))
	}

	fpsAvg, tpsAvg := mean(fps), mean(tps)
	sort.Float64s(fps)
	sort.Float64s(tps)
	fpsMin, tpsMin := first(fps), first(tps)
	fpsP95, tpsP95 := percentile(fps, profileP95), percentile(tps, profileP95)

	pageW, pageH := p.Size()
	log.Printf("profile seconds=%.0f samples=%d page=%dx%d window=%dx%d vsync=%v fps_avg=%.1f fps_min=%.1f fps_p95=%.1f tps_avg=%.1f tps_min=%.1f tps_p95=%.1f",
		elapsed.Seconds(), len(samples), pageW, pageH, winW, winH, ebiten.IsVsyncEnabled(),
		fpsAvg, fpsMin, fpsP95, tpsAvg, tpsMin, tpsP95)
	log.Printf("profile fps_series=%v", series)

	st := p.Probe().Stats()
	log.Printf("profile page redraws=%d parses=%d cascades=%d layouts=%d repaints=%d last_redraw=%v last_draw=%v template=%v style=%v layout=%v paint=%v other=%v dirty_ops=%d dirty_regions=%d dirty_area=%d changed_ops=%d alloc_frame=%dB",
		st.Redraws, st.Parses, st.Cascades, st.Layouts, st.Repaints, st.LastRedraw, st.LastDraw,
		st.LastTemplate, st.StyleTime, st.LayoutTime, st.PaintTime, st.OtherTime,
		st.DirtyOps, st.DirtyRegions, st.DirtyArea, st.ChangedOps, st.AllocFrame)

	mo := p.Probe().Motion()
	log.Printf("profile motion bound=%d active=%d sampled=%d restamped=%d rebinds=%d inactive_timelines=%d gaps=%d",
		mo.Bound, mo.Active, mo.Sampled, mo.Restamped, mo.Rebinds, mo.InactiveTimelines, len(mo.Gaps))

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Printf("profile memory heap=%dB sys=%dB goroutines=%d", mem.HeapAlloc, mem.Sys, runtime.NumGoroutine())
}

func mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}

	sum := 0.0
	for _, x := range v {
		sum += x
	}

	return sum / float64(len(v))
}

// first returns the smallest value after the caller sorted v.
func first(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}

	return v[0]
}

// percentile returns the nearest-rank p percentile of sorted v.
func percentile(v []float64, p float64) float64 {
	if len(v) == 0 {
		return 0
	}

	i := int(math.Ceil(p*float64(len(v)))) - 1
	if i < 0 {
		i = 0
	}
	if i >= len(v) {
		i = len(v) - 1
	}

	return v[i]
}
