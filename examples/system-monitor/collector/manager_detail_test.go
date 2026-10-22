package collector

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestManagerTrackDetail checks that a tracked process receives details and
// that switching the selection drops stale results.
func TestManagerTrackDetail(t *testing.T) {
	m := New(Options{
		Source:          &fakeSource{},
		SummaryInterval: time.Hour,
		ProcessInterval: 10 * time.Millisecond,
		DetailInterval:  10 * time.Millisecond,
		Deadline:        time.Second,
	})
	if err := m.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = m.Close(t.Context()) }()

	first := domain.ProcessIdentity{PID: 10, Start: 1}
	m.Track(first)

	if !waitFor(t, 2*time.Second, func() bool {
		got := m.Tracked()

		return !got.Loading && got.Detail.ID == first
	}) {
		t.Fatalf("detail never arrived: %+v", m.Tracked())
	}
	if got := m.Tracked().Detail.Command; got != "one --flag" {
		t.Fatalf("command = %q", got)
	}

	second := domain.ProcessIdentity{PID: 11, Start: 2}
	m.Track(second)
	if !m.Tracked().Loading {
		t.Fatal("switching selection did not start loading")
	}
	if !waitFor(t, 2*time.Second, func() bool {
		got := m.Tracked()

		return !got.Loading && got.Detail.ID == second
	}) {
		t.Fatal("second detail never arrived")
	}

	m.Track(domain.ProcessIdentity{})
	if m.Tracked().Loading || m.Tracked().ID != (domain.ProcessIdentity{}) {
		t.Fatal("clearing the selection failed")
	}
}
