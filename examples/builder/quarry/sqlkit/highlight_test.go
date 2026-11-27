package sqlkit

import (
	"strings"
	"testing"
)

// kinds renders non-space tokens as "kind:text" for compact assertions.
func kinds(src string, d Dialect) []string {
	var out []string
	for _, t := range Tokenize(src, d) {
		if t.Kind == KindSpace {
			continue
		}
		out = append(out, t.Kind.String()+":"+t.Text(src))
	}
	return out
}

func TestTokenizeTable(t *testing.T) {
	tests := []struct {
		name string
		src  string
		d    Dialect
		want []string
	}{
		{"keywords and ident", "select id from users", SQLite,
			[]string{"keyword:select", "ident:id", "keyword:from", "ident:users"}},
		{"function needs paren", "SELECT count(*), count FROM t", SQLite,
			[]string{"keyword:SELECT", "function:count", "operator:(", "operator:*", "operator:)", "operator:,", "ident:count", "keyword:FROM", "ident:t"}},
		{"escaped single quote", "SELECT 'it''s' x", SQLite,
			[]string{"keyword:SELECT", "string:'it''s'", "ident:x"}},
		{"double quoted ident with escape", `SELECT "a""b" FROM t`, SQLite,
			[]string{"keyword:SELECT", `quoted-ident:"a""b"`, "keyword:FROM", "ident:t"}},
		{"numbers", "1 2.5 .5 1e-3 0xFF", SQLite,
			[]string{"number:1", "number:2.5", "number:.5", "number:1e-3", "number:0xFF"}},
		{"line comment", "SELECT 1 -- hi; there\nFROM t", SQLite,
			[]string{"keyword:SELECT", "number:1", "comment:-- hi; there", "keyword:FROM", "ident:t"}},
		{"sqlite comment does not nest", "/* a /* b */ c */", SQLite,
			[]string{"comment:/* a /* b */", "ident:c", "operator:*", "operator:/"}},
		{"postgres nested comment", "/* a /* b */ c */ x", Postgres,
			[]string{"comment:/* a /* b */ c */", "ident:x"}},
		{"unterminated block comment", "SELECT /* oops", SQLite,
			[]string{"keyword:SELECT", "comment:/* oops"}},
		{"dollar quoted", "SELECT $$a'b;c$$, $fn$x$$y$fn$", Postgres,
			[]string{"keyword:SELECT", "string:$$a'b;c$$", "operator:,", "string:$fn$x$$y$fn$"}},
		{"unterminated dollar", "SELECT $$abc", Postgres,
			[]string{"keyword:SELECT", "string:$$abc"}},
		{"positional param vs dollar string", "WHERE a = $1 AND b = $2", Postgres,
			[]string{"keyword:WHERE", "ident:a", "operator:=", "param:$1", "keyword:AND", "ident:b", "operator:=", "param:$2"}},
		{"cast to type", "SELECT x::jsonb, y::text", Postgres,
			[]string{"keyword:SELECT", "ident:x", "operator:::", "type:jsonb", "operator:,", "ident:y", "operator:::", "type:text"}},
		{"escape string backslash", `SELECT E'it\'s' z`, Postgres,
			[]string{"keyword:SELECT", "ident:E", `string:'it\'s'`, "ident:z"}},
		{"sqlite params", "? ?2 :name @id", SQLite,
			[]string{"param:?", "param:?2", "param::name", "param:@id"}},
		{"types", "CREATE TABLE t (a INTEGER, b VARCHAR(10), c JSONB)", Postgres,
			[]string{"keyword:CREATE", "keyword:TABLE", "ident:t", "operator:(", "ident:a", "type:INTEGER", "operator:,", "ident:b", "type:VARCHAR", "operator:(", "number:10", "operator:)", "operator:,", "ident:c", "type:JSONB", "operator:)"}},
		{"json operators", "a->>'k'", Postgres,
			[]string{"ident:a", "operator:->>", "string:'k'"}},
		{"sqlite bracket ident", "SELECT [my col] FROM t", SQLite,
			[]string{"keyword:SELECT", "quoted-ident:[my col]", "keyword:FROM", "ident:t"}},
		{"after dot is ident", "t.select", SQLite,
			[]string{"ident:t", "operator:.", "ident:select"}},
		{"sqlite date function", "date('now')", SQLite,
			[]string{"function:date", "operator:(", "string:'now'", "operator:)"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := kinds(tc.src, tc.d)
			if strings.Join(got, "|") != strings.Join(tc.want, "|") {
				t.Fatalf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestTokenizeCoversEveryByte(t *testing.T) {
	srcs := []string{"", "SELECT 'x", "héllo wörld -- ü", "a;;b", "$$", "[", "\"", "/*", "日本語 = 1", "x\x00y"}
	for _, d := range []Dialect{SQLite, Postgres} {
		for _, src := range srcs {
			pos := 0
			for _, tk := range Tokenize(src, d) {
				if tk.Start != pos || tk.End <= tk.Start {
					t.Fatalf("%q: gap or empty token %v at %d", src, tk, pos)
				}
				pos = tk.End
			}
			if pos != len(src) {
				t.Fatalf("%q: covered %d of %d bytes", src, pos, len(src))
			}
		}
	}
}

func TestDialectKeywords(t *testing.T) {
	if k := kinds("PRAGMA x", SQLite); k[0] != "keyword:PRAGMA" {
		t.Fatalf("sqlite: %v", k)
	}
	if k := kinds("PRAGMA x", Postgres); k[0] != "ident:PRAGMA" {
		t.Fatalf("postgres: %v", k)
	}
	if k := kinds("ILIKE", Postgres); k[0] != "keyword:ILIKE" {
		t.Fatalf("pg ilike: %v", k)
	}
}

func BenchmarkTokenize(b *testing.B) {
	src := strings.Repeat("SELECT o.id, count(*) FROM orders o -- c\nWHERE total > 100 AND name = 'x''y';\n", 200)
	b.SetBytes(int64(len(src)))
	for b.Loop() {
		Tokenize(src, Postgres)
	}
}
