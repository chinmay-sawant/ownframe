package storage

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/collector"
	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestCoreFlowReopen checks that a recorded session and the seeded fixture
// survive a reopen, and that the fixture is not written twice.
func TestCoreFlowReopen(t *testing.T) {
	dir, id := coreFlow(t)
	ctx := t.Context()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	sessions, err := st.Sessions(ctx, 10)
	if err != nil || len(sessions) != 2 {
		t.Fatalf("sessions = %+v err = %v", sessions, err)
	}

	if version, err := st.FixtureVersion(ctx); err != nil || version != 1 {
		t.Fatalf("version = %d err = %v", version, err)
	}

	rows, err := st.History(ctx, Query{SessionID: id, Metric: domain.MetricCPU, Limit: 100})
	if err != nil || len(rows) < 5 {
		t.Fatalf("reopened rows = %d err = %v", len(rows), err)
	}

	got, ok, err := st.Session(ctx, id)
	if err != nil || !ok || got.State != StateDone || got.Rows < 5 {
		t.Fatalf("session = %+v ok=%v err=%v", got, ok, err)
	}
}

// TestCoreFlowSeedTwice checks that reopening and seeding the same fixture
// version again writes nothing.
func TestCoreFlowSeedTwice(t *testing.T) {
	dir, _ := coreFlow(t)
	ctx := t.Context()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	wrote, err := st.Seed(ctx, collector.DummyFixture(1, time.Now(), 120, 15*time.Second))
	if err != nil || wrote {
		t.Fatalf("second seed wrote=%v err=%v", wrote, err)
	}
}
