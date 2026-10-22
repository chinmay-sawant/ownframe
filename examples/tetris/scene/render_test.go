package scene

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestPageReplaysAndMinSize(t *testing.T) {
	s, _ := newTestScene(t, &fakeModel{}, nil, Options{})

	if s.page.Display() == nil {
		t.Fatal("the game page fell back to the bitmap path")
	}

	w, h := s.page.MinSize()
	if w != MinWidth || h != MinHeight {
		t.Fatalf("min size = %d x %d", w, h)
	}

	bw, bh := s.page.Size()
	if bw != Width || bh != Height {
		t.Fatalf("size = %d x %d", bw, bh)
	}
}

func TestTickPaintsScoreBoardAndPreview(t *testing.T) {
	m := &fakeModel{}
	s, clock := newTestScene(t, m, nil, Options{Stepper: &fakeStepper{}})

	m.f.Score = 42
	m.f.Board[19][0] = game.Cell(game.PieceT)
	m.f.Piece = game.PieceI
	m.f.Cells = [4]game.Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}}
	m.f.Ghost = [4]game.Point{{X: 0, Y: 18}, {X: 1, Y: 18}, {X: 2, Y: 18}, {X: 3, Y: 18}}
	m.f.Next = game.PieceO

	tickAt(t, s, clock, time.Second/60)

	d := s.page.Display()
	boxes := s.page.Boxes()

	score := textAt(d, boxes, "t-score")
	if score == nil || score.Text != padScore(42) {
		t.Fatalf("score text = %v", score)
	}

	status := textAt(d, boxes, "t-status")
	if status == nil || status.Text != s.status {
		t.Fatalf("status text = %v, want %q", status, s.status)
	}

	r, g, b := paletteFor(false).kinds[game.PieceT].floats()
	locked := fillAt(d, boxes, "b19-0")
	if locked == nil || locked.R != r || locked.G != g || locked.B != b {
		t.Fatalf("locked cell = %v", locked)
	}

	ghost := fillAt(d, boxes, "b18-0")
	if ghost == nil || ghost.Alpha != paletteFor(false).ghostAlpha {
		t.Fatalf("ghost cell = %v", ghost)
	}

	pr, pg, pb := paletteFor(false).kinds[game.PieceO].floats()
	preview := fillAt(d, boxes, "n0-1")
	if preview == nil || preview.R != pr || preview.G != pg || preview.B != pb {
		t.Fatalf("preview cell = %v", preview)
	}
}
