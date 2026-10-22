package collector

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestManagerDetailGone checks that a gone process ends loading with an error
// the UI can show.
func TestManagerDetailGone(t *testing.T) {
	m := New(Options{
		Source:          &fakeSource{gone: true},
		SummaryInterval: time.Hour,
		ProcessInterval: 10 * time.Millisecond,
		DetailInterval:  10 * time.Millisecond,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	m.Track(domain.ProcessIdentity{PID: 10, Start: 1})

	if !waitFor(t, 2*time.Second, func() bool {
		got := m.Tracked()

		return !got.Loading && got.Err != ""
	}) {
		t.Fatalf("no failure surfaced: %+v", m.Tracked())
	}
}
