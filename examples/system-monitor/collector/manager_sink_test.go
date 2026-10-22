package collector

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// countSink counts recorded samples.
type countSink struct {
	ch chan domain.Sample
}

func (c *countSink) RecordSample(s domain.Sample) bool {
	select {
	case c.ch <- s:
	default:
	}

	return true
}

// TestManagerSink checks that recording receives published samples and stops
// when the sink is removed.
func TestManagerSink(t *testing.T) {
	sink := &countSink{ch: make(chan domain.Sample, 4)}
	m := New(Options{
		Source:          &fakeSource{},
		SummaryInterval: 10 * time.Millisecond,
		ProcessInterval: time.Hour,
		Deadline:        time.Second,
	})
	m.SetSink(sink)

	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	select {
	case <-sink.ch:
	case <-time.After(2 * time.Second):
		t.Fatal("sink received nothing")
	}

	m.SetSink(nil)
}
