package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Lock tuning. The heartbeat keeps a live lock fresh; a lock with no
// heartbeat for lockStale is treated as a crash leftover and reclaimed.
const (
	lockName      = "instance.lock"
	lockHeartbeat = 5 * time.Second
)

// lockStale is a seam tests shorten.
var lockStale = 20 * time.Second

// Lock takes the single-instance lock beside the database. The returned
// release is idempotent. A stale lock from a dead process is reclaimed; a
// live one returns ErrLocked. Temporary mode takes no lock.
func (s *Store) Lock() (func(), error) {
	if s.memory {
		return func() {}, nil
	}

	path := filepath.Join(s.dir, lockName)

	for attempt := 0; attempt < 2; attempt++ {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, storeFilePerm) //nolint:gosec // reason: lock file is a fixed name in the own data dir.
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n%d\n", os.Getpid(), nowMS())
			_ = file.Close()

			return s.heartbeat(path), nil
		}

		if !os.IsExist(err) {
			return nil, err
		}

		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}

		if time.Since(info.ModTime()) > lockStale {
			_ = os.Remove(path)

			continue
		}

		return nil, fmt.Errorf("%w: %s", ErrLocked, path)
	}

	return nil, fmt.Errorf("%w: %s", ErrLocked, path)
}

// heartbeat touches the lock file until release runs.
func (s *Store) heartbeat(path string) func() {
	stop := make(chan struct{})
	releaseOnce := sync.Once{}

	go func() {
		ticker := time.NewTicker(lockHeartbeat)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				now := time.Now()
				_ = os.Chtimes(path, now, now)
			}
		}
	}()

	return func() {
		releaseOnce.Do(func() {
			close(stop)
			_ = os.Remove(path)
		})
	}
}
