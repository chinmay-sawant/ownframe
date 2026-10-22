package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/input"
)

func TestSettingsDefaultWhenEmpty(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	set, err := s.Settings(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := DefaultSettings()
	if fmt.Sprint(set) != fmt.Sprint(want) {
		t.Fatalf("empty settings = %+v, want %+v", set, want)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	s, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ctx := context.Background()

	want := Settings{Keymap: input.DefaultKeymap(), Theme: "dark"}
	want.Keymap.HardDrop = []string{"k"}
	want.Ghost = false

	if err := s.SaveSettings(ctx, want); err != nil {
		t.Fatal(err)
	}

	got, err := s.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}

	want.Theme = "light"
	want.Ghost = true

	if err := s.SaveSettings(ctx, want); err != nil {
		t.Fatal(err)
	}

	got, err = s.Settings(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if got.Theme != "light" || !got.Ghost {
		t.Fatalf("overwrite returned %+v", got)
	}
}
