package sqlkit

import (
	"bytes"
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// ctxBG is the context every statement in this file runs under.
func ctxBG() context.Context { return context.Background() }

func openMem(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestSQLiteImportEditExportRoundTrip drives the generated SQL against a real
// SQLite engine: CSV import, staged edits, and SQL export re-import.
func TestSQLiteImportEditExportRoundTrip(t *testing.T) {
	db := openMem(t)
	csvSrc := "id,name,price,active\n1,Ann,9.5,true\n2,\"B,o\",10,false\n3,Cy,,true\n"
	data, err := ParseCSV(strings.NewReader(csvSrc), ImportOptions{})
	must(t, err)
	if _, err := db.ExecContext(ctxBG(), data.CreateTableSQL("items", SQLite)); err != nil {
		t.Fatal(err)
	}
	for _, r := range data.Rows {
		if _, err := db.ExecContext(ctxBG(), data.InsertSQL("items", SQLite), r...); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	var typ string
	must(t, db.QueryRowContext(ctxBG(), "SELECT count(*), typeof(price) FROM items WHERE id = 1").Scan(&n, &typ))
	if n != 1 || typ != "real" {
		t.Fatalf("count %d type %s", n, typ)
	}

	cs := NewChangeset(Config{Dialect: SQLite, Table: "items", HasRowID: true})
	must(t, cs.UpdateCell(pk(int64(1)), "name", "Ann", "Anna"))
	must(t, cs.DeleteRow(pk(int64(2))))
	_, err = cs.InsertRow([]string{"id", "name", "price", "active"}, []any{int64(4), "Di'", 1.25, true})
	must(t, err)
	sts, err := cs.Statements()
	must(t, err)
	tx, _ := db.BeginTx(ctxBG(), nil)
	for _, s := range sts {
		if _, err := tx.ExecContext(ctxBG(), s.SQL, s.Args...); err != nil {
			t.Fatalf("%s: %v", s.SQL, err)
		}
	}
	must(t, tx.Commit())
	// The inlined preview must also be valid SQL.
	prev, _ := cs.Preview()
	if _, err := openMem(t).ExecContext(ctxBG(), "CREATE TABLE items (id, name, price, active); "+prev); err != nil {
		t.Fatalf("preview not executable: %v\n%s", err, prev)
	}

	rows, err := db.QueryContext(ctxBG(), "SELECT id, name, price, active FROM items ORDER BY id")
	must(t, err)
	defer func() { _ = rows.Close() }()
	var out [][]any
	for rows.Next() {
		r := make([]any, 4)
		ptrs := []any{&r[0], &r[1], &r[2], &r[3]}
		must(t, rows.Scan(ptrs...))
		out = append(out, r)
	}
	if len(out) != 3 || out[0][1] != "Anna" || out[1][0] != int64(3) || out[2][1] != "Di'" {
		t.Fatalf("%v", out)
	}

	var buf bytes.Buffer
	must(t, WriteSQL(&buf, []string{"id", "name", "price", "active"}, SliceRows(out), ExportOptions{Table: "copy", Dialect: SQLite, BatchSize: 2}))
	db2 := openMem(t)
	if _, err := db2.ExecContext(ctxBG(), `CREATE TABLE "copy" (id, name, price, active)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db2.ExecContext(ctxBG(), buf.String()); err != nil {
		t.Fatalf("exported SQL does not execute: %v\n%s", err, buf.String())
	}
	var name string
	must(t, db2.QueryRowContext(ctxBG(), `SELECT name FROM "copy" WHERE id = 4`).Scan(&name))
	if name != "Di'" {
		t.Fatal(name)
	}
}

func TestSQLiteSplitExecutesEachStatement(t *testing.T) {
	db := openMem(t)
	script := "CREATE TABLE t (a); -- make\nINSERT INTO t VALUES ('x;y');\nCREATE TRIGGER tr AFTER INSERT ON t BEGIN INSERT INTO t VALUES ('t'); END;\nINSERT INTO t VALUES ('z')"
	for _, st := range Script(script, SQLite) {
		if _, err := db.ExecContext(ctxBG(), st); err != nil {
			t.Fatalf("%q: %v", st, err)
		}
	}
	var n int
	must(t, db.QueryRowContext(ctxBG(), "SELECT count(*) FROM t").Scan(&n))
	if n != 3 { // 'x;y' and 'z', plus one row the trigger inserted after 'z'
		t.Fatalf("rows %d", n)
	}
}
