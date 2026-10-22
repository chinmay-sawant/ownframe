package store

import (
	"path/filepath"
	"testing"
)

func TestBackupRestore(t *testing.T) {
	st := newStore(t)

	setup, err := st.EnsureDummy(bg(), DummyOptions{Count: 50})
	if err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), dbName)

	if err := st.Backup(bg(), dest); err != nil {
		t.Fatal(err)
	}

	if err := st.Backup(bg(), dest); err == nil {
		t.Fatal("backup overwrote an existing file")
	}

	st2, err := Open(filepath.Dir(dest))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st2.Close() }()

	got, err := st2.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	want, err := st.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if got.Entries != want.Entries || got.Sources != want.Sources {
		t.Fatalf("backup entries=%d sources=%d, want %d/%d",
			got.Entries, got.Sources, want.Entries, want.Sources)
	}

	again, err := st2.EnsureDummy(bg(), DummyOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if again.Seeded {
		t.Fatal("the backup lost the seed marker")
	}

	if _, err := st2.Page(bg(), Query{Session: &setup.Session.ID}, PageOptions{Limit: 5}); err != nil {
		t.Fatal(err)
	}
}
