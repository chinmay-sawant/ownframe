package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/store"
)

// TestRecoverLoadsActiveJobs picks up persisted queued and paused rows.
func TestRecoverLoadsActiveJobs(t *testing.T) {
	st, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	ctx := context.Background()
	queued := testSchedulerJob("recover-queued", domain.StateQueued)
	paused := testSchedulerJob("recover-paused", domain.StatePaused)

	for _, job := range []domain.Job{queued, paused} {
		if err := st.SaveJob(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	eng, err := New(Options{
		Transport: instantTransport{}, Store: st,
		Workers: 1, Now: time.Now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := eng.Recover(ctx); err != nil {
		t.Fatal(err)
	}

	eng.Start(ctx)
	defer func() { _ = eng.Close(context.Background()) }()

	waitState(t, eng, queued.ID, domain.StateCompleted)

	if got, _ := eng.Job(paused.ID); got.State != domain.StatePaused {
		t.Errorf("paused job restarted: %s", got.State)
	}
}
