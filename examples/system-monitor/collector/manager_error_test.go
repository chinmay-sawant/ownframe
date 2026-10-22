package collector

import (
	"testing"
	"time"
)

// TestManagerClearsErrorOnRecovery checks that a failing step is reported and
// that a later success removes it, so the UI stops showing a stale failure.
func TestManagerClearsErrorOnRecovery(t *testing.T) {
	src := &fakeSource{fail: true}

	m := New(Options{
		Source:          src,
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: time.Hour,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	if !waitFor(t, 2*time.Second, func() bool { return len(m.Errors()) > 0 }) {
		t.Fatal("failure never reported")
	}

	src.setFail(false)

	if !waitFor(t, 2*time.Second, func() bool { return len(m.Errors()) == 0 }) {
		t.Fatalf("failure not cleared: %+v", m.Errors())
	}
}

// TestManagerNoDiskIORates checks that a source without disk IO counters
// leaves its disk rates unavailable instead of reporting 0 B/s.
func TestManagerNoDiskIORates(t *testing.T) {
	m := New(Options{
		Source:          &fakeSource{},
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: time.Hour,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	if !waitFor(t, 2*time.Second, func() bool { return m.HaveReading() }) {
		t.Fatal("no reading")
	}

	for _, disk := range m.Reading().Disks {
		if disk.ReadRate.Valid || disk.WriteRate.Valid {
			t.Fatalf("disk %s reported a rate without IO counters", disk.Device)
		}
	}
}
