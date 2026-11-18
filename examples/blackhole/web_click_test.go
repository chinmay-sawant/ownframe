package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestWebClickUsesButtonHitBox(t *testing.T) {
	ctx := context.Background()
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var button ownframe.Box
	for _, box := range app.Page().Boxes() {
		if box.ID == "btn-orbit" {
			button = box
			break
		}
	}
	if button.ID == "" {
		t.Fatal("btn-orbit box missing")
	}
	x, y := button.X+button.W/2, button.Y+button.H/2
	if err := app.Page().Testing().SelectAt(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if !app.orbitPaused {
		t.Fatal("Pause Orbit click did not toggle orbit")
	}
}

// The -web preview executes the template again on every scene frame, so the
// toggled label must live in the template data, not only in a patch.
func TestOrbitLabelSurvivesRedraw(t *testing.T) {
	ctx := context.Background()
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.onClick(ctx, ownframe.Box{Action: "toggle-orbit"}); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, box := range app.Page().Boxes() {
		if box.ID == "btn-orbit" {
			if !strings.Contains(box.Text, "Resume Orbit") {
				t.Fatalf("orbit label after redraw = %q, want Resume Orbit", box.Text)
			}
			return
		}
	}
	t.Fatal("btn-orbit box missing")
}
