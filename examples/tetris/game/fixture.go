package game

import "fmt"

// Fixture is a named starting position for tests and demos. Piece zero
// means the game spawns its own first piece.
type Fixture struct {
	Name  string
	Board Board
	Piece Piece
	Rot   Rotation
	X, Y  int
	Note  string
}

// Fixtures returns the selectable board fixtures in UI order. Each one
// starts running, so a test can Step or Apply right away.
func Fixtures() []Fixture {
	return []Fixture{
		{Name: "empty", Note: "valid empty board, default play"},
		{
			Name: "near-top-out", Board: nearTopOutBoard(), Piece: PieceT,
			X: fixtureTopX, Y: 0, Note: "stack two rows from the top",
		},
		{
			Name: "clear-1", Board: wellBoard(1), Piece: PieceI, Rot: RotR,
			X: -2, Y: fixtureWellY, Note: "vertical I clears one row",
		},
		{
			Name: "clear-2", Board: wellBoard(wellDepth2), Piece: PieceI, Rot: RotR,
			X: -2, Y: fixtureWellY, Note: "vertical I clears two rows",
		},
		{
			Name: "clear-3", Board: wellBoard(wellDepth3), Piece: PieceI, Rot: RotR,
			X: -2, Y: fixtureWellY, Note: "vertical I clears three rows",
		},
		{
			Name: "clear-4", Board: wellBoard(wellDepth4), Piece: PieceI, Rot: RotR,
			X: -2, Y: fixtureWellY, Note: "vertical I clears four rows",
		},
	}
}

// FixtureNames lists the fixture names in UI order.
func FixtureNames() []string {
	fs := Fixtures()
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Name
	}

	return out
}

// NewFromFixture starts a running game on a named fixture.
func NewFromFixture(name string, seed uint64) (*Game, error) {
	for _, f := range Fixtures() {
		if f.Name != name {
			continue
		}

		g := New(seed)
		g.Phase = PhaseRunning
		g.Board = f.Board

		if f.Piece.Valid() {
			g.Piece = f.Piece
			g.Rot = f.Rot % rotationCount
			g.X, g.Y = f.X, f.Y
			g.fillPreview()
		} else {
			g.spawn()
		}

		return g, nil
	}

	return nil, fmt.Errorf("game: unknown fixture %q", name)
}
