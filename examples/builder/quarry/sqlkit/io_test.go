package sqlkit

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

var exportCols = []string{"id", "name", "note", "score", "ok", "blob"}
var exportRows = [][]any{
	{int64(1), "Ada, \"Countess\"", "line1\nline2", 9.5, true, []byte("hi")},
	{int64(2), "Bob", nil, 3.0, false, []byte{0xff, 0x00}},
	{int64(3), "O'Neil", "", 0.1, nil, nil},
}

func TestWriteCSVRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, exportCols, SliceRows(exportRows), ExportOptions{Null: "NULL"}); err != nil {
		t.Fatal(err)
	}
	recs, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 4 || !reflect.DeepEqual(recs[0], exportCols) {
		t.Fatalf("%q", recs)
	}
	if recs[1][1] != `Ada, "Countess"` || recs[1][2] != "line1\nline2" || recs[1][3] != "9.5" || recs[1][4] != "true" {
		t.Fatalf("row 1 %q", recs[1])
	}
	if recs[2][2] != "NULL" || recs[2][5] != `\xff00` || recs[3][2] != "" || recs[3][4] != "NULL" {
		t.Fatalf("null/blob handling %q %q", recs[2], recs[3])
	}
	// Options: tab delimiter and no header.
	buf.Reset()
	_ = WriteCSV(&buf, []string{"a", "b"}, SliceRows([][]any{{1, "x"}}), ExportOptions{Delimiter: '\t', NoHeader: true})
	if buf.String() != "1\tx\n" {
		t.Fatalf("%q", buf.String())
	}
}

func TestWriteCSVThenParseCSV(t *testing.T) {
	var buf bytes.Buffer
	rows := [][]any{{int64(1), "a", 1.5, true}, {int64(2), "b,c", 2.0, false}}
	_ = WriteCSV(&buf, []string{"i", "s", "f", "b"}, SliceRows(rows), ExportOptions{})
	d, err := ParseCSV(&buf, ImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	types := []ColumnType{TypeInteger, TypeText, TypeReal, TypeBoolean}
	for i, c := range d.Columns {
		if c.Type != types[i] {
			t.Fatalf("col %s: %v", c.Name, c.Type)
		}
	}
	if !reflect.DeepEqual(d.Rows[1], []any{int64(2), "b,c", 2.0, false}) {
		t.Fatalf("%#v", d.Rows[1])
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	rows := append([][]any{}, exportRows...)
	rows = append(rows, []any{int64(4), ts, nil, 1.0, true, []byte("x")})
	if err := WriteJSON(&buf, exportCols, SliceRows(rows)); err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(got) != 4 || got[0]["name"] != `Ada, "Countess"` || got[1]["note"] != nil || got[1]["blob"] != `\xff00` || got[3]["name"] != "2026-01-02T03:04:05Z" {
		t.Fatalf("%v", got)
	}
	if !strings.HasPrefix(buf.String(), "[\n  {\n    \"id\": 1,\n    \"name\"") {
		t.Fatalf("not pretty / column order lost:\n%s", buf.String())
	}
	buf.Reset()
	_ = WriteJSON(&buf, exportCols, SliceRows(nil))
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("empty: %q", buf.String())
	}
}

func TestWriteNDJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteNDJSON(&buf, exportCols, SliceRows(exportRows)); err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(&buf)
	n := 0
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			t.Fatalf("line %d: %v: %s", n, err, sc.Text())
		}
		n++
	}
	if n != 3 {
		t.Fatalf("lines %d", n)
	}
}

func TestLiteral(t *testing.T) {
	tests := []struct {
		v    any
		d    Dialect
		want string
	}{
		{nil, SQLite, "NULL"},
		{true, SQLite, "1"}, {false, SQLite, "0"},
		{true, Postgres, "TRUE"},
		{42, SQLite, "42"}, {int64(-7), Postgres, "-7"},
		{3.0, SQLite, "3.0"}, {1.5e300, SQLite, "1.5e+300"},
		{"it's", SQLite, "'it''s'"},
		{[]byte{0xde, 0xad}, SQLite, "X'dead'"},
		{[]byte{0xde, 0xad}, Postgres, `'\xdead'`},
	}
	for _, tc := range tests {
		if got := Literal(tc.v, tc.d); got != tc.want {
			t.Errorf("Literal(%v, %v) = %s want %s", tc.v, tc.d, got, tc.want)
		}
	}
}

func TestWriteSQL(t *testing.T) {
	var buf bytes.Buffer
	err := WriteSQL(&buf, []string{"id", "name"}, SliceRows([][]any{{1, "a'b"}, {2, nil}, {3, "c"}}), ExportOptions{Table: "t x", Dialect: SQLite, BatchSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := "INSERT INTO \"t x\" (\"id\", \"name\") VALUES\n  (1, 'a''b'),\n  (2, NULL);\nINSERT INTO \"t x\" (\"id\", \"name\") VALUES\n  (3, 'c');\n"
	if buf.String() != want {
		t.Fatalf("got\n%s\nwant\n%s", buf.String(), want)
	}
	if err := WriteSQL(&buf, nil, SliceRows(nil), ExportOptions{}); err == nil {
		t.Fatal("missing table accepted")
	}
	buf.Reset()
	_ = WriteSQL(&buf, []string{"a"}, SliceRows([][]any{{1}}), ExportOptions{Table: "t", Schema: "s", Dialect: Postgres})
	if buf.String() != "INSERT INTO \"s\".\"t\" (\"a\") VALUES\n  (1);\n" {
		t.Fatal(buf.String())
	}
	buf.Reset()
	_ = WriteSQL(&buf, []string{"a"}, SliceRows(nil), ExportOptions{Table: "t"})
	if buf.Len() != 0 {
		t.Fatal("empty result must emit nothing")
	}
}

func TestParseCSVInference(t *testing.T) {
	src := "\xef\xbb\xbfid,price,active,name,empty,mixed,bm\n1,9.99,true,Ann,,1,true\n2,10,FALSE,\"Bo, B\",,true,1\n3,,true,Cy,,x,false\n"
	d, err := ParseCSV(strings.NewReader(src), ImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []ImportColumn{
		{"id", TypeInteger}, {"price", TypeReal}, {"active", TypeBoolean}, {"name", TypeText},
		{"empty", TypeText}, {"mixed", TypeText}, {"bm", TypeText},
	}
	if !reflect.DeepEqual(d.Columns, want) {
		t.Fatalf("got %v want %v", d.Columns, want)
	}
	if len(d.Rows) != 3 {
		t.Fatal(len(d.Rows))
	}
	if !reflect.DeepEqual(d.Rows[1][:4], []any{int64(2), 10.0, false, "Bo, B"}) || d.Rows[2][1] != nil || d.Rows[0][4] != nil {
		t.Fatalf("%#v", d.Rows)
	}
	if len(d.Preview(2)) != 2 || len(d.Preview(50)) != 3 || len(d.Preview(-1)) != 0 {
		t.Fatal("preview bounds")
	}
	if got := d.CreateTableSQL("imp", SQLite); got != `CREATE TABLE "imp" ("id" INTEGER, "price" REAL, "active" BOOLEAN, "name" TEXT, "empty" TEXT, "mixed" TEXT, "bm" TEXT)` {
		t.Fatal(got)
	}
	if got := d.CreateTableSQL("imp", Postgres); !strings.Contains(got, `"id" BIGINT, "price" DOUBLE PRECISION`) {
		t.Fatal(got)
	}
	if got := d.InsertSQL("imp", Postgres); got != `INSERT INTO "imp" ("id", "price", "active", "name", "empty", "mixed", "bm") VALUES ($1, $2, $3, $4, $5, $6, $7)` {
		t.Fatal(got)
	}
	if got := d.InsertSQL("imp", SQLite); !strings.HasSuffix(got, "VALUES (?, ?, ?, ?, ?, ?, ?)") {
		t.Fatal(got)
	}
}

func TestParseCSVHeadersAndShape(t *testing.T) {
	d, err := ParseCSV(strings.NewReader("a,A,,b\n1,2,3\n4,5,6,7,8\n"), ImportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, c := range d.Columns {
		names = append(names, c.Name)
	}
	if !reflect.DeepEqual(names, []string{"a", "A_2", "column_3", "b", "column_5"}) {
		t.Fatalf("%v", names)
	}
	// Ragged rows are padded with NULL to the widest record.
	if len(d.Rows[0]) != len(d.Columns) || d.Rows[0][len(d.Columns)-1] != nil {
		t.Fatalf("%#v", d.Rows[0])
	}
	d, err = ParseCSV(strings.NewReader("1;x\n2;y\n"), ImportOptions{Delimiter: ';', NoHeader: true})
	if err != nil || len(d.Rows) != 2 || d.Columns[0].Name != "column_1" || d.Columns[0].Type != TypeInteger {
		t.Fatalf("%v %+v", err, d)
	}
	if _, err := ParseCSV(strings.NewReader(""), ImportOptions{}); err == nil {
		t.Fatal("empty file accepted")
	}
	d, _ = ParseCSV(strings.NewReader("a\nNA\n3\n"), ImportOptions{Null: "NA"})
	if d.Columns[0].Type != TypeInteger || d.Rows[0][0] != nil {
		t.Fatalf("%+v", d)
	}
	d, _ = ParseCSV(strings.NewReader("a\n0x10\n1e5\ninf\n"), ImportOptions{})
	if d.Columns[0].Type != TypeText {
		t.Fatalf("hex/inf must stay text: %v", d.Columns[0].Type)
	}
}
