package ui

import (
	"strings"
	"testing"
)

func TestRunQueryShowsGrid(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	t0 := r.a.activeTabPtr()
	t0.ed = newEditor("SELECT id, full_name, email, created_at FROM users ORDER BY id")
	r.chord("enter", false)
	r.settle()
	if t0.grid == nil || len(t0.grid.rows) == 0 {
		t.Fatalf("no rows: err=%+v", t0.err)
	}
	r.png("grid")
	if !strings.Contains(r.texts(), "full_name") {
		t.Fatal("header not painted")
	}
}
