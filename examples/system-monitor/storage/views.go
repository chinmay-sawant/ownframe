package storage

import (
	"context"
	"time"
)

// View is a saved process table filter. Query is the search text, Sort names
// the column, and Desc marks a descending sort.
type View struct {
	ID      string
	Name    string
	Query   string
	Sort    string
	Desc    bool
	Updated time.Time
}

// Views returns every saved view, most recently updated first.
func (s *Store) Views(ctx context.Context) ([]View, error) {
	if err := s.ready(); err != nil {
		return nil, err
	}

	ctx, cancel := s.opCtx(ctx)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, query, sort, descending, updated_ns
FROM saved_views ORDER BY updated_ns DESC, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []View

	for rows.Next() {
		var (
			v  View
			ns int64
			d  int
		)
		if err := rows.Scan(&v.ID, &v.Name, &v.Query, &v.Sort, &d, &ns); err != nil {
			return nil, err
		}
		v.Desc = d != 0
		v.Updated = time.Unix(0, ns)
		out = append(out, v)
	}

	return out, rows.Err()
}
