package driver

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestCellValueFormatting(t *testing.T) {
	ts := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	cases := []struct {
		name string
		v    CellValue
		want string
		lit  string
	}{
		{"null", Null(), "NULL", "NULL"},
		{"int", Int(-42), "-42", "-42"},
		{"float", Float(1.5), "1.5", "1.5"},
		{"text", Text("it's"), "it's", "'it''s'"},
		{"bool", Bool(true), "true", "TRUE"},
		{"time", Time(ts), "2025-03-04 05:06:07", "'2025-03-04 05:06:07'"},
		{"json", JSON(`{"a":1}`), `{"a":1}`, `'{"a":1}'`},
		{"blob", Blob([]byte{0xde, 0xad, 0xbe, 0xef}), "<4 bytes> 0xdeadbeef", "X'deadbeef'"},
		{"longblob", Blob(make([]byte, 20)), "<20 bytes> 0x0000000000000000...", "X'" + strings.Repeat("00", 20) + "'"},
	}
	for _, tc := range cases {
		if got := tc.v.String(); got != tc.want {
			t.Errorf("%s String = %q, want %q", tc.name, got, tc.want)
		}
		if got := tc.v.SQLLiteral(); got != tc.lit {
			t.Errorf("%s SQLLiteral = %q, want %q", tc.name, got, tc.lit)
		}
		same := tc.v
		if !tc.v.Equal(same) {
			t.Errorf("%s not equal to itself", tc.name)
		}
	}
	if Int(1).Equal(Float(1)) {
		t.Error("different kinds must not be equal")
	}
	if !Null().IsNull() || Int(0).IsNull() {
		t.Error("IsNull misreports")
	}
}

func TestFromAny(t *testing.T) {
	if FromAny(nil, "").Kind != KindNull {
		t.Error("nil should be Null")
	}
	if c := FromAny(int64(1), "BOOLEAN"); c.Kind != KindBool || !c.B {
		t.Errorf("BOOLEAN int = %+v", c)
	}
	if c := FromAny([]byte(`{"x":1}`), "JSON"); c.Kind != KindJSON {
		t.Errorf("json bytes = %+v", c)
	}
	if c := FromAny([]byte("abc"), "TEXT"); c.Kind != KindText {
		t.Errorf("text bytes = %+v", c)
	}
	if c := FromAny([]byte{1, 2}, ""); c.Kind != KindBlob {
		t.Errorf("blob bytes = %+v", c)
	}
}

func TestReadOnlyStatementBlocking(t *testing.T) {
	blocked := []string{
		"INSERT INTO t VALUES (1)", "update t set a=1", "DELETE FROM t", "DROP TABLE t",
		"ALTER TABLE t ADD c int", "TRUNCATE t", "CREATE TABLE x(a)", "  /* c */ insert into t values (1)",
		"SELECT 1; DELETE FROM t", "WITH x AS (DELETE FROM t RETURNING *) SELECT * FROM x",
		"SELECT * INTO copy FROM t", "PRAGMA journal_mode = wal", "EXPLAIN ANALYZE UPDATE t SET a = 1",
		"VACUUM", "ATTACH DATABASE 'x' AS y",
	}
	for _, s := range blocked {
		if err := CheckReadOnly(s); !errors.Is(err, ErrReadOnly) {
			t.Errorf("CheckReadOnly(%q) = %v, want ErrReadOnly", s, err)
		}
	}
	allowed := []string{
		"SELECT * FROM t", "select 'DROP TABLE t'", "SELECT \"delete\" FROM t -- insert here",
		"WITH x AS (SELECT 1) SELECT * FROM x", "EXPLAIN QUERY PLAN SELECT 1", "PRAGMA table_info(t)",
		"VALUES (1), (2)", "SHOW server_version", "SELECT 1;", "", "  ;  ", "SELECT * FROM t FOR UPDATE",
	}
	for _, s := range allowed {
		if err := CheckReadOnly(s); err != nil {
			t.Errorf("CheckReadOnly(%q) = %v, want nil", s, err)
		}
	}
	if IsReadStatement("") || !IsReadStatement("SELECT 1") || IsReadStatement("DROP TABLE t") {
		t.Error("IsReadStatement misclassifies")
	}
}

func TestQuoting(t *testing.T) {
	if got := QuoteIdent(`we"ird`); got != `"we""ird"` {
		t.Errorf("QuoteIdent = %s", got)
	}
	if got := QualifiedTable("", "t"); got != `"t"` {
		t.Errorf("QualifiedTable = %s", got)
	}
	if got := QualifiedTable("s", "t"); got != `"s"."t"` {
		t.Errorf("QualifiedTable = %s", got)
	}
}
