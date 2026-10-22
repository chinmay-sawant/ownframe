package store

import (
	"database/sql"
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed seed/*.sql
var seedFS embed.FS

// Seed fills an empty database from the SQL files under seed/. The files
// hold the initial data; a seeded database is marked so later runs load the
// state instead of seeding again.
func (s *Store) Seed() error {
	names, err := fs.Glob(seedFS, "seed/*.sql")
	if err != nil {
		return err
	}

	sort.Strings(names)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, name := range names {
		script, err := seedFS.ReadFile(name)
		if err != nil {
			return fail(tx, err)
		}

		if err := execScript(tx, string(script)); err != nil {
			return fail(tx, err)
		}
	}

	if _, err := tx.Exec(`INSERT OR REPLACE INTO meta (key, value) VALUES ('seeded', '1')`); err != nil {
		return fail(tx, err)
	}

	return tx.Commit()
}

// execScript runs a SQL script that holds one statement per line. The seed
// files are compiled into the binary with go:embed, so they are trusted
// build input, never user input. A line holding more than one statement is
// rejected, so one Exec call can never run stacked statements.
func execScript(tx *sql.Tx, script string) error {
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		stmt, ok := singleStatement(line)
		if !ok {
			return errMultiStatement
		}

		if _, err := tx.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
