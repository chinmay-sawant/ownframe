package monitor

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRingWrapAround(t *testing.T) {
	r := NewRing[int](3)
	if r.Cap() != 3 || r.Len() != 0 {
		t.Fatal("new ring")
	}
	if _, ok := r.Last(); ok {
		t.Fatal("empty Last")
	}
	for i := 1; i <= 2; i++ {
		r.Push(i)
	}
	if !reflect.DeepEqual(r.Snapshot(), []int{1, 2}) {
		t.Fatal(r.Snapshot())
	}
	for i := 3; i <= 8; i++ {
		r.Push(i)
	}
	if r.Len() != 3 || r.Cap() != 3 {
		t.Fatalf("grew: len %d cap %d", r.Len(), r.Cap())
	}
	if !reflect.DeepEqual(r.Snapshot(), []int{6, 7, 8}) {
		t.Fatal(r.Snapshot())
	}
	if v, _ := r.Last(); v != 8 || r.At(0) != 6 {
		t.Fatal("At/Last")
	}
	r.Reset()
	if r.Len() != 0 || len(r.Snapshot()) != 0 {
		t.Fatal("reset")
	}
	r.Push(9)
	if r.At(0) != 9 {
		t.Fatal("push after reset")
	}
}

func TestRingDefaultCapacityAndBounds(t *testing.T) {
	r := NewRing[float64](0)
	if r.Cap() != DefaultCapacity || DefaultCapacity != 120 {
		t.Fatal(r.Cap())
	}
	for i := range 10_000 {
		r.Push(float64(i))
	}
	if r.Len() != 120 || len(r.buf) != 120 {
		t.Fatalf("unbounded: %d %d", r.Len(), len(r.buf))
	}
	defer func() {
		if recover() == nil {
			t.Fatal("out-of-range At must panic")
		}
	}()
	r.At(120)
}

func TestSeriesAndRange(t *testing.T) {
	r := NewRing[struct{ v int }](4)
	for _, v := range []int{3, 1, 4} {
		r.Push(struct{ v int }{v})
	}
	s := Series(r, func(x struct{ v int }) float64 { return float64(x.v) })
	lo, hi, ok := Range(s)
	if !ok || lo != 1 || hi != 4 || !reflect.DeepEqual(s, []float64{3, 1, 4}) {
		t.Fatal(s, lo, hi)
	}
	if _, _, ok := Range(nil); ok {
		t.Fatal("empty range")
	}
}

func TestRateTracker(t *testing.T) {
	var rt RateTracker
	t0 := time.Unix(1000, 0)
	if _, ok := rt.Update(t0, 100); ok {
		t.Fatal("first sample has no rate")
	}
	// 150 transactions over 2 s = 75 TPS.
	if rate, ok := rt.Update(t0.Add(2*time.Second), 250); !ok || rate != 75 {
		t.Fatalf("rate %v %v", rate, ok)
	}
	// Idle: zero, not "unknown".
	if rate, ok := rt.Update(t0.Add(3*time.Second), 250); !ok || rate != 0 {
		t.Fatalf("idle %v %v", rate, ok)
	}
	// Counter reset restarts the baseline instead of going negative.
	if _, ok := rt.Update(t0.Add(4*time.Second), 10); ok {
		t.Fatal("reset must not yield a rate")
	}
	if rate, ok := rt.Update(t0.Add(5*time.Second), 60); !ok || rate != 50 {
		t.Fatalf("after reset %v %v", rate, ok)
	}
	// Zero elapsed time is rejected.
	if _, ok := rt.Update(t0.Add(5*time.Second), 70); ok {
		t.Fatal("dt=0")
	}
	rt.Reset()
	if _, ok := rt.Update(t0, 1); ok {
		t.Fatal("reset tracker")
	}
}

func TestLatencyPercentiles(t *testing.T) {
	l := NewLatencyTracker(0)
	if l.P50() != 0 || l.P95() != 0 {
		t.Fatal("empty")
	}
	for i := 1; i <= 100; i++ {
		l.Observe(time.Duration(i) * time.Millisecond)
	}
	if l.P50() != 50*time.Millisecond || l.P95() != 95*time.Millisecond || l.Percentile(100) != 100*time.Millisecond || l.Percentile(0) != time.Millisecond {
		t.Fatalf("p50 %v p95 %v", l.P50(), l.P95())
	}
	if l.Count() != 100 {
		t.Fatal(l.Count())
	}
	// Window keeps only the newest samples.
	w := NewLatencyTracker(3)
	for _, ms := range []int{900, 1, 2, 3} {
		w.Observe(time.Duration(ms) * time.Millisecond)
	}
	if w.Percentile(100) != 3*time.Millisecond || w.Count() != 4 {
		t.Fatalf("window %v %d", w.Percentile(100), w.Count())
	}
	one := NewLatencyTracker(5)
	one.Observe(7 * time.Millisecond)
	if one.P50() != 7*time.Millisecond || one.P95() != 7*time.Millisecond {
		t.Fatal("single sample")
	}
}

func TestLatencyConcurrent(t *testing.T) {
	l := NewLatencyTracker(64)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				l.Observe(time.Millisecond)
				l.P95()
			}
		}()
	}
	wg.Wait()
	if l.Count() != 4000 {
		t.Fatal(l.Count())
	}
}

func TestSlowLog(t *testing.T) {
	s := NewSlowLog(0, 3)
	if s.Threshold() != 50*time.Millisecond {
		t.Fatal(s.Threshold())
	}
	if s.Observe(SlowQuery{SQL: "fast", Duration: 49 * time.Millisecond}) {
		t.Fatal("below threshold captured")
	}
	if !s.Observe(SlowQuery{SQL: "edge", Duration: 50 * time.Millisecond}) {
		t.Fatal("threshold itself must capture")
	}
	for _, sql := range []string{"b", "c", "d"} {
		s.Observe(SlowQuery{SQL: sql, Duration: time.Second})
	}
	got := s.Entries()
	if len(got) != 3 || got[0].SQL != "d" || got[2].SQL != "b" {
		t.Fatalf("%+v", got)
	}
	s.SetThreshold(2 * time.Second)
	if s.Observe(SlowQuery{SQL: "e", Duration: time.Second}) {
		t.Fatal("new threshold ignored")
	}
	s.Clear()
	if len(s.Entries()) != 0 {
		t.Fatal("clear")
	}
}
