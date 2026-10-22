package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/calls"
)

// saveCalls replaces the calls, voicemail, and tab tables with d.
func saveCalls(tx *sql.Tx, d calls.Data) error {
	for _, t := range []string{"calls", "voicemails", "calls_state"} {
		if _, err := tx.ExecContext(context.Background(), "DELETE FROM "+t); err != nil { //nolint:gosec // reason: table names are internal constants from a fixed list, never user input
			return err
		}
	}

	for i, c := range d.AllHistory() {
		_, err := tx.ExecContext(context.Background(), `INSERT INTO calls (id, position, name, initials, color, kind,
			time, duration, missed, incoming) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, i, c.Name, c.Initials, c.Color, c.Kind, c.Time, c.Duration,
			c.Missed, c.Incoming)
		if err != nil {
			return err
		}
	}

	for i, v := range d.AllVoicemails() {
		_, err := tx.ExecContext(context.Background(), `INSERT INTO voicemails (id, position, name, initials,
			color, time, duration, transcript) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			v.ID, i, v.Name, v.Initials, v.Color, v.Time, v.Duration, v.Transcript)
		if err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(context.Background(), `INSERT OR REPLACE INTO calls_state (id, tab) VALUES (1, ?)`, d.Tab)

	return err
}
