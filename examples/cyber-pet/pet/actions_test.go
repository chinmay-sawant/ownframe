package pet

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestSleepAndTalk(t *testing.T) {
	a, _ := newT(t)
	a.Sleep()
	if !a.State().Sleeping || a.State().Mood != MoodSleeping {
		t.Fatal("pet did not enter sleep")
	}
	a.Sleep()
	if a.State().Sleeping {
		t.Fatal("pet did not wake up")
	}

	h0 := a.State().Happiness
	a.Talk()
	if a.State().Happiness <= h0 || a.State().Speech == "" {
		t.Fatal("talk failed")
	}
}

func TestClickButtons(t *testing.T) {
	a, _ := newT(t)
	ctx := context.Background()

	a.state.Hunger = 20.0
	if err := a.onClick(ctx, ownframe.Box{ID: "feed"}); err != nil {
		t.Fatal(err)
	}
	if a.State().Hunger != 45.0 {
		t.Fatalf("feed click hunger %f", a.State().Hunger)
	}

	if err := a.onClick(ctx, ownframe.Box{ID: "pet-face"}); err != nil {
		t.Fatal(err)
	}
	if a.State().Mood != MoodPurring {
		t.Fatalf("face click mood %s", a.State().Mood)
	}
}

func TestKeyboardShortcuts(t *testing.T) {
	a, _ := newT(t)
	ctx := context.Background()

	a.state.Hunger = 10.0
	if err := a.onKeyDown(ctx, "1"); err != nil {
		t.Fatal(err)
	}
	if a.State().Hunger != 35.0 {
		t.Fatalf("key 1 hunger %f", a.State().Hunger)
	}

	if err := a.onKeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}
	if a.State().Mood != MoodPurring {
		t.Fatalf("space mood %s", a.State().Mood)
	}
}

func TestDragPetFace(t *testing.T) {
	a, c := newT(t)
	ctx := context.Background()

	ok, err := a.onDragStart(ctx, ownframe.Box{ID: "face-body"})
	if err != nil || !ok {
		t.Fatalf("drag start: %v, %v", ok, err)
	}
	c.adv(300 * time.Millisecond)
	if err := a.onDragMove(ctx, 10, 5); err != nil {
		t.Fatal(err)
	}
	if a.State().Mood != MoodPurring {
		t.Fatalf("drag mood: %s", a.State().Mood)
	}
	if err := a.onDragEnd(ctx); err != nil {
		t.Fatal(err)
	}
}
