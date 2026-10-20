// Package corebackend adapts the spreadsheet core (workbook, formula,
// storage) to the ui.Backend seam. One Backend is owned by the ui worker
// goroutine, so the UI loop never calls the workbook directly. Storage
// calls run on that same goroutine, off the UI loop.
package corebackend

import (
	"path/filepath"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/storage"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// Backend is one open store and the workbook the screen edits.
type Backend struct {
	st   *storage.Store
	wb   *workbook.Workbook
	undo []step
	redo []step
}

// step is one saved command and the command that reverses it.
type step struct {
	fwd workbook.Command
	inv workbook.Command
}

// Open opens the store in dir. An empty dir means the core default under
// os.UserConfigDir()/ownframe/spreadsheet.
func Open(dir string) (*Backend, error) {
	st, err := storage.Open(dir)
	if err != nil {
		return nil, err
	}

	return &Backend{st: st}, nil
}

// Dir is the data directory the store lives in, for exports.
func (b *Backend) Dir() string {
	path := b.st.Path()
	if path == storage.Memory || path == "" {
		return "."
	}

	return filepath.Dir(path)
}
