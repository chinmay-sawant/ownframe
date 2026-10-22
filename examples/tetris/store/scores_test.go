package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestSaveGameIsIdempotentByID(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()
	res := testResult("game-1", 1000)

	if err := s.SaveGame(ctx, res, nil); err != nil {
		t.Fatal(err)
	}

	res.Score = 9999

	if err := s.SaveGame(ctx, res, nil); err != nil {
		t.Fatal(err)
	}

	top, err := s.TopScores(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 1 {
		t.Fatalf("%d rows after a duplicate save", len(top))
	}

	if top[0].Score != 1000 {
		t.Fatalf("a duplicate save changed the score to %d", top[0].Score)
	}
}

func TestTieBreakByIDAscending(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	for _, id := range []string{"b", "a", "c"} {
		if err := s.SaveGame(ctx, testResult(id, 500), nil); err != nil {
			t.Fatal(err)
		}
	}

	top, err := s.TopScores(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	got := fmt.Sprintf("%s %s %s", top[0].ID, top[1].ID, top[2].ID)
	if got != "a b c" {
		t.Fatalf("tied order %q, want a b c", got)
	}
}

func TestLiveAndDummyRankingsStaySeparate(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	if err := s.SaveGame(ctx, testResult("live", game.MaxScore), nil); err != nil {
		t.Fatal(err)
	}

	live, err := s.TopScores(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(live) != 1 || live[0].ID != "live" {
		t.Fatalf("live ranking holds %v", live)
	}

	dummy, err := s.DummyScores(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(dummy) != len(dummyScores) {
		t.Fatalf("dummy ranking holds %d rows", len(dummy))
	}

	for _, r := range dummy {
		if !r.Dummy || r.ID == "live" {
			t.Fatalf("row %s leaked between rankings", r.ID)
		}
	}
}
