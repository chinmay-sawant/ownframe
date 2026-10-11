package page

import (
	"context"
	"net/url"
)

// Handlers are optional callbacks.
type Handlers struct {
	Hover HoverHandler
	Click func(ctx context.Context, box Box) error
	// Press can claim pointer-down activation.
	Press func(ctx context.Context, box Box, x, y float64) (bool, error)

	// LongPress handles a hold.
	LongPress func(ctx context.Context, box Box) error

	// KeyDown receives lowercase names and does not draw.
	KeyDown func(ctx context.Context, key string) error

	// KeyUp receives released names and does not draw.
	KeyUp func(ctx context.Context, key string) error

	Swipe func(ctx context.Context, dx, dy float64) error

	// DragStart claims a pointer gesture beginning inside a box.
	DragStart func(ctx context.Context, box Box) (bool, error)
	// DragMove receives CSS-pixel deltas while claimed.
	DragMove func(ctx context.Context, dx, dy float64) error
	DragEnd  func(ctx context.Context) error

	// BeforeEdit may reject an edit.
	BeforeEdit func(ctx context.Context, box Box) error

	// Change runs after edits; errors skip redraw.
	Change     func(ctx context.Context, box Box) error
	Type       func(ctx context.Context, text string) error
	Backspace  func(ctx context.Context) error
	DeleteWord func(ctx context.Context) error
	Submit     func(ctx context.Context) error
	Copy       func(ctx context.Context) (text string, ok bool, err error)
	Cut        func(ctx context.Context) (text string, ok bool, err error)
	Paste      func(ctx context.Context, text string) error
	SelectAll  func(ctx context.Context) error
	Undo       func(ctx context.Context) error
	Redo       func(ctx context.Context) error

	// Drop readers work only during this callback.
	Drop func(ctx context.Context, files []Drop) error
	// OpenLink receives validated outbound links.
	OpenLink func(ctx context.Context, target *url.URL) error
	// DeepLink handles validated host-delivered URLs on the page loop.
	DeepLink func(ctx context.Context, target *url.URL) error
}
