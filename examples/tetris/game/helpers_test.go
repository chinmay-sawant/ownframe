package game

import (
	"sort"
	"testing"
)

// running returns a running game holding a chosen piece.
func running(p Piece, r Rotation, x, y int) *Game {
	g := New(1)
	g.Phase = PhaseRunning
	g.Piece = p
	g.Rot = r
	g.X, g.Y = x, y
	g.fillPreview()

	return g
}

// solidBoard fills rows from start to the bottom.
func solidBoard(start int, p Piece) Board {
	var b Board
	for y := start; y < Rows; y++ {
		for x := 0; x < Cols; x++ {
			b[y][x] = Cell(p)
		}
	}

	return b
}

// cellsEqual compares two cell lists as sets.
func cellsEqual(got []Point, want ...Point) bool {
	if len(got) != len(want) {
		return false
	}

	key := func(p Point) int { return p.Y*1000 + p.X }

	gk := make([]int, len(got))
	for i, p := range got {
		gk[i] = key(p)
	}

	wk := make([]int, len(want))
	for i, p := range want {
		wk[i] = key(p)
	}

	sort.Ints(gk)
	sort.Ints(wk)

	for i := range gk {
		if gk[i] != wk[i] {
			return false
		}
	}

	return true
}

// expectPhase fails the test unless the game is in want.
func expectPhase(t *testing.T, g *Game, want Phase) {
	t.Helper()

	if g.Phase != want {
		t.Fatalf("phase %v, want %v", g.Phase, want)
	}
}

// haveEvent reports whether ev holds a kind.
func haveEvent(ev []Event, kind EventKind) bool {
	for _, e := range ev {
		if e.Kind == kind {
			return true
		}
	}

	return false
}

// clearEvent returns the first clear event in ev.
func clearEvent(ev []Event) (Event, bool) {
	for _, e := range ev {
		if e.Kind == EventClear {
			return e, true
		}
	}

	return Event{}, false
}
