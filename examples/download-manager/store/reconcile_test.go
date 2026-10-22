package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/transfer"
)

// reconcileJob stores one row with a fresh destination directory.
func reconcileJob(t *testing.T, s *Store, id string, state domain.State, expected int64) (domain.Job, string) {
	t.Helper()

	job := testJob(id)
	job.State = state
	job.Expected = expected
	job.Name = id + ".bin"
	job.Destination = filepath.Join(t.TempDir(), id+".bin")

	if err := s.SaveJob(context.Background(), job); err != nil {
		t.Fatal(err)
	}

	return job, job.Destination
}

// TestReconcileAdoptsFinalizedFile completes a running row whose file is
// already in place, the crash-after-finalize case.
func TestReconcileAdoptsFinalizedFile(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, dest := reconcileJob(t, s, "adopt", domain.StateRunning, 4)

	if err := os.WriteFile(dest, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Completed != 1 {
		t.Fatalf("report %+v", rep)
	}

	got, _ := s.Job(ctx, "adopt")
	if got.State != domain.StateCompleted || got.Done != 4 {
		t.Errorf("job %+v", got)
	}
}

// TestReconcileFinalizesCompletePartial renames a full partial into place.
func TestReconcileFinalizesCompletePartial(t *testing.T) {
	s := newStore(t)
	ctx := context.Background()

	_, dest := reconcileJob(t, s, "finish", domain.StateRunning, 4)

	if err := os.WriteFile(transfer.PartialPath(dest), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	rep, err := s.Reconcile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if rep.Completed != 1 {
		t.Fatalf("report %+v", rep)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Errorf("final file missing: %v", err)
	}

	if _, err := os.Stat(transfer.PartialPath(dest)); !os.IsNotExist(err) {
		t.Error("partial kept after finalize")
	}
}
