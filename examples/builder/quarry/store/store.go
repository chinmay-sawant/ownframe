// Package store is Quarry's local config database: saved connection profiles,
// query history, snippets, and UI preferences, kept in a small SQLite file
// under os.UserConfigDir()/ownframe/quarry/state.db. It has no UI dependency.
package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

// Memory is the Open argument that selects a private in-memory store (tests).
const Memory = ":memory:"

// StateFile is the database file name inside the data directory.
// dirMode keeps the state directory private to the user; fileMaxConns bounds
// the state database pool.
const (
	dirMode      = 0o700
	fileMaxConns = 2
)

const StateFile = "state.db"

// MaxHistory bounds stored history rows; older rows are pruned on insert.
const MaxHistory = 2000

// Preference keys used by the UI.
const (
	PrefTheme = "theme" // "dark" or "light"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("store: not found")

// SavedConnection is a persisted connection profile. The password is stored
// only when SavePassword is true.
type SavedConnection struct {
	driver.ConnectionConfig
	SavePassword bool
	LastUsedAt   time.Time // zero when never opened
}

// HistoryEntry is one executed statement.
type HistoryEntry struct {
	ID           int64
	ConnectionID string
	SQL          string
	RanAt        time.Time
	Duration     time.Duration
	Rows         int64
	Error        string
}

// Snippet is a saved, named query.
type Snippet struct {
	ID        int64
	Name      string
	SQL       string
	CreatedAt time.Time
}

// Store is safe for concurrent use.
type Store struct {
	db   *sql.DB
	path string
}

// DefaultDir returns os.UserConfigDir()/ownframe/quarry.
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ownframe", "quarry"), nil
}

// Open opens the store in dir, creating it when needed. An empty dir selects
// DefaultDir; Memory selects a private in-memory database.
func Open(dir string) (*Store, error) { return OpenContext(context.Background(), dir) }

// OpenContext is Open with a caller context for the schema migration.
func OpenContext(ctx context.Context, dir string) (*Store, error) {
	if dir == "" {
		d, err := DefaultDir()
		if err != nil {
			return nil, err
		}
		dir = d
	}
	var dsn, path string
	if dir == Memory {
		path = Memory
		dsn = "file::memory:?_pragma=foreign_keys(1)"
	} else {
		if err := os.MkdirAll(dir, dirMode); err != nil {
			return nil, err
		}
		path = filepath.Join(dir, StateFile)
		esc := strings.NewReplacer("%", "%25", "?", "%3f", "#", "%23").Replace(path)
		dsn = "file:" + esc + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(wal)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if dir == Memory {
		db.SetMaxOpenConns(1) // each in-memory connection is its own database
	} else {
		db.SetMaxOpenConns(fileMaxConns)
	}
	s := &Store{db: db, path: path}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Path returns the database file path, or Memory.
func (s *Store) Path() string { return s.path }

// Close releases the database.
func (s *Store) Close() error { return s.db.Close() }

const schemaVersion = 1

func (s *Store) migrate(ctx context.Context) error {
	var v int
	if err := s.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v); err != nil {
		return err
	}
	if v >= schemaVersion {
		return nil
	}
	const ddl = `
CREATE TABLE IF NOT EXISTS connections (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    engine        TEXT NOT NULL,
    env_tag       TEXT NOT NULL DEFAULT 'Local',
    read_only     INTEGER NOT NULL DEFAULT 0,
    file_path     TEXT NOT NULL DEFAULT '',
    host          TEXT NOT NULL DEFAULT '',
    port          INTEGER NOT NULL DEFAULT 0,
    database_name TEXT NOT NULL DEFAULT '',
    user_name     TEXT NOT NULL DEFAULT '',
    password      TEXT NOT NULL DEFAULT '',
    save_password INTEGER NOT NULL DEFAULT 0,
    ssl_mode      TEXT NOT NULL DEFAULT '',
    mock          INTEGER NOT NULL DEFAULT 0,
    last_used_at  INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_id TEXT NOT NULL DEFAULT '',
    sql_text      TEXT NOT NULL,
    ran_at        INTEGER NOT NULL,
    duration_ns   INTEGER NOT NULL DEFAULT 0,
    row_count     INTEGER NOT NULL DEFAULT 0,
    error_text    TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_history_conn ON history(connection_id, id DESC);
CREATE TABLE IF NOT EXISTS snippets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    sql_text   TEXT NOT NULL,
    created_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS prefs (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
PRAGMA user_version = 1;`
	_, err := s.db.ExecContext(ctx, ddl)
	return err
}

func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// SaveConnection inserts or updates a profile and returns it with its ID. The
// password is discarded unless c.SavePassword is true. Name and engine are
// required, plus a file path (SQLite) or host and database (PostgreSQL, unless
// Mock).
func (s *Store) SaveConnection(ctx context.Context, c SavedConnection) (SavedConnection, error) {
	if strings.TrimSpace(c.Name) == "" {
		return c, errors.New("store: connection name is required")
	}
	switch c.Engine {
	case driver.EngineSQLite:
		if c.FilePath == "" {
			return c, errors.New("store: sqlite connection needs a file path")
		}
	case driver.EnginePostgres:
		if !c.Mock && (c.Host == "" || c.Database == "") {
			return c, errors.New("store: postgres connection needs host and database")
		}
	default:
		return c, fmt.Errorf("store: unknown engine %q", c.Engine)
	}
	if c.EnvTag == "" {
		c.EnvTag = driver.EnvLocal
	}
	if c.ID == "" {
		c.ID = newID()
	}
	pw := ""
	if c.SavePassword {
		pw = c.Password
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO connections(id, name, engine, env_tag, read_only, file_path, host, port, database_name,
                        user_name, password, save_password, ssl_mode, mock, last_used_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET name=excluded.name, engine=excluded.engine, env_tag=excluded.env_tag,
  read_only=excluded.read_only, file_path=excluded.file_path, host=excluded.host, port=excluded.port,
  database_name=excluded.database_name, user_name=excluded.user_name, password=excluded.password,
  save_password=excluded.save_password, ssl_mode=excluded.ssl_mode, mock=excluded.mock,
  last_used_at=excluded.last_used_at`,
		c.ID, c.Name, string(c.Engine), c.EnvTag, b2i(c.ReadOnly), c.FilePath, c.Host, c.Port, c.Database,
		c.User, pw, b2i(c.SavePassword), c.SSLMode, b2i(c.Mock), c.LastUsedAt.UnixNano())
	if err != nil {
		return c, err
	}
	if !c.SavePassword {
		c.Password = ""
	}
	return c, nil
}

const connCols = `id, name, engine, env_tag, read_only, file_path, host, port, database_name,
 user_name, password, save_password, ssl_mode, mock, last_used_at`

func scanConn(sc interface{ Scan(...any) error }) (SavedConnection, error) {
	var c SavedConnection
	var eng string
	var ro, sp, mock int
	var used int64
	err := sc.Scan(&c.ID, &c.Name, &eng, &c.EnvTag, &ro, &c.FilePath, &c.Host, &c.Port, &c.Database,
		&c.User, &c.Password, &sp, &c.SSLMode, &mock, &used)
	if err != nil {
		return c, err
	}
	c.Engine = driver.EngineKind(eng)
	c.ReadOnly, c.SavePassword, c.Mock = ro != 0, sp != 0, mock != 0
	if used > 0 {
		c.LastUsedAt = time.Unix(0, used)
	}
	return c, nil
}

// GetConnection loads one profile or returns ErrNotFound.
func (s *Store) GetConnection(ctx context.Context, id string) (SavedConnection, error) {
	c, err := scanConn(s.db.QueryRowContext(ctx, `SELECT `+connCols+` FROM connections WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

// ListConnections returns profiles, most recently used first, then by name.
func (s *Store) ListConnections(ctx context.Context) ([]SavedConnection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+connCols+` FROM connections ORDER BY last_used_at DESC, name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []SavedConnection
	for rows.Next() {
		c, err := scanConn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DeleteConnection removes a profile and its history.
func (s *Store) DeleteConnection(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM history WHERE connection_id = ?`, id); err != nil {
		return err
	}
	r, err := s.db.ExecContext(ctx, `DELETE FROM connections WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// TouchConnection stamps a profile as just used (recent connections).
func (s *Store) TouchConnection(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE connections SET last_used_at = ? WHERE id = ?`, time.Now().UnixNano(), id)
	return err
}

// RecentSQLiteFile upserts a SQLite file profile for path (drag-and-drop or
// File > Open), named after the file, and stamps it as used. An existing
// profile for the same path is reused.
func (s *Store) RecentSQLiteFile(ctx context.Context, path string) (SavedConnection, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM connections WHERE engine = 'sqlite' AND file_path = ?`, path).Scan(&id)
	c := SavedConnection{ConnectionConfig: driver.ConnectionConfig{
		Name: filepath.Base(path), Engine: driver.EngineSQLite, EnvTag: driver.EnvLocal, FilePath: path,
	}}
	if err == nil {
		c, err = s.GetConnection(ctx, id)
		if err != nil {
			return c, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return c, err
	}
	c.LastUsedAt = time.Now()
	return s.SaveConnection(ctx, c)
}

// AddHistory records an executed statement and prunes old rows.
func (s *Store) AddHistory(ctx context.Context, h HistoryEntry) (int64, error) {
	if h.RanAt.IsZero() {
		h.RanAt = time.Now()
	}
	r, err := s.db.ExecContext(ctx,
		`INSERT INTO history(connection_id, sql_text, ran_at, duration_ns, row_count, error_text) VALUES (?,?,?,?,?,?)`,
		h.ConnectionID, h.SQL, h.RanAt.UnixNano(), int64(h.Duration), h.Rows, h.Error)
	if err != nil {
		return 0, err
	}
	id, _ := r.LastInsertId()
	_, err = s.db.ExecContext(ctx, `DELETE FROM history WHERE id <= ?`, id-MaxHistory)
	return id, err
}

// ListHistory returns the newest entries first; connectionID "" means all.
func (s *Store) ListHistory(ctx context.Context, connectionID string, limit int) ([]HistoryEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT id, connection_id, sql_text, ran_at, duration_ns, row_count, error_text FROM history`
	args := []any{}
	if connectionID != "" {
		q += ` WHERE connection_id = ?`
		args = append(args, connectionID)
	}
	q += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []HistoryEntry
	for rows.Next() {
		var h HistoryEntry
		var ran, dur int64
		if err := rows.Scan(&h.ID, &h.ConnectionID, &h.SQL, &ran, &dur, &h.Rows, &h.Error); err != nil {
			return nil, err
		}
		h.RanAt, h.Duration = time.Unix(0, ran), time.Duration(dur)
		out = append(out, h)
	}
	return out, rows.Err()
}

// ClearHistory deletes history for a connection, or all when id is "".
func (s *Store) ClearHistory(ctx context.Context, connectionID string) error {
	if connectionID == "" {
		_, err := s.db.ExecContext(ctx, `DELETE FROM history`)
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM history WHERE connection_id = ?`, connectionID)
	return err
}

// SaveSnippet inserts (ID == 0) or updates a snippet and returns it.
func (s *Store) SaveSnippet(ctx context.Context, sn Snippet) (Snippet, error) {
	if strings.TrimSpace(sn.Name) == "" || strings.TrimSpace(sn.SQL) == "" {
		return sn, errors.New("store: snippet needs a name and SQL")
	}
	if sn.ID == 0 {
		sn.CreatedAt = time.Now()
		r, err := s.db.ExecContext(ctx, `INSERT INTO snippets(name, sql_text, created_at) VALUES (?,?,?)`,
			sn.Name, sn.SQL, sn.CreatedAt.UnixNano())
		if err != nil {
			return sn, err
		}
		sn.ID, _ = r.LastInsertId()
		return sn, nil
	}
	r, err := s.db.ExecContext(ctx, `UPDATE snippets SET name = ?, sql_text = ? WHERE id = ?`, sn.Name, sn.SQL, sn.ID)
	if err != nil {
		return sn, err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return sn, ErrNotFound
	}
	return sn, nil
}

// ListSnippets returns snippets by name.
func (s *Store) ListSnippets(ctx context.Context) ([]Snippet, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, sql_text, created_at FROM snippets ORDER BY name COLLATE NOCASE, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []Snippet
	for rows.Next() {
		var sn Snippet
		var ts int64
		if err := rows.Scan(&sn.ID, &sn.Name, &sn.SQL, &ts); err != nil {
			return nil, err
		}
		sn.CreatedAt = time.Unix(0, ts)
		out = append(out, sn)
	}
	return out, rows.Err()
}

// DeleteSnippet removes a snippet.
func (s *Store) DeleteSnippet(ctx context.Context, id int64) error {
	r, err := s.db.ExecContext(ctx, `DELETE FROM snippets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetPref returns the stored preference or def when unset.
func (s *Store) GetPref(ctx context.Context, key, def string) string {
	var v string
	if err := s.db.QueryRowContext(ctx, `SELECT value FROM prefs WHERE key = ?`, key).Scan(&v); err != nil {
		return def
	}
	return v
}

// SetPref stores a preference.
func (s *Store) SetPref(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO prefs(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// Theme returns "dark" (the default) or "light".
func (s *Store) Theme(ctx context.Context) string {
	if s.GetPref(ctx, PrefTheme, "dark") == "light" {
		return "light"
	}
	return "dark"
}

// SetTheme persists the theme preference; dark is true for QuarryDark.
func (s *Store) SetTheme(ctx context.Context, dark bool) error {
	v := "light"
	if dark {
		v = "dark"
	}
	return s.SetPref(ctx, PrefTheme, v)
}
