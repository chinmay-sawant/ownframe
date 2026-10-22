package scene

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestCellKindPriority(t *testing.T) {
	var f Frame
	f.Board[5][3] = game.Cell(game.PieceJ)

	if k, ghost := f.cellKind(5, 3); k != game.PieceJ || ghost {
		t.Fatalf("locked cell = %v ghost=%v", k, ghost)
	}

	f.Piece = game.PieceT
	f.Cells = [4]game.Point{{X: 3, Y: 5}, {X: 4, Y: 5}, {X: 5, Y: 5}, {X: 4, Y: 6}}
	f.Ghost = [4]game.Point{{X: 3, Y: 7}, {X: 4, Y: 7}, {X: 5, Y: 7}, {X: 4, Y: 8}}

	if k, ghost := f.cellKind(5, 4); k != game.PieceT || ghost {
		t.Fatalf("piece cell = %v ghost=%v", k, ghost)
	}

	if k, ghost := f.cellKind(7, 4); k != game.PieceT || !ghost {
		t.Fatalf("ghost cell = %v ghost=%v", k, ghost)
	}

	if k, ghost := f.cellKind(9, 9); k != 0 || ghost {
		t.Fatalf("empty cell = %v ghost=%v", k, ghost)
	}
}

func TestLetter(t *testing.T) {
	want := []byte{0, 'i', 'o', 't', 's', 'z', 'j', 'l'}

	for k := game.Piece(0); k <= game.PieceL+1; k++ {
		got := letter(k)
		idx := int(k)
		if idx < len(want) && got != want[idx] {
			t.Fatalf("letter(%d) = %q, want %q", k, got, want[idx])
		}

		if int(k) >= len(want) && got != 0 {
			t.Fatalf("letter(%d) = %q, want 0", k, got)
		}
	}
}

func TestNextSetShapes(t *testing.T) {
	for k := game.PieceI; k <= game.PieceL; k++ {
		set := nextSet(k)
		n := 0

		for r := 0; r < 4; r++ {
			for c := 0; c < 4; c++ {
				if set[r][c] {
					n++
				}
			}
		}

		if n != 4 {
			t.Fatalf("piece %v has %d preview cells", k, n)
		}
	}

	if got := nextSet(game.Piece(0)); got != ([4][4]bool{}) {
		t.Fatal("an empty piece has preview cells")
	}
}
