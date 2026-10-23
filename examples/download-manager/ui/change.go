package ui

import (
	"context"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// onChange handles the destination file picker: its directory becomes the
// destination folder.
func (a *App) onChange(ctx context.Context, box ownframe.Box) error {
	if box.ID != "dest-file" {
		return nil
	}

	path := a.page.Form().FormValue("dest-file")
	if path == "" {
		return nil
	}

	a.page.Form().SetFormValue("dest-file", "")
	a.page.Form().SetFormValue("dest", filepath.Dir(path))
	a.page.SetData(a.view)

	return nil
}

// onSubmit adds the job when Enter lands in the URL or folder field.
func (a *App) onSubmit(ctx context.Context) error {
	switch a.page.Form().FocusID() {
	case "url", "dest":
		a.addJob()
		a.syncPage()
		a.page.SetData(a.view)
	}

	return nil
}
