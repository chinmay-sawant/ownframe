package transfer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUniqueDestination avoids existing and case-folded names.
func TestUniqueDestination(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "FILE.TXT"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := UniqueDestination(dir, "file.txt")
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(dir, "file (1).txt"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	got, err = UniqueDestination(dir, "fresh.bin")
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(dir, "fresh.bin"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	if _, err := UniqueDestination(dir, "CON"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "_CON")); !os.IsNotExist(err) {
		t.Error("UniqueDestination created the file")
	}
}

// TestNumbered inserts the counter before the extension.
func TestNumbered(t *testing.T) {
	if got := Numbered("a.bin", 0); got != "a.bin" {
		t.Errorf("Numbered(0) = %q", got)
	}

	if got := Numbered("a.bin", 3); got != "a (3).bin" {
		t.Errorf("Numbered(3) = %q", got)
	}

	if got := Numbered("noext", 2); got != "noext (2)" {
		t.Errorf("Numbered(2) = %q", got)
	}
}
