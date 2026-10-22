package transfer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCheckChecksum passes on a match and fails on a mismatch.
func TestCheckChecksum(t *testing.T) {
	path := filepath.Join(t.TempDir(), "body.bin")

	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	// sha256("hello")
	const sum = "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	if err := CheckChecksum(path, sum); err != nil {
		t.Fatalf("good checksum failed: %v", err)
	}

	if err := CheckChecksum(path, ""); err != nil {
		t.Fatalf("empty checksum failed: %v", err)
	}

	if err := CheckChecksum(path, "sha256:00"); err == nil {
		t.Fatal("mismatch accepted")
	}

	if err := CheckChecksum(path, "md5:00"); err == nil {
		t.Fatal("unsupported algorithm accepted")
	}
}

// TestFinalizeMovesAndRefusesOverwrite covers the rename contract.
func TestFinalizeMovesAndRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	partial := filepath.Join(dir, "a.part")
	dest := filepath.Join(dir, "a.bin")

	if err := os.WriteFile(partial, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := Finalize(partial, dest, FinalizeOptions{Sync: true}); err != nil {
		t.Fatalf("finalize: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("dest missing: %v", err)
	}

	if err := os.WriteFile(partial, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := Finalize(partial, dest, FinalizeOptions{})
	if err == nil {
		t.Fatal("overwrite accepted")
	}

	if _, statErr := os.Stat(partial); statErr != nil {
		t.Error("partial removed after refused finalize")
	}

	data, readErr := os.ReadFile(dest) //nolint:gosec // reason: test reads back the file it wrote.
	if readErr != nil || string(data) != "data" {
		t.Errorf("dest changed: %q %v", data, readErr)
	}
}
