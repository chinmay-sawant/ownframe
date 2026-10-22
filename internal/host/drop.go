package host

import "context"

// Drop is one file or directory dropped on the window.
type Drop struct {
	Name  string
	Size  int64
	IsDir bool
	// Path is the local desktop path and is empty in a browser. It is a
	// sensitive local path. Copy bytes from Read during the Drop callback;
	// do not keep Path after return. A later Read fails closed.
	Path string
	Read func() ([]byte, error)
}

// Dropper is a screen that accepts dropped files.
type Dropper interface {
	Drop(ctx context.Context, files []Drop) error
}
