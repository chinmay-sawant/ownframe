package store

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func testPolicy() entry.Policy {
	pol := entry.DefaultPolicy()
	pol.Poll = 5 * time.Millisecond
	pol.MultilineHold = 50 * time.Millisecond

	return pol
}

func TestIngestDummyEndToEnd(t *testing.T) {
	st, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	setup, err := st.EnsureDummy(bg(), DummyOptions{Count: 300, Rate: 500})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(bg())

	var (
		wg  sync.WaitGroup
		ins []*Ingestor
	)

	for _, src := range setup.Sources {
		in, err := st.Ingest(ctx, src.ID, testPolicy())
		if err != nil {
			t.Fatal(err)
		}

		ins = append(ins, in)
		wg.Add(1)

		go func(in *Ingestor) {
			defer wg.Done()

			_ = in.Run(ctx)
		}(in)
	}

	count := waitForEntries(t, st, 320, 5*time.Second)

	if count < 320 {
		t.Fatalf("entries after streaming = %d", count)
	}

	cancel()
	wg.Wait()

	for _, in := range ins {
		if err := in.Close(); err != nil {
			t.Fatal(err)
		}
	}

	sid := setup.Session.ID

	p, err := st.Page(bg(), Query{Session: &sid}, PageOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}

	if len(p.Entries) == 0 || p.HighWater == 0 {
		t.Fatalf("page = %+v", p)
	}

	var dupes int

	if err := st.db.QueryRowContext(bg(),
		`SELECT count(*) FROM (SELECT source_id, generation, position
		 FROM entries GROUP BY source_id, generation, position
		 HAVING count(*) > 1)`).Scan(&dupes); err != nil {
		t.Fatal(err)
	}

	if dupes != 0 {
		t.Fatalf("%d duplicate positions", dupes)
	}
}

func waitForEntries(t *testing.T, st *Store, want int64, d time.Duration) int64 {
	t.Helper()

	deadline := time.Now().Add(d)

	var count int64

	for time.Now().Before(deadline) {
		s, err := st.Stats(bg())
		if err != nil {
			t.Fatal(err)
		}

		count = s.Entries
		if count >= want {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	return count
}
