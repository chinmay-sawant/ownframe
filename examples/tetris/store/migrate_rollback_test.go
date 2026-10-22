package store

import (
	"context"
	"database/sql"
	"testing"
)

func TestFailedMigrationRollsBack(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)

	orig := migrations
	migrations = []string{schemaV1, `CREATE TABLE broken (;`}

	defer func() { migrations = orig }()

	ctx := context.Background()
	if err := migrate(ctx, db); err == nil {
		t.Fatal("a broken migration succeeded")
	}

	v, err := storedVersion(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	if v != 1 {
		t.Fatalf("version %d after the failed step, want 1", v)
	}

	var name string
	if err := db.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE name = 'broken'`).Scan(&name); err == nil {
		t.Fatal("the failed migration left its table behind")
	}
}
