package store

import (
	"context"
	"testing"
)

// TestSeedReopensWithoutDuplicates reopens the same directory.
func TestSeedReopensWithoutDuplicates(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SeedDummy(ctx, 50); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s2.Close() }()

	if err := s2.SeedDummy(ctx, 50); err != nil {
		t.Fatal(err)
	}

	counts, err := s2.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if counts.Total != 50 {
		t.Errorf("reopen duplicated rows: %d", counts.Total)
	}
}
