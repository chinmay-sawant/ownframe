package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/files"
)

// saveFiles replaces the files and filter tables with d.
func saveFiles(tx *sql.Tx, d files.Data) error {
	for _, t := range []string{"files", "files_state"} {
		if _, err := tx.Exec("DELETE FROM " + t); err != nil { //nolint:gosec // reason: table names are internal constants from a fixed list, never user input
			return err
		}
	}

	for i, f := range d.AllFiles() {
		if _, err := tx.Exec(`INSERT INTO files (id, position, name, badge, badge_text, kind, modified, modified_by, size, location, team, shared, starred) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, f.ID, i, f.Name, f.Badge, f.BadgeText, f.Kind, f.Modified, f.ModifiedBy, f.Size, f.Where, f.Team, f.Shared, f.Starred); err != nil {
			return err
		}
	}

	_, err := tx.Exec(`INSERT OR REPLACE INTO files_state (id, filter, active) VALUES (1, ?, ?)`, d.Filter, d.Active)

	return err
}

// loadFiles reads the file rows and the filter state back.
func loadFiles(db *sql.DB) (out files.Data, err error) {
	var all []files.File

	err = readInto(db, `SELECT id, name, badge, badge_text, kind, modified, modified_by, size, location, team, shared, starred FROM files ORDER BY position`, &all, func(r *sql.Rows, f *files.File) error {
		return r.Scan(&f.ID, &f.Name, &f.Badge, &f.BadgeText, &f.Kind, &f.Modified, &f.ModifiedBy, &f.Size, &f.Where, &f.Team, &f.Shared, &f.Starred)
	})
	if err != nil {
		return out, err
	}

	var filter, active string

	err = db.QueryRow(`SELECT filter, active FROM files_state WHERE id = 1`).Scan(&filter, &active)
	if err != nil && err != sql.ErrNoRows {
		return out, err
	}

	return files.FromDB(all, filter, active), nil
}
