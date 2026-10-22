package window

import (
	"errors"
	"testing"
	"time"
)

func TestPerfSamplerValues(t *testing.T) {
	t.Parallel()
	var p perfSampler
	if !p.maybe(time.Now()) {
		t.Fatal("first sample must run")
	}
	if p.maybe(time.Now()) {
		t.Fatal("resample must gate")
	}
	if p.goroutines < 1 {
		t.Fatalf("goroutines = %d, want >= 1", p.goroutines)
	}
	s := &shell{perf: true}
	s.perfSampleRuntime()
	_, rss, _, g := s.perfRuntime()
	if g < 1 {
		t.Fatalf("shell goroutines = %d, want >= 1", g)
	}
	if got := s.dev.sampler.rss; got != rss {
		t.Fatalf("rss = %d, want %d", rss, got)
	}

	off := &shell{}
	off.perfSampleRuntime()
	off.perfFrameSample()
	if a, p95, p99, l := off.perfSummary(); a != 0 || p95 != 0 || p99 != 0 || l != 0 {
		t.Fatal("perf off must not sample")
	}
}

func TestPerfStageZeroSafe(t *testing.T) {
	t.Parallel()
	var st perfStage
	if st != (perfStage{}) {
		t.Fatal("stage must start zero")
	}
	perfTimed(&st.draw, func() {})
	if st.draw < 0 {
		t.Fatalf("draw = %v, want >= 0", st.draw)
	}
	want := errors.New("tick boom")
	if err := perfTimedErr(&st.tick, func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("timed err = %v, want %v", err, want)
	}
	var s shell
	if a, p95, p99, l := s.perfSummary(); a != 0 || p95 != 0 || p99 != 0 || l != 0 {
		t.Fatal("empty summary must read zero")
	}
}
