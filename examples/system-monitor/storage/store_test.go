package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestOpenAndReopen checks migration, reopen, and the journal mode.
func TestOpenAndReopen(t *testing.T) {
	dir := t.TempDir()

	st, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if st.Path() != filepath.Join(dir, dbName) {
		t.Fatalf("path = %q", st.Path())
	}
	if mode := st.JournalMode(); mode != "wal" && mode != "delete" {
		t.Fatalf("journal = %q", mode)
	}
	if err := st.SetSetting(t.Context(), "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	st, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	got, ok, err := st.Setting(t.Context(), "theme")
	if err != nil || !ok || got != "dark" {
		t.Fatalf("setting = %q ok=%v err=%v", got, ok, err)
	}
}

// TestForceRollback checks the documented journal fallback.
func TestForceRollback(t *testing.T) {
	st, err := OpenWithOptions(Options{Dir: t.TempDir(), ForceRollback: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if st.JournalMode() != "delete" {
		t.Fatalf("journal = %q", st.JournalMode())
	}
}

// TestTempStore checks the in-memory temporary mode.
func TestTempStore(t *testing.T) {
	st, err := OpenWithOptions(Options{Temp: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if !st.Temp() || st.Path() != memoryPath || st.JournalMode() != "memory" {
		t.Fatalf("temp=%v path=%q journal=%q", st.Temp(), st.Path(), st.JournalMode())
	}
	if _, _, err := st.Setting(t.Context(), "missing"); err != nil {
		t.Fatal(err)
	}
}

// TestDefaultDir checks the config directory override.
func TestDefaultDir(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	t.Setenv("HOME", base)

	dir, err := DefaultDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, filepath.Join("ownframe", "system-monitor")) {
		t.Fatalf("dir = %q", dir)
	}
}
