package store

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenMemoryWorks(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if s.JournalMode() != "memory" {
		t.Fatalf("journal mode %q, want memory", s.JournalMode())
	}

	ctx := context.Background()

	dummy, err := s.DummyScores(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(dummy) != len(dummyScores) {
		t.Fatalf("seeded %d dummy scores, want %d", len(dummy), len(dummyScores))
	}

	top, err := s.TopScores(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 0 {
		t.Fatalf("live rankings started with %d rows", len(top))
	}
}

func TestSeparateMemoryDatabasesAreIsolated(t *testing.T) {
	a, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()
	b, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()
	ctx := context.Background()

	if err := a.SaveGame(ctx, testResult("mem-a", 500), nil); err != nil {
		t.Fatal(err)
	}

	top, err := b.TopScores(ctx, 20)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 0 {
		t.Fatalf("the second memory database saw %d scores", len(top))
	}
}

func TestClosedStoreRejectsWork(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if err := s.SaveSettings(context.Background(), DefaultSettings()); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed store returned %v", err)
	}

	if _, err := s.TopScores(context.Background(), 5); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed read returned %v", err)
	}
}

func TestDefaultDirLivesUnderOwnframeTetris(t *testing.T) {
	dir, err := DefaultDir()
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join("ownframe", AppDir)
	if !strings.HasSuffix(dir, want) {
		t.Fatalf("DefaultDir = %q, want suffix %q", dir, want)
	}
}
