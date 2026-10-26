package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func clickID(t *testing.T, app *App, id string) {
	t.Helper()

	box, ok := boxByID(app.Page().Boxes(), id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	if err := app.Page().Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("click %s: %v", id, err)
	}
}

func TestPointerControls(t *testing.T) {
	app := newTestApp(t)

	clickID(t, app, "act-changes")

	if app.View().Mode != "changes" {
		t.Fatalf("changes mode = %q", app.View().Mode)
	}

	clickID(t, app, "act-explorer")

	if app.View().Mode != "explorer" {
		t.Fatalf("explorer mode = %q", app.View().Mode)
	}

	clickID(t, app, "tree-backend/probe.go")

	if app.View().Active != "backend/probe.go" || len(app.View().Tabs) != 2 {
		t.Fatalf("file click active=%q tabs=%d", app.View().Active, len(app.View().Tabs))
	}

	clickID(t, app, "act-collapse")

	if len(app.folders) != 0 {
		t.Fatalf("folders = %v", app.folders)
	}

	clickID(t, app, "act-details")

	if !app.View().Details {
		t.Fatal("details not open")
	}

	clickID(t, app, "det-close")

	if app.View().Details {
		t.Fatal("details not closed")
	}

	clickID(t, app, "act-changes")

	if app.View().Mode != "changes" {
		t.Fatalf("changes mode = %q", app.View().Mode)
	}

	clickID(t, app, "change-web/theme.css")

	if app.View().Diff.Selected != "web/theme.css" {
		t.Fatalf("selected = %q", app.View().Diff.Selected)
	}

	clickID(t, app, "seg-split")

	if !app.View().Diff.Split {
		t.Fatal("split not on")
	}

	clickID(t, app, "sb-devtools")

	if app.devOn {
		t.Fatal("devtools still on")
	}

	clickID(t, app, "act-explorer")
	clickID(t, app, "tab-close-1")

	if len(app.View().Tabs) != 1 {
		t.Fatalf("tabs = %d", len(app.View().Tabs))
	}
}

func TestQuickOpenPalette(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	clickID(t, app, "cmdbar")

	if !app.View().Palette.Open {
		t.Fatal("palette did not open")
	}

	if app.Page().Form().FocusID() != "palette" {
		t.Fatalf("focus = %q", app.Page().Form().FocusID())
	}

	if len(app.View().Palette.Rows) != paletteDefaultRows {
		t.Fatalf("default rows = %d", len(app.View().Palette.Rows))
	}

	app.view.PaletteQuery = "probe"

	if err := app.onChange(ctx, ownframe.Box{ID: "palette"}); err != nil {
		t.Fatal(err)
	}

	rows := app.View().Palette.Rows
	if len(rows) == 0 || rows[0].Path != "backend/probe.go" {
		t.Fatalf("query rows = %+v", rows)
	}

	clickID(t, app, "palette-backend/probe.go")

	if app.View().Palette.Open {
		t.Fatal("palette stayed open")
	}

	clickID(t, app, "cmdbar")

	if !app.View().Palette.Open {
		t.Fatal("palette did not reopen")
	}

	back, ok := boxByID(app.Page().Boxes(), "palette-back")
	if !ok {
		t.Fatal("no palette backdrop")
	}

	if err := app.Page().Click(ctx, back.X+10, back.Y+back.H/2); err != nil {
		t.Fatal(err)
	}

	if app.View().Palette.Open {
		t.Fatal("backdrop did not close the palette")
	}

	if app.View().Active != "backend/probe.go" {
		t.Fatalf("active = %q", app.View().Active)
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "p"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Palette.Open {
		t.Fatal("ctrl+p did not open the palette")
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if app.View().Palette.Open {
		t.Fatal("escape did not close the palette")
	}
}

func TestMarkdownPreview(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if app.View().Lang != langMarkdown {
		t.Fatalf("initial language = %s", app.View().Lang)
	}

	app.togglePreview()

	if !app.View().Preview || len(app.View().PreviewRows) == 0 {
		t.Fatalf("preview = %v rows = %d", app.View().Preview, len(app.View().PreviewRows))
	}

	first := lineText(app.View().PreviewRows[0].Segments)
	if first != "Beacon" {
		t.Fatalf("first preview row = %q", first)
	}

	app.togglePreview()

	if app.View().Preview || len(app.View().Rows) == 0 {
		t.Fatalf("source restore preview=%v rows=%d", app.View().Preview, len(app.View().Rows))
	}

	if lineText(app.View().Rows[0].Segments) != "# Beacon" {
		t.Fatalf("first source row = %q", lineText(app.View().Rows[0].Segments))
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "shift"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "v"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Preview {
		t.Fatal("ctrl+shift+v did not turn preview on")
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "shift"); err != nil {
		t.Fatal(err)
	}

	app.openFile("backend/probe.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().Previewable || app.View().Preview {
		t.Fatal("preview leaked onto a Go file")
	}

	if _, ok := boxByID(app.Page().Boxes(), "preview-toggle"); ok {
		t.Fatal("preview button shown for a Go file")
	}
}

func TestPreviewButton(t *testing.T) {
	app := newTestApp(t)

	clickID(t, app, "preview-toggle")

	if !app.View().Preview {
		t.Fatal("preview button did not turn preview on")
	}

	clickID(t, app, "preview-toggle")

	if app.View().Preview {
		t.Fatal("preview button did not turn preview off")
	}
}

func TestActionControlsHaveIDs(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	check := func() {
		t.Helper()

		for _, box := range app.Page().Boxes() {
			if box.Action != "" && box.ID == "" {
				t.Errorf("action %q has no id at (%.0f,%.0f)", box.Action, box.X, box.Y)
			}
		}
	}

	check()

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}

	check()

	app.togglePreview()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	if err := app.openPalette(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	if err := app.closePalette(ctx); err != nil {
		t.Fatal(err)
	}

	app.toggleDetails()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()

	app.showToast("info", "Check", "The toast has an id.")
	app.refresh()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	check()
}
