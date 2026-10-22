package workbook

import (
	"errors"
	"testing"
)

func TestCSVTypes(t *testing.T) {
	tab, err := ParseCSV([]byte("0,007,=1+1,text,0x10\n"), DefaultCSVOptions())
	if err != nil {
		t.Fatal(err)
	}

	row := tab.Cells[0]

	if row[0].Kind != Number || row[0].Number != 0 {
		t.Errorf("0 = %+v", row[0])
	}

	if row[1].Kind != Number || row[1].Number != 7 {
		t.Errorf("007 = %+v", row[1])
	}

	if row[2].Kind != Formula || row[2].Source != "1+1" {
		t.Errorf("=1+1 = %+v", row[2])
	}

	if row[3].Kind != Text {
		t.Errorf("text = %+v", row[3])
	}

	if row[4].Kind != Text || row[4].Text != "0x10" {
		t.Errorf("0x10 = %+v, want text", row[4])
	}

	opt := DefaultCSVOptions()
	opt.InterpretNumbers = false
	opt.InterpretFormulas = false

	tab, err = ParseCSV([]byte("12,=1+1\n"), opt)
	if err != nil {
		t.Fatal(err)
	}

	if tab.Cells[0][0].Kind != Text || tab.Cells[0][1].Kind != Text {
		t.Fatalf("plain mode = %+v %+v", tab.Cells[0][0], tab.Cells[0][1])
	}
}

func TestCSVTooLarge(t *testing.T) {
	opt := DefaultCSVOptions()
	opt.MaxCells = 2

	if _, err := ParseCSV([]byte("1,2,3\n"), opt); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
}
