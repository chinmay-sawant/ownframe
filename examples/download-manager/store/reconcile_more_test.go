package store

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/transfer"
)

// TestReconcileTruncatesLargerPartial restarts a partial that outgrew the
// expected length.
func TestReconcileTruncatesLargerPartial(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, dest := reconcileJob(t, s, "larger", domain.StateRunning, 4)

	if err := os.WriteFile(transfer.PartialPath(dest), []byte("toolong"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Reset != 1 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "larger")
	if got.Done != 0 || got.State != domain.StatePaused {
		t.Errorf("job %+v", got)
	}
}

// TestReconcileMissingFinalFile fails a completed row whose file vanished.
func TestReconcileMissingFinalFile(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	reconcileJob(t, s, "gone", domain.StateCompleted, 4)

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Missing != 1 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "gone")
	if got.State != domain.StateFailed {
		t.Errorf("job state %s", got.State)
	}
}
