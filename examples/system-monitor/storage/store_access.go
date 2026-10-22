package storage

import "context"

// Close flushes WAL and closes the database. It is safe to call twice. The
// checkpoint matters on Windows: a rename or delete of the database needs the
// WAL folded back first, and the WAL and SHM files are never deleted by hand.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	if s.closed.Swap(true) {
		return nil
	}

	if s.journal == "wal" {
		_, _ = s.db.ExecContext(context.Background(), "PRAGMA wal_checkpoint(TRUNCATE)")
	}

	return s.db.Close()
}

// Path returns the database file path, or ":memory:" in temporary mode.
func (s *Store) Path() string { return s.path }

// Temp reports whether this store is the in-memory temporary mode.
func (s *Store) Temp() bool { return s.temp }

// JournalMode returns the journal mode the connection actually uses, "wal",
// "delete", or "memory". A WAL request on storage that cannot support it
// falls back and reports the fallback here.
func (s *Store) JournalMode() string { return s.journal }
