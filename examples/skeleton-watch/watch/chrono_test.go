package watch

import (
	"testing"
	"time"
)

func TestChronoTiming(t *testing.T) {
	c := &Chrono{}
	t0 := time.Now()
	if c.Running() || c.Elapsed(t0) != 0 {
		t.Fatal("chrono should start inactive at 0s")
	}

	c.Start(t0)
	if !c.Running() {
		t.Fatal("chrono should be running")
	}

	t1 := t0.Add(5 * time.Second)
	if c.Elapsed(t1) != 5*time.Second {
		t.Fatalf("expected 5s, got %v", c.Elapsed(t1))
	}

	c.Stop(t1)
	if c.Running() {
		t.Fatal("chrono should be stopped")
	}

	t2 := t1.Add(10 * time.Second)
	if c.Elapsed(t2) != 5*time.Second {
		t.Fatalf("expected 5s held, got %v", c.Elapsed(t2))
	}

	c.Reset(t2)
	if c.Elapsed(t2) != 0 {
		t.Fatalf("expected 0 after reset, got %v", c.Elapsed(t2))
	}
}

func TestChronoLapAndFormat(t *testing.T) {
	c := &Chrono{}
	t0 := time.Now()
	c.Start(t0)

	t1 := t0.Add(65 * time.Second)
	c.Lap(t1)
	if !c.lapActive || c.DisplayDuration(t1) != 65*time.Second {
		t.Fatal("lap should freeze display duration")
	}

	formatted := c.Format(t1)
	if formatted != "01:05.00" {
		t.Fatalf("expected 01:05.00, got %s", formatted)
	}

	c.Lap(t1.Add(10 * time.Second))
	if c.lapActive {
		t.Fatal("lap should be deactivated")
	}
}

func TestPowerReserve(t *testing.T) {
	pr := newPowerReserve(48.0)
	if pr.Percent() != 67 {
		t.Fatalf("expected 67%%, got %d%%", pr.Percent())
	}
	pr.Wind(10.0)
	if pr.Hours() != 58.0 {
		t.Fatalf("expected 58.0h, got %f", pr.Hours())
	}
	pr.Wind(30.0)
	if pr.Hours() != MaxPowerReserve {
		t.Fatalf("expected clamped %f, got %f", MaxPowerReserve, pr.Hours())
	}

	pr.Drain(3600.0) // drain 1 hour
	if pr.Hours() != MaxPowerReserve-1.0 {
		t.Fatalf("expected %f, got %f", MaxPowerReserve-1.0, pr.Hours())
	}

	empty := newPowerReserve(0)
	if !empty.Depleted() {
		t.Fatal("expected empty reserve to be depleted")
	}
}
