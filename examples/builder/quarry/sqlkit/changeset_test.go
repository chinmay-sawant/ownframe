package sqlkit

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func pk(v any) RowID { return RowID{Cols: []string{"id"}, Vals: []any{v}} }

func cfg(d Dialect) Config {
	return Config{Dialect: d, Table: "orders", HasRowID: true}
}

func TestChangesetSQLiteGeneration(t *testing.T) {
	c := NewChangeset(cfg(SQLite))
	must(t, c.UpdateCell(pk(1), "status", "open", "paid"))
	must(t, c.UpdateCell(pk(1), "total", 5.0, 9.5))
	_, err := c.InsertRow([]string{"customer_id", "status"}, []any{7, nil})
	must(t, err)
	must(t, c.DeleteRow(pk(3)))
	sts, err := c.Statements()
	must(t, err)
	want := []Statement{
		{"UPDATE", `UPDATE "orders" SET "status" = ?, "total" = ? WHERE "id" = ?`, []any{"paid", 9.5, 1}},
		{"INSERT", `INSERT INTO "orders" ("customer_id", "status") VALUES (?, ?)`, []any{7, nil}},
		{"DELETE", `DELETE FROM "orders" WHERE "id" = ?`, []any{3}},
	}
	if !reflect.DeepEqual(sts, want) {
		t.Fatalf("got %#v\nwant %#v", sts, want)
	}
	if c.Pending() != 4 {
		t.Fatalf("pending %d", c.Pending())
	}
}

func TestChangesetPostgresGeneration(t *testing.T) {
	c := NewChangeset(Config{Dialect: Postgres, Schema: "public", Table: "we\"ird", HasRowID: true})
	must(t, c.UpdateCell(RowID{Cols: []string{"a", "b"}, Vals: []any{1, "x"}}, "n", 1, 2))
	must(t, c.DeleteRow(RowID{Cols: []string{"ctid"}, Vals: []any{"(0,1)"}}))
	sts, err := c.Statements()
	must(t, err)
	if sts[0].SQL != `UPDATE "public"."we""ird" SET "n" = $1 WHERE "a" = $2 AND "b" = $3` {
		t.Fatal(sts[0].SQL)
	}
	if !reflect.DeepEqual(sts[0].Args, []any{2, 1, "x"}) {
		t.Fatal(sts[0].Args)
	}
	if sts[1].SQL != `DELETE FROM "public"."we""ird" WHERE "ctid" = $1::tid` {
		t.Fatal(sts[1].SQL)
	}
}

func TestChangesetRefusals(t *testing.T) {
	ro := NewChangeset(Config{Table: "t", ReadOnly: true, HasRowID: true})
	if err := ro.UpdateCell(pk(1), "a", 1, 2); !errors.Is(err, ErrReadOnly) {
		t.Fatal(err)
	}
	if _, err := ro.InsertRow([]string{"a"}, []any{1}); !errors.Is(err, ErrReadOnly) {
		t.Fatal(err)
	}
	if err := ro.DeleteRow(pk(1)); !errors.Is(err, ErrReadOnly) {
		t.Fatal(err)
	}
	if _, err := ro.Statements(); !errors.Is(err, ErrReadOnly) {
		t.Fatal(err)
	}
	noid := NewChangeset(Config{Table: "v"})
	if err := noid.UpdateCell(pk(1), "a", 1, 2); !errors.Is(err, ErrNoRowID) {
		t.Fatal(err)
	}
	if err := noid.DeleteRow(RowID{}); !errors.Is(err, ErrNoRowID) {
		t.Fatal(err)
	}
	if noid.Pending() != 0 {
		t.Fatal("refused edits must not stage")
	}
	// Inserts need no row identifier.
	if _, err := noid.InsertRow([]string{"a"}, []any{1}); err != nil {
		t.Fatal(err)
	}
	if _, err := NewChangeset(cfg(SQLite)).Statements(); !errors.Is(err, ErrNoChange) {
		t.Fatal(err)
	}
}

func TestChangesetEditSemantics(t *testing.T) {
	c := NewChangeset(cfg(SQLite))
	must(t, c.UpdateCell(pk(1), "a", "x", "x")) // unchanged
	if c.Pending() != 0 {
		t.Fatal("no-op edit staged")
	}
	must(t, c.UpdateCell(pk(1), "a", "x", "y"))
	must(t, c.UpdateCell(pk(1), "a", "x", "z")) // overwrite, original stays "x"
	if v, ok := c.EditedValue(pk(1), "a"); !ok || v != "z" || c.Pending() != 1 {
		t.Fatalf("%v %v %d", v, ok, c.Pending())
	}
	must(t, c.UpdateCell(pk(1), "a", "x", "x")) // revert to original drops it
	if c.Pending() != 0 {
		t.Fatalf("revert left %d", c.Pending())
	}
	// ID types are part of the key.
	must(t, c.UpdateCell(pk(1), "a", 1, 2))
	if _, ok := c.EditedValue(pk("1"), "a"); ok {
		t.Fatal("int 1 and string \"1\" must differ")
	}
	// Deleting drops pending edits and blocks further ones.
	must(t, c.DeleteRow(pk(1)))
	if c.Pending() != 1 || !c.IsDeleted(pk(1)) {
		t.Fatalf("pending %d", c.Pending())
	}
	if err := c.UpdateCell(pk(1), "a", 1, 3); err == nil {
		t.Fatal("edit of deleted row accepted")
	}
	must(t, c.DeleteRow(pk(1))) // idempotent
	if c.Pending() != 1 {
		t.Fatal("duplicate delete staged")
	}
	if !c.UnstageDelete(pk(1)) || c.Pending() != 0 {
		t.Fatal("unstage delete")
	}
	h, _ := c.InsertRow([]string{"a"}, []any{1})
	if len(c.Inserts()) != 1 || !c.RemoveInsert(h) || c.RemoveInsert(h) || c.Pending() != 0 {
		t.Fatal("insert handle")
	}
	must(t, c.UpdateCell(pk(2), "a", 1, 2))
	c.Discard()
	if c.Pending() != 0 {
		t.Fatal("discard")
	}
}

func TestChangesetEmptyInsertAndMismatch(t *testing.T) {
	c := NewChangeset(cfg(Postgres))
	if _, err := c.InsertRow([]string{"a"}, nil); err == nil {
		t.Fatal("mismatch accepted")
	}
	if _, err := c.InsertRow(nil, nil); err != nil {
		t.Fatal(err)
	}
	sts, _ := c.Statements()
	if sts[0].SQL != `INSERT INTO "orders" DEFAULT VALUES` {
		t.Fatal(sts[0].SQL)
	}
}

func TestChangesetPreviewInlinesLiterals(t *testing.T) {
	c := NewChangeset(cfg(Postgres))
	must(t, c.UpdateCell(pk(1), "note", nil, "it's"))
	must(t, c.DeleteRow(pk(2)))
	p, err := c.Preview()
	must(t, err)
	want := "UPDATE \"orders\" SET \"note\" = 'it''s' WHERE \"id\" = 1;\nDELETE FROM \"orders\" WHERE \"id\" = 2;\n"
	if p != want {
		t.Fatalf("got\n%s\nwant\n%s", p, want)
	}
	// A '?' or "$1" inside a string literal is not a placeholder.
	got := Inline(`UPDATE t SET a = '?' WHERE b = ?`, []any{5}, SQLite)
	if got != `UPDATE t SET a = '?' WHERE b = 5` {
		t.Fatal(got)
	}
	if got := RowIDFallback(Postgres) + RowIDFallback(SQLite); got != "ctidrowid" {
		t.Fatal(got)
	}
	if !strings.HasPrefix(NewChangeset(cfg(SQLite)).Config().Table, "orders") {
		t.Fatal("config")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
