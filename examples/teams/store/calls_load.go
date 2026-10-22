package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/calls"
)

// loadCalls reads the call history, voicemails, and tab back.
func loadCalls(db *sql.DB) (out calls.Data, err error) {
	var history []calls.Call

	err = readInto(db, `SELECT id, name, initials, color, kind, time, duration,
		missed, incoming FROM calls ORDER BY position`, &history,
		func(r *sql.Rows, c *calls.Call) error {
			return r.Scan(&c.ID, &c.Name, &c.Initials, &c.Color, &c.Kind, &c.Time,
				&c.Duration, &c.Missed, &c.Incoming)
		})
	if err != nil {
		return out, err
	}

	var voicemails []calls.Voicemail

	err = readInto(db, `SELECT id, name, initials, color, time, duration, transcript
		FROM voicemails ORDER BY position`, &voicemails,
		func(r *sql.Rows, v *calls.Voicemail) error {
			return r.Scan(&v.ID, &v.Name, &v.Initials, &v.Color, &v.Time,
				&v.Duration, &v.Transcript)
		})
	if err != nil {
		return out, err
	}

	var tab string

	err = db.QueryRowContext(context.Background(), `SELECT COALESCE((SELECT tab FROM calls_state WHERE id = 1), 'history')`).Scan(&tab)
	if err != nil {
		return out, err
	}

	return calls.FromDB(history, voicemails, tab), nil
}
