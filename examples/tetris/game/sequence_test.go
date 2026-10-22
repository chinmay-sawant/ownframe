package game

import (
	"slices"
	"testing"
)

func TestSequenceIsDeterministicAndBagged(t *testing.T) {
	a := Sequence(7, 14)
	b := Sequence(7, 14)

	if !slices.Equal(a, b) {
		t.Fatalf("the same seed gave %v then %v", a, b)
	}

	if len(a) != 14 {
		t.Fatalf("sequence length %d, want 14", len(a))
	}

	for bag := 0; bag < 2; bag++ {
		seen := map[Piece]bool{}
		for i := 0; i < PieceCount; i++ {
			p := a[bag*PieceCount+i]
			if !p.Valid() || seen[p] {
				t.Fatalf("bag %d is not a permutation: %v", bag, a[bag*PieceCount:(bag+1)*PieceCount])
			}

			seen[p] = true
		}
	}

	if slices.Equal(Sequence(1, 14), Sequence(2, 14)) {
		t.Fatal("different seeds produced the same pieces")
	}
}

func TestSequenceWithNoCount(t *testing.T) {
	if got := Sequence(7, -1); len(got) != 0 {
		t.Fatalf("negative count returned %v", got)
	}
}
