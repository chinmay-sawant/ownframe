package synth

import (
	"context"
	"testing"
)

func TestKeyboardShortcuts(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Close()

	ctx := context.Background()

	_ = app.onKeyDown(ctx, "a")
	if app.view.ActiveNote != "C4" {
		t.Errorf("got note %q, want C4", app.view.ActiveNote)
	}

	_ = app.onKeyDown(ctx, "w")
	if app.view.ActiveNote != "C#4" {
		t.Errorf("got note %q, want C#4", app.view.ActiveNote)
	}

	_ = app.onKeyDown(ctx, "k")
	if app.view.ActiveNote != "C5" {
		t.Errorf("got note %q, want C5", app.view.ActiveNote)
	}

	_ = app.onKeyDown(ctx, "x")
	if app.view.Octave != 5 {
		t.Errorf("got octave %d, want 5", app.view.Octave)
	}

	_ = app.onKeyDown(ctx, "z")
	if app.view.Octave != 4 {
		t.Errorf("got octave %d, want 4", app.view.Octave)
	}

	_ = app.onKeyDown(ctx, "tab")
	if app.view.Preset != "sawtooth" {
		t.Errorf("got preset %q, want sawtooth", app.view.Preset)
	}

	_ = app.onKeyDown(ctx, "space")
	if !app.view.Playing {
		t.Errorf("expected playing after space")
	}
}

func TestTickAnimation(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Close()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := app.Tick(ctx); err != nil {
			t.Fatalf("Tick() failed: %v", err)
		}
	}

	if len(app.view.Bars) != scopeBarCount {
		t.Fatalf("expected %d bars", scopeBarCount)
	}
	for i, b := range app.view.Bars {
		if b < 0 || b > 100 {
			t.Errorf("bar %d out of bounds: %d", i, b)
		}
	}

	app.view.Playing = true
	app.seqStep = 0
	app.lastStep = app.lastStep.Add(-200)

	if err := app.Tick(ctx); err != nil {
		t.Fatalf("Tick() during playing failed: %v", err)
	}
	if app.view.ActiveNote == "--" {
		t.Errorf("expected active note to update during sequence playback")
	}
}
