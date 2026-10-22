package collector

import (
	"testing"
	"time"
)

// TestManagerPublishes checks that both loops deliver readings and process
// snapshots, and that rates appear on the second sample.
func TestManagerPublishes(t *testing.T) {
	m := New(Options{
		Source:          &fakeSource{},
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: 20 * time.Millisecond,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	if !waitFor(t, 2*time.Second, func() bool { return m.HaveReading() }) {
		t.Fatal("no reading published")
	}
	if !waitFor(t, 2*time.Second, func() bool { return m.Reading().CPUPercent.Valid }) {
		t.Fatal("no cpu rate published")
	}
	if !waitFor(t, 2*time.Second, func() bool {
		snap, ok := m.Processes()

		return ok && len(snap.Rows) == 2
	}) {
		t.Fatal("no process snapshot published")
	}

	stats := m.Stats()
	if stats.Samples < 2 || stats.ProcessSamples < 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.Source != "fake" {
		t.Fatalf("source = %q", stats.Source)
	}
}

// TestManagerSeries checks the graph buffers: keys exist and points collect.
func TestManagerSeries(t *testing.T) {
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

	key := keyCPU()
	if !waitFor(t, 2*time.Second, func() bool {
		s, ok := m.Series(key)

		return ok && len(s.Points) >= 2
	}) {
		t.Fatal("cpu series never collected two points")
	}

	if len(m.Keys()) == 0 {
		t.Fatal("no series keys")
	}
}
