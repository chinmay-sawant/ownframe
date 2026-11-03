package window

import (
	"runtime"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// debugPublishGap bounds snapshot copies to about twice a second, the same
// cadence as the runtime sampler. A GUI window with no debug listener never
// reaches the clock call: maybePublishDebug returns on the enabled flag.
const debugPublishGap = 500 * time.Millisecond

// windowDebugState is the JSON body of GET /debug/state on the window's
// debug listener. It mirrors the web endpoint's names so one sampler reads
// both, and adds the native frame ring the web snapshot lacks.
type windowDebugState struct {
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Generation uint64         `json:"generation"`
	Fallback   bool           `json:"fallback"`
	Boxes      []layout.Box   `json:"boxes"`
	Stats      *host.Stats    `json:"stats,omitempty"`
	Ops        map[string]int `json:"ops,omitempty"`
	Runtime    map[string]any `json:"runtime,omitempty"`
	Frame      map[string]any `json:"frame,omitempty"`
}

// maybePublishDebug copies one profiling snapshot for the debug listener.
// It runs on the game thread at the end of Draw. With no listener it is one
// bool check and no clock call, so a shipped window pays nothing.
func (s *shell) maybePublishDebug() {
	if !s.debugEnabled {
		return
	}
	now := time.Now()
	if !s.debugLast.IsZero() && now.Sub(s.debugLast) < debugPublishGap {
		return
	}
	s.debugLast = now
	s.publishDebug()
}

// publishDebug builds a fresh immutable snapshot and stores it for the HTTP
// thread. Every map and slice is newly allocated: once stored, the game
// thread never mutates it, so the handler loads without a mutex.
func (s *shell) publishDebug() {
	snap := &windowDebugState{}
	front := host.SelectFront(s.app)
	snap.Width = front.Width
	snap.Height = front.Height
	snap.Generation = front.Generation
	snap.Fallback = front.Fallback
	snap.Boxes = s.app.Boxes()
	if front.Display != nil {
		snap.Ops = debugOpCountsMap(front.Display)
	}
	if insp, ok := s.app.(host.Inspector); ok {
		stats := insp.Stats()
		stats.LastReloadError = sanitizeDebugLine(stats.LastReloadError)
		snap.Stats = &stats
	}
	if s.perf {
		snap.Runtime = debugRuntimeSnapshot()
		avg, p95, p99, long := s.perfSummary()
		snap.Frame = map[string]any{
			"avg_ns":      int64(avg),
			"p95_ns":      int64(p95),
			"p99_ns":      int64(p99),
			"long_frames": long,
		}
	}
	s.debugSnapshot.Store(snap)
}

// debugOpCountsMap counts the operations that paint, keyed by the web
// endpoint's kind names so one agent reads both servers.
func debugOpCountsMap(display *layout.Display) map[string]int {
	counts := map[string]int{}
	if display == nil {
		return counts
	}
	for i := range display.Ops {
		name := debugOpKindName(display.Ops[i].Kind)
		if name == "" {
			continue
		}
		counts[name]++
		counts["total"]++
	}
	return counts
}

// debugOpKindName names one operation kind with the web endpoint's words.
// Noop, unknown, and link operations paint nothing and stay uncounted.
func debugOpKindName(kind layout.DisplayKind) string {
	switch kind {
	case layout.DisplayOpFillRect:
		return "fill"
	case layout.DisplayOpStrokeRect:
		return "stroke"
	case layout.DisplayOpLine:
		return "line"
	case layout.DisplayOpText:
		return "text"
	case layout.DisplayOpBullet:
		return "bullet"
	case layout.DisplayOpImage:
		return "image"
	case layout.DisplayOpGridRun:
		return "grid"
	case layout.DisplayOpPath:
		return "path"
	}
	return ""
}

// debugRuntimeSnapshot samples this process like the web endpoint: heap,
// system bytes, total allocations, collection count, goroutines, and
// resident bytes when the platform reports them.
func debugRuntimeSnapshot() map[string]any {
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	snapshot := map[string]any{
		"heap_alloc":  memory.HeapAlloc,
		"heap_sys":    memory.HeapSys,
		"total_alloc": memory.TotalAlloc,
		"goroutines":  runtime.NumGoroutine(),
		"num_gc":      memory.NumGC,
	}
	if rss := perfRSS(); rss != 0 {
		snapshot["rss_bytes"] = rss
	}
	return snapshot
}

// sanitizeDebugLine strips CR/LF so error text cannot forge log or JSON
// lines, and truncates reload errors that may carry file paths.
func sanitizeDebugLine(s string) string {
	const sanitizeMaxLen = 500
	s = strings.ReplaceAll(s, "\r", "_")
	s = strings.ReplaceAll(s, "\n", "_")
	if len(s) > sanitizeMaxLen {
		s = s[:sanitizeMaxLen]
	}
	return s
}
