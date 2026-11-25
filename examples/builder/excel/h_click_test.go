package main

import (
	"context"
	"testing"
	"time"
)

// TestSingleClickSelectsWithoutEditing: a lone press and release on a cell,
// filled or blank, only selects it.
func TestSingleClickSelectsWithoutEditing(t *testing.T) {
	for _, c := range []CellRef{{Row: 2, Col: 2}, {Row: 30, Col: 10}} {
		a, _ := newTestApp(t)
		x, y := cellPoint(t, a, c)
		pointer(t, a, x, y)
		if a.editing {
			t.Fatalf("single click on %v entered edit mode", c)
		}
		if a.active != c {
			t.Fatalf("active %v, want %v", a.active, c)
		}
	}
}

// TestWindowTapEchoDoesNotDoubleClick replays what the desktop window does
// for one mouse click: a claimed press and release, then the tap echo that
// presses the same point again at once. That echo must not read as a double
// click.
func TestWindowTapEchoDoesNotDoubleClick(t *testing.T) {
	for _, c := range []CellRef{{Row: 2, Col: 2}, {Row: 30, Col: 10}} {
		a, _ := newTestApp(t)
		a.tapEcho = tapEchoGap
		x, y := cellPoint(t, a, c)
		pointer(t, a, x, y)
		pointer(t, a, x, y) // the echo
		if a.editing {
			t.Fatalf("click on %v entered edit mode through the tap echo", c)
		}
	}
}

// TestSlowSecondClickStillEdits keeps a real double click working.
func TestSlowSecondClickStillEdits(t *testing.T) {
	a, _ := newTestApp(t)
	a.tapEcho = tapEchoGap
	c := CellRef{Row: 2, Col: 2}
	x, y := cellPoint(t, a, c)
	pointer(t, a, x, y)
	time.Sleep(2 * tapEchoGap)
	pointer(t, a, x, y)
	if !a.editing {
		t.Fatal("double click did not enter edit mode")
	}
}

// TestBlankCellClickKeepsCellSize: after a click on a blank cell the active
// outline is the size of the cell, not larger, and nothing is in edit mode.
func TestBlankCellClickKeepsCellSize(t *testing.T) {
	for _, pct := range []int{100, 150} {
		a, _ := newTestApp(t)
		a.setZoom(pct)
		a.tapEcho = tapEchoGap
		c := CellRef{Row: 12, Col: 6}
		x, y := cellPoint(t, a, c)
		pointer(t, a, x, y)
		if err := a.page.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		if a.editing {
			t.Fatalf("zoom %d: blank cell click entered edit mode", pct)
		}
		cw, ch := a.canvasSize()
		col, _ := slotFor(a.layoutCols(cw), c.Col)
		row, _ := slotFor(a.layoutRows(ch), c.Row)
		act := mustBox(t, a, "xl-gov-active")
		t.Logf("zoom %d: active outline %vx%v, cell %dx%d", pct, act.W, act.H, col.size, row.size)
		if int(act.W) > col.size+2 || int(act.H) > row.size+2 {
			t.Fatalf("zoom %d: active outline %vx%v is larger than the %dx%d cell", pct, act.W, act.H, col.size, row.size)
		}
	}
}
