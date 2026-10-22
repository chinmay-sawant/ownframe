package drop

import (
	"context"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe"
)

// onDrop adds the base name of every dropped file. The whole window takes
// a drop, so any file can land anywhere in it. Path stays on file for the
// call but is never rendered, so an absolute desktop path cannot leak into
// the page; in a browser Path is empty and the entry's name stands in.
func (a *App) onDrop(_ context.Context, files []ownframe.Drop) error {
	for _, file := range files {
		name := file.Name
		if name == "" {
			name = filepath.Base(file.Path)
		} else {
			name = filepath.Base(name)
		}
		if name == "." || name == "/" {
			continue
		}
		a.add(name, "")
	}
	return nil
}
