package store

import (
	"context"
	"testing"
)

func TestSeedIsIdempotentOnReopen(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s2.Close() }()
	scores, err := s2.DummyScores(context.Background(), 50)
	if err != nil {
		t.Fatal(err)
	}

	if len(scores) != len(dummyScores) {
		t.Fatalf("%d dummy scores after reopen, want %d", len(scores), len(dummyScores))
	}
}

func TestReopenDoesNotOverwriteUserChanges(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	s.exec(t, `UPDATE scores SET score = 12345 WHERE id = 'dummy-0001'`)

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	s2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s2.Close() }()
	scores, err := s2.DummyScores(context.Background(), 50)
	if err != nil {
		t.Fatal(err)
	}

	found := false

	for _, r := range scores {
		if r.ID != "dummy-0001" {
			continue
		}

		found = true

		if r.Score != 12345 {
			t.Fatalf("reseed overwrote the score to %d", r.Score)
		}
	}

	if !found {
		t.Fatal("dummy-0001 disappeared on reopen")
	}
}

func TestDummyScoresAreOrderedByScore(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	scores, err := s.DummyScores(context.Background(), 50)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(scores); i++ {
		if scores[i-1].Score < scores[i].Score {
			t.Fatalf("rank %d scores %d before %d", i, scores[i-1].Score, scores[i].Score)
		}
	}

	for _, r := range scores {
		if !r.Dummy {
			t.Fatalf("live row %s appeared in the dummy list", r.ID)
		}
	}
}
