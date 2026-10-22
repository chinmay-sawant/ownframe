package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDSNSpecialPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sp ace #1 ü")

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.EnsureSession(bg(), "x", "file"); err != nil {
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

	sessions, err := st2.Sessions(bg())
	if err != nil {
		t.Fatal(err)
	}

	if len(sessions) != 1 || sessions[0].Name != "x" {
		t.Fatalf("sessions = %+v", sessions)
	}
}

func TestDeletedStorageRecreates(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := st.EnsureDummy(bg(), DummyOptions{Count: 10}); err != nil {
		t.Fatal(err)
	}

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	st2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st2.Close() }()

	s, err := st2.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if s.Entries != 0 {
		t.Fatalf("recreated store has %d entries", s.Entries)
	}
}

func TestMemoryStoresAreIndependent(t *testing.T) {
	a, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = a.Close() }()

	b, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = b.Close() }()

	if _, err := a.EnsureSession(bg(), "only-a", "file"); err != nil {
		t.Fatal(err)
	}

	sa, err := a.Sessions(bg())
	if err != nil {
		t.Fatal(err)
	}

	sb, err := b.Sessions(bg())
	if err != nil {
		t.Fatal(err)
	}

	if len(sa) != 1 || len(sb) != 0 {
		t.Fatalf("memory stores share state: a=%d b=%d", len(sa), len(sb))
	}
}
