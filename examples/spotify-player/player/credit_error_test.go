package player

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/music"
)

func TestCreditShowsAudioError(t *testing.T) {
	engine := music.NewEngine(&fakeResolver{clip: music.Clip{
		Title: "Test clip", Creator: "Tester", License: "CC0 1.0", Data: []byte("x"),
	}})
	engine.Open = func([]byte) (music.Voice, error) { return nil, errors.New("boom") }

	app, err := NewWith("", engine)
	if err != nil {
		t.Fatalf("NewWith: %v", err)
	}

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	app.apply("test", sampleTracks(), nil, nil)

	for i := 0; i < 2000 && app.audio.Err() == nil; i++ {
		if err := app.Tick(context.Background()); err != nil {
			t.Fatalf("Tick: %v", err)
		}

		time.Sleep(time.Millisecond)
	}

	if app.audio.Err() == nil {
		t.Fatal("the open error never surfaced")
	}

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	// The tick patches only the retained op; template truth moves on commit.
	d := app.Page().Probe().Display()
	credit, ok := findBox(app.Boxes(), "credit")
	if !ok {
		t.Fatal("no credit box")
	}

	if op := ownframe.Text(d, credit); op == nil || op.Text != "Audio unavailable" {
		t.Fatalf("credit op = %+v (loading=%v err=%v)", op, app.audio.Loading(), app.audio.Err())
	}

	app.syncViewFromAudio()
	app.page.SetData(app.view)

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if got := app.View().Credit; got != "Audio unavailable" {
		t.Fatalf("Credit = %q (loading=%v err=%v)", got, app.audio.Loading(), app.audio.Err())
	}
}
