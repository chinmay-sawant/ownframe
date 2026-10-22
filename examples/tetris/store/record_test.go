package store

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestInvalidResultsAreRejected(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	bad := []game.Result{
		{},
		{ID: "x", Level: 0},
		{ID: "x", Level: 1, Score: game.MaxScore + 1},
		{ID: "x", Level: 1, DurationMS: -1},
		{ID: "x", Level: 1, Lines: game.MaxLines + 1},
	}

	for i, r := range bad {
		if err := s.SaveGame(ctx, r, nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("bad result %d returned %v", i, err)
		}
	}
}

func TestSaveGameWithReplayRoundTrips(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()
	rep := &game.Replay{
		Seed:           5,
		Ruleset:        game.Ruleset,
		FixtureVersion: game.FixtureVersion,
		Events:         []game.InputEvent{{Step: 1, Action: game.ActionLeft}},
	}

	if err := s.SaveGame(ctx, testResult("with-replay", 10), rep); err != nil {
		t.Fatal(err)
	}

	got, err := s.Replay(ctx, "with-replay")
	if err != nil {
		t.Fatal(err)
	}

	if got.Seed != 5 || len(got.Events) != 1 || got.Events[0].Action != game.ActionLeft {
		t.Fatalf("replay round trip returned %+v", got)
	}

	if _, err := s.Replay(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing replay returned %v", err)
	}
}
