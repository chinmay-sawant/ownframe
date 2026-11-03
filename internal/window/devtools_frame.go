package window

import (
	"fmt"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
