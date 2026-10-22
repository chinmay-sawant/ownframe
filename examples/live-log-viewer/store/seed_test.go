package store

import (
	"testing"
)

func TestEnsureDummySeedsOnce(t *testing.T) {
	st := newStore(t)

	setup, err := st.EnsureDummy(bg(), DummyOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if !setup.Seeded || len(setup.Sources) != 3 {
		t.Fatalf("setup = %+v", setup)
	}

	s1, err := st.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if s1.Entries < 10000 || s1.Entries > 10010 {
		t.Fatalf("seeded entries = %d", s1.Entries)
	}

	for _, src := range setup.Sources {
		if src.Position == 0 {
			t.Fatalf("source %d has no checkpoint", src.ID)
		}
	}

	again, err := st.EnsureDummy(bg(), DummyOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if again.Seeded {
		t.Fatal("second call reseeded")
	}

	s2, err := st.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if s2.Entries != s1.Entries {
		t.Fatalf("entry count changed %d -> %d", s1.Entries, s2.Entries)
	}

	dir := st.Dir()

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}

	st2, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st2.Close() }()

	third, err := st2.EnsureDummy(bg(), DummyOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if third.Seeded {
		t.Fatal("reopen reseeded")
	}

	s3, err := st2.Stats(bg())
	if err != nil {
		t.Fatal(err)
	}

	if s3.Entries != s1.Entries {
		t.Fatalf("reopen count %d, want %d", s3.Entries, s1.Entries)
	}
}
