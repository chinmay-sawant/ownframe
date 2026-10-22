package store

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
)

func TestSecondInstance(t *testing.T) {
	dir := t.TempDir()

	a, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = a.Close() }()

	b, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = b.Close() }()

	sa, err := a.EnsureDummy(bg(), DummyOptions{Count: 30})
	if err != nil {
		t.Fatal(err)
	}

	sb, err := b.EnsureDummy(bg(), DummyOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if !sa.Seeded || sb.Seeded {
		t.Fatalf("seeded flags a=%v b=%v", sa.Seeded, sb.Seeded)
	}

	src := sa.Sources[0]

	_, err = a.Commit(bg(), src.ID, CommitMeta{
		Generation: 1, Position: src.Position + 1, State: entry.StateLive,
	}, []entry.Entry{{
		Session: sa.Session.ID, Source: src.ID, Position: src.Position,
		Generation: 1, Message: "from a", Received: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	n, err := b.Count(bg(), Query{Source: &src.ID})
	if err != nil {
		t.Fatal(err)
	}

	if n < 1 {
		t.Fatalf("second instance sees %d entries", n)
	}

	got, err := b.Source(bg(), src.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Position != src.Position+1 {
		t.Fatalf("second instance checkpoint = %d", got.Position)
	}
}
