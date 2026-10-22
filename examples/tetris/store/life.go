package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store is one open database with one serialized worker.
type Store struct {
	db   *sqlDB
	path string
	mode string

	mu     sync.Mutex
	closed bool
	reqs   chan *request
	quit   chan struct{}
	done   chan struct{}
	once   sync.Once
}

// Open opens the database in dir, creating it when needed. An empty dir
// uses DefaultDir.
func Open(dir string) (*Store, error) {
	if dir == "" {
		d, err := DefaultDir()
		if err != nil {
			return nil, err
		}

		dir = d
	}

	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return nil, fmt.Errorf("store: data dir: %w", err)
	}

	path := filepath.Join(dir, DBName)

	return open(fileDSN(path), path)
}

// OpenMemory opens a private in-memory database for tests.
func OpenMemory() (*Store, error) { return open(memoryDSN(), Memory) }

// Close drains queued work, checkpoints, and closes the database. Call it
// after producers stop sending.
func (s *Store) Close() error {
	if s == nil {
		return nil
	}

	s.once.Do(func() {
		s.mu.Lock()
		s.closed = true
		close(s.quit)
		s.mu.Unlock()
		<-s.done
	})

	return nil
}

// JournalMode reports the active journal mode: "wal", a rollback mode
// such as "delete", or "memory" for an in-memory database.
func (s *Store) JournalMode() string { return s.mode }

// QueueDepth reports requests waiting for the worker.
func (s *Store) QueueDepth() int { return len(s.reqs) }
