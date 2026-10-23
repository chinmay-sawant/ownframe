package draw

import "testing"

func TestPerfBeginSample(t *testing.T) {
	t.Parallel()

	on, clear := BeginSample(true)
	if !on.Track || clear || on.Mem.TotalAlloc == 0 {
		t.Fatal("on")
	}

	off, clear := BeginSample(false)
	if off.Track || !clear {
		t.Fatal("off")
	}
	off.StartClock()
	if off.Start.IsZero() {
		t.Fatal("clock")
	}
}
