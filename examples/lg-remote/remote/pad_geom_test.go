package remote

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestPadBoxIsTheHitTarget(t *testing.T) {
	app := newTest(t, WithPhone(true))
	clickID(t, app, "tab-pad")

	var pad, note ownframe.Box
	for _, item := range app.Page().Boxes() {
		if item.ID == "pad" {
			pad = item
		}
		if item.Tag == "span" && item.Y > pad.Y && item.Y+item.H < pad.Y+pad.H {
			note = item
		}
	}
	if pad.H < 290 || pad.H > 310 {
		t.Fatalf("pad h %.1f", pad.H)
	}
	if pad.Y+pad.H > 1000 {
		t.Fatalf("pad off screen %.1f", pad.Y+pad.H)
	}
	if note.ID != "" || note.W < 1 {
		t.Fatalf("hint id %q w %.1f", note.ID, note.W)
	}

	fake := &fakeLink{}
	app.SetLink(fake)
	err := app.Page().Click(context.Background(), note.X+note.W/2, note.Y+note.H/2)
	if err != nil {
		t.Fatal(err)
	}
	if err = app.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if fake.spec != "click:" {
		t.Fatalf("note click %s", fake.spec)
	}
}
