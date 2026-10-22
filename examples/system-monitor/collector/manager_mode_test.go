package collector

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// keyCPU is the overview CPU series key.
func keyCPU() domain.SeriesKey {
	return domain.SeriesKey{Metric: domain.MetricCPU}
}

// TestManagerSetModeResets checks that a mode switch bumps the generation and
// clears readings, baselines, and graph buffers.
func TestManagerSetModeResets(t *testing.T) {
	m := New(Options{
		Source:          &fakeSource{},
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: 10 * time.Millisecond,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	if !waitFor(t, 2*time.Second, func() bool { return m.HaveReading() }) {
		t.Fatal("no reading")
	}
	if err := m.SetMode(ModeLive); err != nil {
		t.Fatal(err)
	}

	if m.Mode() != ModeLive {
		t.Fatalf("mode = %v", m.Mode())
	}
	if len(m.Keys()) != 0 {
		t.Fatal("graph buffers survived the mode switch")
	}
	if m.Stats().Generation == 0 {
		t.Fatal("generation did not change")
	}

	if !waitFor(t, 2*time.Second, func() bool { return m.HaveReading() }) {
		t.Fatal("new mode never published a reading")
	}
}

// TestManagerSetModeRejectsClosed checks that a closed manager refuses a mode
// switch.
func TestManagerSetModeRejectsClosed(t *testing.T) {
	m := New(Options{Source: &fakeSource{}, SummaryInterval: time.Hour, ProcessInterval: time.Hour})

	if err := m.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := m.SetMode(ModeLive); err != ErrClosed {
		t.Fatalf("err = %v, want ErrClosed", err)
	}
}
