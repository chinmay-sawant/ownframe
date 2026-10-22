package collector

import (
	"testing"
	"time"
)

// TestManagerStress checks that a manager over the stress fixture stays
// bounded: the pool has a fixed worker count and the process snapshot is
// delivered whole.
func TestManagerStress(t *testing.T) {
	m := New(Options{
		Mode:            ModeDummy,
		Stress:          true,
		SummaryInterval: 20 * time.Millisecond,
		ProcessInterval: 40 * time.Millisecond,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if snap, ok := m.Processes(); ok && len(snap.Rows) == StressProcesses {
			stats := m.Stats()
			if stats.Pool.Workers != DefaultWorkers {
				t.Fatalf("workers = %d", stats.Pool.Workers)
			}

			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("stress snapshot never arrived")
}
