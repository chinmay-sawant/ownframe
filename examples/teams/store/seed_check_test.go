package store

import (
	"errors"
	"testing"
)

func TestSingleStatement(t *testing.T) {
	stmt, ok := singleStatement("INSERT INTO meta (key, value) VALUES ('a', 'b');")
	if !ok || stmt != "INSERT INTO meta (key, value) VALUES ('a', 'b')" {
		t.Fatalf("stmt = %q %v", stmt, ok)
	}
	if _, ok := singleStatement("SELECT 1; SELECT 2;"); ok {
		t.Fatal("stacked statements accepted")
	}
	if _, ok := singleStatement(";"); ok {
		t.Fatal("empty statement accepted")
	}
}

func TestExecScriptRejectsStacked(t *testing.T) {
	st, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	tx, err := st.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := execScript(tx, "INSERT INTO meta (key, value) VALUES ('a', 'b');"); err != nil {
		t.Fatal(err)
	}
	err = execScript(tx, "INSERT INTO meta (key, value) VALUES ('c', 'd'); DELETE FROM meta;")
	if !errors.Is(err, errMultiStatement) {
		t.Fatalf("err = %v", err)
	}
}
