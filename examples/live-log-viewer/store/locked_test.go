package store

import (
	"database/sql"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func TestLockedDatabase(t *testing.T) {
	st := newStore(t)

	setup, err := st.EnsureDummy(bg(), DummyOptions{Count: 3})
	if err != nil {
		t.Fatal(err)
	}

	// Surface the lock immediately instead of waiting out busy_timeout.
	if _, err := st.db.ExecContext(bg(), "PRAGMA busy_timeout = 0"); err != nil {
		t.Fatal(err)
	}

	raw, err := sql.Open("sqlite", dsn(st.Path()))
	if err != nil {
		t.Fatal(err)
	}

	raw.SetMaxOpenConns(1)

	if _, err := raw.ExecContext(bg(), "BEGIN EXCLUSIVE"); err != nil {
		t.Fatal(err)
	}

	src := setup.Sources[0]

	_, err = st.Commit(bg(), src.ID, CommitMeta{
		Generation: 1, Position: src.Position, State: entry.StateLive,
	}, nil)
	if err == nil {
		t.Fatal("commit succeeded while the database was locked")
	}

	if _, err := raw.ExecContext(bg(), "ROLLBACK"); err != nil {
		t.Fatal(err)
	}

	_ = raw.Close()

	if _, err := st.Stats(bg()); err != nil {
		t.Fatalf("store unusable after unlock: %v", err)
	}
}
