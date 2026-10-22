package store

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// TestReconcileRenamesReusedDestination moves an active job aside when an
// outside file took its path.
func TestReconcileRenamesReusedDestination(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, dest := reconcileJob(t, s, "reused", domain.StatePaused, 10)

	if err := os.WriteFile(dest, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Renamed != 1 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "reused")
	if got.Destination == dest || got.State != domain.StatePaused {
		t.Errorf("job %+v", got)
	}
}

// TestReconcileRunningWithoutFiles pauses with zero progress.
func TestReconcileRunningWithoutFiles(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	reconcileJob(t, s, "empty", domain.StateRunning, 10)

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Recovered != 1 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "empty")
	if got.State != domain.StatePaused || got.Done != 0 {
		t.Errorf("job %+v", got)
	}
}
