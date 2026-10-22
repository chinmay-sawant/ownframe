package store

import (
	"context"
	"testing"
)

func TestPruneDummyKeepsTheBest(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	n, err := s.PruneDummy(ctx, 5)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(dummyScores)-5 {
		t.Fatalf("pruned %d rows, want %d", n, len(dummyScores)-5)
	}

	left, err := s.DummyScores(ctx, 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(left) != 5 {
		t.Fatalf("%d dummy rows left, want 5", len(left))
	}

	if left[0].Score != dummyScores[0].score {
		t.Fatalf("the best dummy score %d was pruned", dummyScores[0].score)
	}
}

func TestPruneDummyNeverTouchesLiveScores(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	if err := s.SaveGame(ctx, testResult("live", 1), nil); err != nil {
		t.Fatal(err)
	}

	if _, err := s.PruneDummy(ctx, 0); err != nil {
		t.Fatal(err)
	}

	top, err := s.TopScores(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(top) != 1 || top[0].ID != "live" {
		t.Fatalf("prune touched live scores: %v", top)
	}

	dummy, err := s.DummyScores(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(dummy) != 0 {
		t.Fatalf("%d dummy rows survived a keep-zero prune", len(dummy))
	}
}
