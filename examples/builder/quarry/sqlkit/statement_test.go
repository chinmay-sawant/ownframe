package sqlkit

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name string
		src  string
		d    Dialect
		want []string
	}{
		{"empty", "  \n ", SQLite, nil},
		{"two", "SELECT 1; SELECT 2;", SQLite, []string{"SELECT 1", "SELECT 2"}},
		{"no trailing semicolon", "SELECT 1; SELECT 2", SQLite, []string{"SELECT 1", "SELECT 2"}},
		{"semicolon in string", "SELECT 'a;b'; SELECT 2", SQLite, []string{"SELECT 'a;b'", "SELECT 2"}},
		{"semicolon in comments", "SELECT 1 -- x;y\n; /* ; */ SELECT 2", SQLite, []string{"SELECT 1 -- x;y", "/* ; */ SELECT 2"}},
		{"semicolon in quoted ident", `SELECT "a;b" FROM t; SELECT 2`, SQLite, []string{`SELECT "a;b" FROM t`, "SELECT 2"}},
		{"empty statements dropped", ";;SELECT 1;;", SQLite, []string{"SELECT 1"}},
		{"comment only dropped", "SELECT 1; -- done\n", SQLite, []string{"SELECT 1"}},
		{"dollar body", "CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql; SELECT 2", Postgres,
			[]string{"CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql", "SELECT 2"}},
		{"sqlite trigger", "CREATE TRIGGER tr AFTER INSERT ON t BEGIN UPDATE u SET x = CASE WHEN 1 THEN 2 END; DELETE FROM v; END; SELECT 1", SQLite,
			[]string{"CREATE TRIGGER tr AFTER INSERT ON t BEGIN UPDATE u SET x = CASE WHEN 1 THEN 2 END; DELETE FROM v; END", "SELECT 1"}},
		{"begin transaction still splits", "BEGIN; INSERT INTO t VALUES (1); COMMIT;", SQLite,
			[]string{"BEGIN", "INSERT INTO t VALUES (1)", "COMMIT"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Script(tc.src, tc.d)
			if (len(got) != 0 || len(tc.want) != 0) && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestStatementAt(t *testing.T) {
	src := "SELECT 1;\nSELECT 'a;b' FROM t;\n\n-- c\nSELECT 3"
	at := func(caret int) string {
		sp, ok := StatementAt(src, caret, SQLite)
		if !ok {
			t.Fatalf("no statement at %d", caret)
		}
		return sp.Text(src)
	}
	tests := []struct {
		caret int
		want  string
	}{
		{0, "SELECT 1"},
		{5, "SELECT 1"},
		{9, "SELECT 1"},             // directly after the semicolon
		{10, "SELECT 'a;b' FROM t"}, // start of line 2
		{18, "SELECT 'a;b' FROM t"}, // inside the string literal
		{len("SELECT 1;\nSELECT 'a;b' FROM t;"), "SELECT 'a;b' FROM t"},
		{len("SELECT 1;\nSELECT 'a;b' FROM t;\n"), "-- c\nSELECT 3"}, // gap goes to next
		{len(src), "-- c\nSELECT 3"},
		{9999, "-- c\nSELECT 3"},
		{-4, "SELECT 1"},
	}
	for _, tc := range tests {
		if got := at(tc.caret); got != tc.want {
			t.Errorf("caret %d: got %q want %q", tc.caret, got, tc.want)
		}
	}
	if _, ok := StatementAt("  ", 1, SQLite); ok {
		t.Fatal("blank script should have no statement")
	}
}

func TestStatementAtTrailingGap(t *testing.T) {
	src := "SELECT 1;\n\n"
	sp, ok := StatementAt(src, len(src), SQLite)
	if !ok || sp.Text(src) != "SELECT 1" {
		t.Fatalf("got %v %v", sp, ok)
	}
}

func TestRunTarget(t *testing.T) {
	src := "SELECT 1; SELECT 2; SELECT 3"
	tg, ok := RunTarget(src, 12, 0, 0, SQLite)
	if !ok || tg.SQL != "SELECT 2" {
		t.Fatalf("caret: %+v %v", tg, ok)
	}
	// Selection wins over the caret statement, even when reversed.
	tg, ok = RunTarget(src, 12, 28, 10, SQLite)
	if !ok || tg.SQL != "SELECT 2; SELECT 3" {
		t.Fatalf("selection: %+v %v", tg, ok)
	}
	// A whitespace-only selection falls back to the caret.
	tg, _ = RunTarget(src, 1, 9, 10, SQLite)
	if tg.SQL != "SELECT 1" {
		t.Fatalf("blank selection: %+v", tg)
	}
	if _, ok := RunTarget("", 0, 0, 0, SQLite); ok {
		t.Fatal("empty buffer must not run")
	}
}

func TestLineCol(t *testing.T) {
	src := "ab\ncdé\nf"
	if l, c := LineCol(src, 0); l != 1 || c != 1 {
		t.Fatal(l, c)
	}
	if l, c := LineCol(src, len("ab\ncdé")); l != 2 || c != 4 {
		t.Fatal(l, c)
	}
	if got := OffsetOfLine(src, 3); got != len("ab\ncdé\n") {
		t.Fatal(got)
	}
	if got := OffsetOfLine(src, 99); got != len(src) {
		t.Fatal(got)
	}
}
