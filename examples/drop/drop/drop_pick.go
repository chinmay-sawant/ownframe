package drop

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// pickID is the template's file control.
const pickID = "pick"

// onChange adds the path the picker stored in the file control.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	if box.ID != pickID {
		return nil
	}

	a.add(a.page.Form().FormValue(pickID), "")

	return nil
}

// add stores the path of one file, preferring the absolute path.
func (a *App) add(path, name string) {
	if path == "" {
		path = name
	}

	if path == "" {
		return
	}

	a.view.Paths = append(a.view.Paths, path)
}
