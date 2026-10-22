package corebackend

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

func clickToolbar(t *testing.T, app *ui.App, action string) {
	t.Helper()

	for _, box := range app.View().Toolbar {
		if string(box.Action) == action {
			clickBox(t, app, float64(box.X+box.W/2), float64(box.Y+box.H/2))

			return
		}
	}

	t.Fatalf("toolbar action %q missing", action)
}

func clickDialog(t *testing.T, app *ui.App, action string) {
	t.Helper()

	for _, box := range app.Page().Boxes() {
		if box.Action == action {
			clickBox(t, app, box.X+box.W/2, box.Y+box.H/2)

			return
		}
	}

	t.Fatalf("dialog action %q missing", action)
}

func clickControl(t *testing.T, app *ui.App, id string) {
	t.Helper()

	for _, box := range app.Page().Boxes() {
		if box.ID == id {
			clickBox(t, app, box.X+box.W/2, box.Y+box.H/2)

			return
		}
	}

	t.Fatalf("control %q missing", id)
}

func clickBox(t *testing.T, app *ui.App, x, y float64) {
	t.Helper()

	if err := app.Page().Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
}
