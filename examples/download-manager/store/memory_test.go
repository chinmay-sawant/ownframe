package store

import (
	"context"
	"testing"
)

// TestMemoryMode keeps two in-memory stores apart.
func TestMemoryMode(t *testing.T) {
	one, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = one.Close() }()

	two, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = two.Close() }()

	if one.JournalMode() != "memory" || one.Memory() != true {
		t.Errorf("memory mode %q", one.JournalMode())
	}

	if err := one.SaveJob(context.Background(), testJob("only-one")); err != nil {
		t.Fatal(err)
	}

	if _, err := two.Job(context.Background(), "only-one"); err == nil {
		t.Error("in-memory stores share data")
	}

	if _, err := Open(""); err == nil {
		t.Error("empty directory accepted")
	}
}
