package collector

import (
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestDummyDetailGone checks that a detail read for a process that is not in
// the table reports ErrGone.
func TestDummyDetailGone(t *testing.T) {
	d := NewDummy(DummyOptions{Seed: 1, Stamp: fixedStamps()})

	if _, err := d.Detail(t.Context(), domain.ProcessIdentity{PID: 4242, Start: 1}); !errors.Is(err, domain.ErrGone) {
		t.Fatalf("err = %v, want ErrGone", err)
	}
}

// TestDummyDetailRates checks that consecutive detail reads advance CPU time
// and fill the detail-only values.
func TestDummyDetailRates(t *testing.T) {
	d := NewDummy(DummyOptions{Seed: 1, Stamp: fixedStamps()})

	rows, err := d.Processes(t.Context())
	if err != nil || len(rows) == 0 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}

	id := rows[0].ID
	first, err := d.Detail(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	second, err := d.Detail(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	if second.CPUTime < first.CPUTime {
		t.Fatal("detail CPU time went backwards")
	}
	if !second.ReadBytes.Valid || !second.Virtual.Valid {
		t.Fatal("detail values missing")
	}
}
