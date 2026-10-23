package draw

import "testing"

func TestPerfBeginSample(t *testing.T) {
	t.Parallel()

	due, clear := BeginSample(true, true)
	if !due.Track || !due.MemOK || clear || due.Mem.TotalAlloc == 0 {
		t.Fatal("due")
	}

	held, clear := BeginSample(true, false)
	if !held.Track || held.MemOK || clear || held.Mem.TotalAlloc != 0 {
		t.Fatal("not due")
	}

	off, clear := BeginSample(false, false)
	if off.Track || off.MemOK || !clear {
		t.Fatal("off")
	}
	off.StartClock()
	if off.Start.IsZero() {
		t.Fatal("clock")
	}
}
