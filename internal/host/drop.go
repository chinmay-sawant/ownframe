package host

import "context"

// Drop is one file or directory dropped on the window. Path is absolute on
// desktop and empty in a browser. Read serves the bytes only during the
// Drop call: copy what is needed before returning, since a later Read
// fails closed after the frame.
type Drop struct {
	Name  string
	Size  int64
	IsDir bool
	Path  string
	Read  func() ([]byte, error)
}

// Dropper is a screen that accepts dropped files.
type Dropper interface {
	Drop(ctx context.Context, files []Drop) error
}
