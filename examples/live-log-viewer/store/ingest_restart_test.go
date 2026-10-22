package store

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func TestIngestRestartResumes(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	setup, err := st.EnsureDummy(bg(), DummyOptions{Count: 200, Rate: 500})
	if err != nil {
		t.Fatal(err)
	}

	runOne(t, st, setup.Sources[0].ID, 210, 5*time.Second)

	before, err := st.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	st2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st2.Close() }()

	setup2, err := st2.EnsureDummy(bg(), DummyOptions{Count: 200, Rate: 500})
	if err != nil {
		t.Fatal(err)
	}

	if setup2.Seeded {
		t.Fatal("restart reseeded the fixture")
	}

	runOne(t, st2, setup2.Sources[0].ID, before.Entries+5, 5*time.Second)

	after, err := st2.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if after.Entries <= before.Entries {
		t.Fatalf("entries %d -> %d after restart", before.Entries, after.Entries)
	}

	var dupes int

	if err := st2.db.QueryRow(
		`SELECT count(*) FROM (SELECT source_id, generation, position
		 FROM entries GROUP BY source_id, generation, position
		 HAVING count(*) > 1)`).Scan(&dupes); err != nil {
		t.Fatal(err)
	}

	if dupes != 0 {
		t.Fatalf("%d duplicate positions after restart", dupes)
	}
}

func runOne(t *testing.T, st *Store, srcID entry.SourceID, want int64, d time.Duration) {
	t.Helper()

	ctx, cancel := context.WithCancel(bg())

	in, err := st.Ingest(ctx, srcID, testPolicy())
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = in.Run(ctx)
	}()

	if got := waitForEntries(t, st, want, d); got < want {
		cancel()
		wg.Wait()

		t.Fatalf("entries = %d, want at least %d", got, want)
	}

	cancel()
	wg.Wait()

	if err := in.Close(); err != nil {
		t.Fatal(err)
	}
}
