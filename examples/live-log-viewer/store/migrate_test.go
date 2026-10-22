package store

import (
	"testing"
)

func TestMigrateReopen(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	got, err := st.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if got.SchemaVersion != schemaVersion || got.Journal != "wal" {
		t.Fatalf("version=%d journal=%q", got.SchemaVersion, got.Journal)
	}

	if err := st.Checkpoint(bg()); err != nil {
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

	got2, err := st2.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if got2.SchemaVersion != schemaVersion {
		t.Fatalf("reopen version = %d", got2.SchemaVersion)
	}

	var rows int

	if err := st2.db.QueryRow(
		`SELECT count(*) FROM meta WHERE key = 'schema_version'`).Scan(&rows); err != nil {
		t.Fatal(err)
	}

	if rows != 1 {
		t.Fatalf("schema_version rows = %d", rows)
	}
}
