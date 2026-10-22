package store

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestSnapshotRoundTripThroughTheStore(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	if _, ok, err := s.LoadSnapshot(ctx); err != nil || ok {
		t.Fatalf("empty slot returned ok=%v err=%v", ok, err)
	}

	g := game.New(5)
	g.Start()
	g.Apply(game.ActionSoftDrop)

	snap := g.Snapshot()

	if err := s.SaveSnapshot(ctx, snap); err != nil {
		t.Fatal(err)
	}

	got, ok, err := s.LoadSnapshot(ctx)
	if err != nil || !ok {
		t.Fatalf("load returned ok=%v err=%v", ok, err)
	}

	if got.ID != snap.ID || got.Score != snap.Score ||
		fmt.Sprint(got.Board) != fmt.Sprint(snap.Board) {
		t.Fatal("the snapshot changed in the store")
	}

	if err := s.ClearSnapshot(ctx); err != nil {
		t.Fatal(err)
	}

	if _, ok, err := s.LoadSnapshot(ctx); err != nil || ok {
		t.Fatalf("cleared slot returned ok=%v err=%v", ok, err)
	}
}

func TestInvalidSnapshotIsRejectedByTheStore(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if err := s.SaveSnapshot(context.Background(), game.Snapshot{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty snapshot returned %v", err)
	}
}
