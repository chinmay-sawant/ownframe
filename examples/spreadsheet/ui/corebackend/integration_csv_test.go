package corebackend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

func TestUICSVImportThroughPicker(t *testing.T) {
	b := newBackend(t, t.TempDir())
	app, err := ui.New(ui.Options{Backend: b, Width: 1000, Height: 700, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = app.Close() }()

	page := app.Page()
	waitFor(t, app, func() bool { return hasCell(app, "row 1") })

	path := filepath.Join(t.TempDir(), "in.csv")
	if err := os.WriteFile(path, []byte("alpha,beta\n1,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	page.Testing().SetPicker(func(context.Context, string) (string, error) {
		return path, nil
	})

	// Open the dialog from the toolbar, then click the file field.
	clickToolbar(t, app, "import")
	clickControl(t, app, "csvpath")
	waitFor(t, app, func() bool { return app.View().Dialog != nil && app.View().Dialog.Note != "" })

	// The preview keeps the grid as it was until Commit.
	if !hasCell(app, "row 1") {
		t.Fatal("preview changed the rendered workbook")
	}

	clickDialog(t, app, "csv-commit")
	waitFor(t, app, func() bool { return hasCell(app, "alpha") && !hasCell(app, "row 1") })
}
