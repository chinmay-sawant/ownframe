package store

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/transfer"
)

// TestReconcileInterruptedBecomesPaused keeps the partial and pauses.
func TestReconcileInterruptedBecomesPaused(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, dest := reconcileJob(t, s, "interrupted", domain.StateRunning, 10)

	if err := os.WriteFile(transfer.PartialPath(dest), []byte("half"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Recovered != 1 || rep.Completed != 0 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "interrupted")
	if got.State != domain.StatePaused || got.Done != 4 {
		t.Errorf("job %+v", got)
	}
}
