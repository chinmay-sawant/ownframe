package window

import (
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image/color"
	"strings"
	"time"
)

// devFrameRows returns the Frame tab: uppercase headers and right-aligned rows.
func (s *shell) devFrameRows(width float64) []devRow {
	st := s.dev.stats
	appW, appH := s.app.Size()
	rs := []devRow{}
	h := func(s string) { rs = append(rs, devRow{}, devRow{line: devText(strings.ToUpper(s), devDim)}) }
	n := func(l string, v any) { rs = append(rs, devFrameRow(width, l, fmt.Sprint(v), devNumInk)) }
	b := func(l string, v bool) { rs = append(rs, devFrameRow(width, l, fmt.Sprint(v), devBoolInk)) }
	h("Sampling")
	label := "Pause auto refresh"
	if s.dev.paused {
		label = "Resume auto refresh"
	}
	rs = append(rs, devRow{line: devPlain(label), act: devActPause})
	gap := s.dev.refreshGap
	if gap == 0 {
		gap = devPanelRefreshGap
	}
	rs = append(rs, devRow{line: devPlain(fmt.Sprintf("Refresh interval: %s", gap)), act: devActRefresh})
	if !s.dev.sampler.last.IsZero() {
		n("Sample age", time.Since(s.dev.sampler.last).Round(time.Millisecond))
	}

	h("Window")
	n("Window", fmt.Sprintf("%dx%d", s.screenW, s.screenH))
	n("Page", fmt.Sprintf("%dx%d", appW, appH))
	n("Scroll", fmt.Sprintf("%d,%d", s.scrollX, s.scrollY))
	b("Stretched", s.stretched())
	b("Fallback", s.fallback)
	n("Sequence", s.seq)
	h("Rendering")
	n("FPS", fmt.Sprintf("%.1f", ebiten.ActualFPS()))
	n("TPS", fmt.Sprintf("%.1f", ebiten.ActualTPS()))
	n("Frame gap", devMS(s.dev.frame))
	if s.perf {
		n("Frame work", devMS(s.dev.stages.work))
	} else {
		n("Frame work", "-")
	}
	n("Draw", devMS(s.dev.draw))
	s.addPerfRows(&rs, width, h, n)
	h("Pipeline")
	n("Redraws", st.Redraws)
	n("Parses", st.Parses)
	n("Cascades", st.Cascades)
	n("Layouts", st.Layouts)
	n("Repaints", st.Repaints)
	n("Relayouts", s.commits)
	n("Skipped", s.skipped)
	n("Boxes", st.Boxes)
	n("Ops", st.Ops)
	n("Last redraw", devMS(st.LastRedraw))
	n("Last draw", devMS(st.LastDraw))
	s.addPipelinePerfRows(&rs, width, h, n, st)
	h("Pacing")
	b("Idle", s.pace.idle)
	n("MaxFPS", s.pace.maxFPS)
	n("Present FPS", fmt.Sprintf("%.1f", s.pace.presentFPS))
	n("Presents", s.pace.presents)
	n("Repaints/sec", fmt.Sprintf("%.1f", s.dev.sampler.repaintsPerSec))
	n("Alloc/sec", devBytes(s.dev.sampler.allocPerSec))
	n("Next", devPaceNext(s.pace.next))
	n("Pace skip", s.pace.skipped)
	s.addMemoryRows(&rs, width, h, n)
	h("Reload")
	n("Reloads", st.Reloads)
	ink := devErrInk
	if st.LastReloadError == "" {
		ink = devDim
	}
	rs = append(rs, devFrameRow(width, "Reload error", devErr(st.LastReloadError), ink))

	return rs[1:]
}

// bytesPerKB and bytesPerMB size the MEMORY rows.
const (
	bytesPerKB = 1024
	bytesPerMB = 1024 * bytesPerKB
)

// devFrameRow builds one right-aligned row. The label and separator widths
// come from the panel memo; only the value, which moves as counters change,
// is measured per call. The value span right-aligns at draw time, so no
// padding run is built or drawn.
func devFrameRow(w float64, l, v string, ink color.RGBA) devRow {
	vw, _ := text.Measure(v, badgeFace, 0)

	return devRow{
		line: devLine{spans: []devSpan{
			{text: l, ink: devDim, w: devTextWidth(l)},
			{text: " ", ink: devPunctInk, w: devTextWidth(" ")},
			{text: v, ink: ink, w: vw, right: true},
		}},
		w: w,
	}
}

// addMemoryRows prints the MEMORY section. A shell without Perf reports
// !ok, so end users see dashes and developers see live runtime numbers.
func (s *shell) addMemoryRows(rs *[]devRow, w float64, h func(string), n func(string, any)) {
	h("Memory")
	if s.perfHooks.devRuntimePerf == nil {
		for _, l := range []string{"Alloc/frame", "Go heap", "RSS", "Goroutines"} {
			n(l, "-")
		}
		return
	}
	alloc, heap, rss, gr, ok := s.perfHooks.devRuntimePerf()
	if !ok {
		for _, l := range []string{"Alloc/frame", "Go heap", "RSS", "Goroutines"} {
			n(l, "-")
		}
		return
	}
	n("Alloc/frame", devBytes(alloc))
	n("Go heap", devBytes(heap))
	n("RSS", devRSS(rss))
	n("Goroutines", gr)
}

// devRSS prints a resident byte count for the panel. Zero means the
// platform reader reported nothing, since genuine RSS is never zero,
// so the row reads "-" instead of "0B".
func devRSS(b uint64) string {
	if b == 0 {
		return "-"
	}
	return devBytes(b)
}

// devBytes prints a byte count for the panel.
func devBytes(b uint64) string {
	if b < bytesPerKB {
		return fmt.Sprintf("%dB", b)
	}
	if b < bytesPerMB {
		return fmt.Sprintf("%.1fKB", float64(b)/bytesPerKB)
	}
	return fmt.Sprintf("%.1fMB", float64(b)/bytesPerMB)
}

// Each window owns its hooks. Parallel windows cannot replace each other.
type perfHooks struct {
	devFramePerf    func() (avg, p95, p99 time.Duration, long uint64, ok bool)
	devPipelinePerf func(st host.Stats) (tpl, lay, dsp, pnt time.Duration, dOps, dReg, chOps int, ok bool)
	devRuntimePerf  func() (alloc, heap, rss uint64, gr int, ok bool)
}

func (s *shell) addPerfRows(rs *[]devRow, w float64, h func(string), n func(string, any)) {
	h("Performance")
	if s.perfHooks.devFramePerf == nil {
		for _, l := range []string{"Frame avg", "Frame p95", "Frame p99", "Long frames"} {
			n(l, "-")
		}
		return
	}
	avg, p95, p99, long, ok := s.perfHooks.devFramePerf()
	if !ok {
		for _, l := range []string{"Frame avg", "Frame p95", "Frame p99", "Long frames"} {
			n(l, "-")
		}
		return
	}
	n("Frame avg", devMS(avg))
	n("Frame p95", devMS(p95))
	n("Frame p99", devMS(p99))
	n("Long frames", long)
}

func (s *shell) addPipelinePerfRows(rs *[]devRow, w float64, h func(string), n func(string, any), st host.Stats) {
	h("Last redraw detail")
	if s.perfHooks.devPipelinePerf == nil {
		for _, l := range []string{"Template", "Style", "Layout", "Paint", "Dirty ops", "Dirty regions", "Changed ops"} {
			n(l, "-")
		}
		return
	}
	tpl, lay, dsp, pnt, dOps, dReg, chOps, ok := s.perfHooks.devPipelinePerf(st)
	if !ok {
		for _, l := range []string{"Template", "Style", "Layout", "Paint", "Dirty ops", "Dirty regions", "Changed ops"} {
			n(l, "-")
		}
		return
	}
	n("Template", devMS(tpl))
	n("Style", devMS(lay))
	n("Layout", devMS(dsp))
	n("Paint", devMS(pnt))
	n("Other", devMS(st.OtherTime))
	n("Dirty area", st.DirtyArea)
	n("Dirty ops", dOps)
	n("Dirty regions", dReg)
	n("Changed ops", chOps)
}
