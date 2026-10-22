package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLockBlocksSecondInstance holds one lock at a time.
func TestLockBlocksSecondInstance(t *testing.T) {
	s := newStore(t)

	release, err := s.Lock()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.Lock(); !errors.Is(err, ErrLocked) {
		t.Fatalf("second lock: %v", err)
	}

	release()
	release() // idempotent

	again, err := s.Lock()
	if err != nil {
		t.Fatalf("lock after release: %v", err)
	}

	again()
}

// TestLockReclaimsStaleFile takes over a lock with no heartbeat.
func TestLockReclaimsStaleFile(t *testing.T) {
	saved := lockStale
	lockStale = time.Millisecond

	t.Cleanup(func() { lockStale = saved })

	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	path := filepath.Join(dir, lockName)
	if err := os.WriteFile(path, []byte("99999\n0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}

	release, err := s.Lock()
	if err != nil {
		t.Fatalf("stale lock not reclaimed: %v", err)
	}

	release()
}

// TestLockMemoryModeIsFree never writes a lock file in temporary mode.
func TestLockMemoryModeIsFree(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()

	release, err := s.Lock()
	if err != nil {
		t.Fatal(err)
	}

	release()
}
