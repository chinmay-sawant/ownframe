package host

import "context"

// Selector places the caret from a click point. A screen without it keeps
// the old click behavior.
type Selector interface {
	SelectAt(ctx context.Context, x, y float64) error
}

// RangeSelector is the drag surface the window drives while the pointer is
// down: caret, drag, word, and line selection.
type RangeSelector interface {
	Selector
	Drag(ctx context.Context, x, y float64) error
	SelectWordAt(ctx context.Context, x, y float64) error
	SelectLineAt(ctx context.Context, x, y float64) error
}
