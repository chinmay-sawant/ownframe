package window

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// perfSource is a screen that reports its own opt-in. The page implements
// it; a custom screen without it relies on DevTools or Options.Perf.
type perfSource interface{ Perf() bool }

// perfEnabled reports whether the screen opted into sampling.
func perfEnabled(app host.Screen) bool {
	src, ok := app.(perfSource)

	return ok && src.Perf()
}

// NewGameWithOptions returns the screen loop with perf sampling configured.
// It keeps the NewGame behavior for every other setting.
func NewGameWithOptions(ctx context.Context, app host.Screen, options Options) ebiten.Game {
	game := newShell(ctx, app)
	game.perfOption = options.Perf
	game.perf = options.Perf || perfEnabled(app)
	game.wirePerf()

	return game
}

// wirePerf installs this window's opt-in performance readers.
func (s *shell) wirePerf() {
	s.perfHooks.devFramePerf = func() (avg, p95, p99 time.Duration, long uint64, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, false
		}
		a, p95v, p99v, l := s.perfSummary()
		if l < 0 {
			l = 0
		}

		return a, p95v, p99v, uint64(l), true
	}
	s.perfHooks.devPipelinePerf = func(st host.Stats) (tpl, lay, dsp, pnt time.Duration, dOps, dReg, chOps int, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, 0, 0, 0, false
		}

		return st.LastTemplate, st.StyleTime, st.LayoutTime, st.PaintTime, st.DirtyOps, st.DirtyRegions, st.ChangedOps, true
	}
	s.perfHooks.devRuntimePerf = func() (alloc, heap, rss uint64, gr int, ok bool) {
		if !s.perf {
			return 0, 0, 0, 0, false
		}
		h, r, a, g := s.perfRuntime()

		return a, h, r, g, true
	}
}
