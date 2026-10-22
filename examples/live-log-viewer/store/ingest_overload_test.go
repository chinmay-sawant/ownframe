package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func TestIngestOverloadLoss(t *testing.T) {
	st := newStore(t)

	sess, err := st.EnsureSession(bg(), "burst", "burst")
	if err != nil {
		t.Fatal(err)
	}

	src, err := st.AddSource(bg(), SourceSpec{
		Session: sess.ID, Kind: entry.KindBurst, Path: "burst-1",
		Rate: 10_000_000, Total: 100000,
	})
	if err != nil {
		t.Fatal(err)
	}

	pol := testPolicy()
	pol.BatchRecords = 10

	in, err := st.Ingest(bg(), src.ID, pol)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(bg())
	done := make(chan error, 1)

	go func() { done <- in.Run(ctx) }()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && in.Stats().Lost == 0 {
		time.Sleep(5 * time.Millisecond)
	}

	cancel()
	if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := in.Close(); err != nil {
		t.Fatal(err)
	}

	if in.Stats().Lost == 0 {
		t.Fatal("a fast live source reported no loss")
	}

	got, err := st.Source(bg(), src.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Lost == 0 {
		t.Fatal("loss was not persisted on the source row")
	}
}
