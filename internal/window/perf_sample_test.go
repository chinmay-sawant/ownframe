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

func TestPerfFrameCostExcludesParkedGap(t *testing.T) {
	s := &shell{perf: true}
	s.dev.frame = 125 * time.Millisecond
	s.dev.stages.pendingUpdate = 2 * time.Millisecond
	s.dev.stages.draw = 3 * time.Millisecond
	s.dev.stages.present = time.Millisecond
	s.perfFrameSample()
	avg, _, p99, long := s.perfSummary()
	if avg != 6*time.Millisecond || p99 != 6*time.Millisecond || long != 0 {
		t.Fatalf("work avg=%s p99=%s long=%d", avg, p99, long)
	}
	if s.dev.stages.pendingUpdate != 0 {
		t.Fatal("update work counted twice")
	}
	s.dev.stages.draw = 20 * time.Millisecond
	s.perfFrameSample()
	if _, _, _, long := s.perfSummary(); long != 1 {
		t.Fatal("20ms draw must exceed the 60Hz budget")
	}
}
