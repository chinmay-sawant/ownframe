package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

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
	h("Pipeline detail")
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
