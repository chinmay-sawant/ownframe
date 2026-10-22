package storage

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

func TestRangeQueryPlan(t *testing.T) {
	st := openMemory(t)

	plan := ""

	err := st.do(context.Background(), func(ctx context.Context, db *sql.DB) error {
		rows, err := db.QueryContext(ctx,
			`EXPLAIN QUERY PLAN
			 SELECT row_ix, col_ix FROM cells
			 WHERE sheet_id = ? AND row_ix >= ? AND row_ix <= ? AND col_ix >= ? AND col_ix <= ?
			 ORDER BY row_ix, col_ix LIMIT ?`,
			1, 0, 10, 0, 5, 10)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var (
				id, parent, notused int
				detail              string
			)

			if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
				return err
			}

			plan += detail + "\n"
		}

		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(plan, "PRIMARY KEY") {
		t.Fatalf("deep row lookup does not use the primary key index:\n%s", plan)
	}
}
