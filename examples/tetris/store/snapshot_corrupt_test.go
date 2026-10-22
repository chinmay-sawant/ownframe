package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestStoredSnapshotThatNoLongerValidatesIsRejected(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	g := game.New(5)
	g.Start()

	if err := s.SaveSnapshot(ctx, g.Snapshot()); err != nil {
		t.Fatal(err)
	}

	s.exec(t, `UPDATE snapshot SET payload = '{"id":"x"}' WHERE id = 1`)

	if _, _, err := s.LoadSnapshot(ctx); !errors.Is(err, ErrInvalid) {
		t.Fatalf("corrupt snapshot returned %v", err)
	}
}

func TestSnapshotUpsertReplacesTheSingleSlot(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	for _, seed := range []uint64{1, 2} {
		g := game.New(seed)
		g.Start()

		if err := s.SaveSnapshot(ctx, g.Snapshot()); err != nil {
			t.Fatal(err)
		}
	}

	var rows int

	err = s.do(ctx, func(ctx context.Context, db *sql.DB) error {
		return db.QueryRowContext(ctx, `SELECT COUNT(*) FROM snapshot`).Scan(&rows)
	})
	if err != nil {
		t.Fatal(err)
	}

	if rows != 1 {
		t.Fatalf("snapshot table holds %d rows", rows)
	}
}
