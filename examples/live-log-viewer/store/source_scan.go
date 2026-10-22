package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

const sourceColumns = `SELECT id, session_id, kind, path, label, identity,
	head_hash, head_len, generation, position, size, total, state, seed, rate,
	lost, updated_ns FROM sources`

type sourceScanner interface{ Scan(dest ...any) error }

func scanSource(row sourceScanner) (entry.Source, error) {
	var (
		src     entry.Source
		state   string
		head    int64
		updated int64
	)

	err := row.Scan(&src.ID, &src.Session, &src.Kind, &src.Path, &src.Label,
		&src.Identity, &head, &src.HeadLen, &src.Generation, &src.Position,
		&src.Size, &src.Total, &state, &src.Seed, &src.Rate, &src.Lost,
		&updated)
	if err != nil {
		return entry.Source{}, err
	}

	src.HeadHash = uint64(head) //nolint:gosec // reason: head hash round-trips through the signed column and back to uint64
	src.State = entry.State(state)
	src.Updated = time.Unix(0, updated)

	return src, nil
}

func sourceByPath(ctx context.Context, db *sql.DB, session entry.SessionID, path string) (entry.Source, error) {
	row := db.QueryRowContext(ctx, sourceColumns+
		` WHERE session_id = ? AND path = ?`, int64(session), path)

	return scanSource(row)
}
