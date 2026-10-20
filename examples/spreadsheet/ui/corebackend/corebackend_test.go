package corebackend

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

func sheetID(t *testing.T, b *Backend, name string) string {
	t.Helper()

	for _, sh := range b.Sheets() {
		if sh.Name == name {
			return sh.ID
		}
	}

	t.Fatalf("sheet %q missing", name)

	return ""
}

func TestSeedAndRange(t *testing.T) {
	b := newBackend(t, t.TempDir())

	sheets := b.Sheets()
	if len(sheets) != 3 || sheets[0].Name != "Numbers" || sheets[2].Name != "Mixed" {
		t.Fatalf("sheets = %+v", sheets)
	}

	id := sheetID(t, b, "Numbers")
	cells, err := b.Range(id, ui.Area{R0: 0, C0: 0, R1: 0, C1: 3})
	if err != nil {
		t.Fatal(err)
	}

	if len(cells) != 4 {
		t.Fatalf("cells = %d", len(cells))
	}

	if cells[0].Raw != "row 1" || cells[1].Raw != "0" || !cells[1].Num {
		t.Fatalf("row 1 = %+v", cells)
	}

	if cells[2].Raw != "=B1*2" || cells[2].Display != "0" {
		t.Fatalf("C1 = %+v", cells[2])
	}

	mixed := sheetID(t, b, "Mixed")
	cells, err = b.Range(mixed, ui.Area{R0: 5, C0: 2, R1: 5, C1: 2})
	if err != nil {
		t.Fatal(err)
	}

	if cells[0].Err == "" {
		t.Fatalf("division error = %+v", cells[0])
	}
}

func TestPrefsAndUsed(t *testing.T) {
	b := newBackend(t, t.TempDir())
	id := sheetID(t, b, "Numbers")

	if err := b.SetPref("theme", "dark"); err != nil {
		t.Fatal(err)
	}

	if v, ok := b.Pref("theme"); !ok || v != "dark" {
		t.Fatalf("pref = %q ok=%v", v, ok)
	}

	area, ok := b.Used(id)
	if !ok || area.R1 < 199 || area.C1 < 19 {
		t.Fatalf("used = %+v ok=%v", area, ok)
	}
}
