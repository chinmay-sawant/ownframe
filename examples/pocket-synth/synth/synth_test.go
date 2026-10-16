package synth

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestNewSynth(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Close()

	if app.Page() == nil {
		t.Fatal("expected non-nil Page")
	}

	w, h := app.Page().Size()
	if w != DefaultWidth || h != DefaultHeight {
		t.Fatalf("got size %dx%d, want %dx%d", w, h, DefaultWidth, DefaultHeight)
	}

	view := app.View()
	if view.ActiveNote != "--" {
		t.Errorf("got note %q, want --", view.ActiveNote)
	}
	if view.Octave != 4 {
		t.Errorf("got octave %d, want 4", view.Octave)
	}
	if view.Preset != "square" {
		t.Errorf("got preset %q, want square", view.Preset)
	}
	if view.Playing {
		t.Errorf("expected playing to be false")
	}
	if len(view.Bars) != scopeBarCount {
		t.Errorf("got %d bars, want %d", len(view.Bars), scopeBarCount)
	}
}

func TestClickControls(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Close()

	ctx := context.Background()

	_ = app.onClick(ctx, ownframe.Box{Action: "preset"})
	if app.view.Preset != "sawtooth" {
		t.Errorf("got preset %q, want sawtooth", app.view.Preset)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "oct-up"})
	if app.view.Octave != 5 {
		t.Errorf("got octave %d, want 5", app.view.Octave)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "oct-down"})
	if app.view.Octave != 4 {
		t.Errorf("got octave %d, want 4", app.view.Octave)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "play-pause"})
	if !app.view.Playing || app.view.Status != "PLAYING" {
		t.Errorf("expected playing state after play click")
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "play-pause"})
	if app.view.Playing {
		t.Errorf("expected paused state after second click")
	}
}
