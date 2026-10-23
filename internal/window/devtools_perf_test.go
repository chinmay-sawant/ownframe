package window

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// TestDevPerfZeroState checks the new sections render dashes on zero state.
func TestDevPerfZeroState(t *testing.T) {
	s := newDevShell(newDevScreen())
	got := devFrameText(s.devFrameRows(320))

	for _, want := range []string{"PERFORMANCE", "Frame avg", "Frame p95", "Frame p99", "PIPELINE DETAIL", "Template", "MEMORY", "Go heap", "Goroutines"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

// TestDevPerfFrameHook checks the frame hook values render.
func TestDevPerfFrameHook(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devFramePerf = func() (time.Duration, time.Duration, time.Duration, uint64, bool) {
		return 16 * time.Millisecond, 20 * time.Millisecond, 25 * time.Millisecond, 3, true
	}

	got := devFrameText(s.devFrameRows(320))
	for _, want := range []string{"16.0ms", "20.0ms", "25.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
	if !strings.Contains(got, "Long frames") {
		t.Fatalf("frame rows = %q, want Long frames", got)
	}
}

// TestDevPerfPipelineHook checks the pipeline hook values render.
func TestDevPerfPipelineHook(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devPipelinePerf = func(st host.Stats) (time.Duration, time.Duration, time.Duration, time.Duration, int, int, int, bool) {
		return time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond, 4 * time.Millisecond, 5, 6, 7, true
	}

	got := devFrameText(s.devFrameRows(320))
	for _, want := range []string{"Style", "Layout", "Dirty ops", "Dirty regions", "Changed ops", "1.0ms", "4.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}
