package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestNestedQueriesOnOneConnectionTimeOut(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	err = s.do(context.Background(), func(ctx context.Context, db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT id FROM scores`)
		if err != nil {
			return err
		}

		defer func() { _ = rows.Close() }()
		inner, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		_, err = db.ExecContext(inner, `UPDATE scores SET score = score`)

		return err
	})
	if err == nil {
		t.Fatal("a nested statement on one connection did not time out")
	}
}

func TestSequentialOperationsOnOneConnection(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		if _, err := s.TopScores(ctx, 5); err != nil {
			t.Fatal(err)
		}

		if err := s.SaveSettings(ctx, DefaultSettings()); err != nil {
			t.Fatal(err)
		}

		if _, err := s.DummyScores(ctx, 5); err != nil {
			t.Fatal(err)
		}
	}
}
