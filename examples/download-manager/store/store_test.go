package store

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/domain"
)

// newStore opens a temporary on-disk store.
func newStore(t *testing.T) *Store {
	t.Helper()

	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = s.Close() })

	return s
}

// testJob builds a valid queued job with a unique destination.
func testJob(id string) domain.Job {
	now := time.Now().UTC().Truncate(time.Millisecond)

	return domain.Job{
		ID:          id,
		URL:         "https://example.invalid/" + id,
		Destination: "/tmp/download-manager-test/" + id,
		Name:        id,
		State:       domain.StateQueued,
		Total:       -1,
		Expected:    -1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// TestOpenMigratesAndReopens keeps a row across a close and open.
func TestOpenMigratesAndReopens(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if s.JournalMode() != "wal" {
		t.Errorf("journal %q, want wal", s.JournalMode())
	}

	if err := s.SaveJob(context.Background(), testJob("a")); err != nil {
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

	got, err := s2.Job(context.Background(), "a")
	if err != nil {
		t.Fatal(err)
	}

	if got.URL != "https://example.invalid/a" || got.State != domain.StateQueued {
		t.Errorf("row changed: %+v", got)
	}
}
