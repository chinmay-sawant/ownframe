package page

import (
	"context"
	"net/url"
)

// Handlers are optional callbacks.
type Handlers struct {
	Hover HoverHandler
	Click func(ctx context.Context, box Box) error
	// ClickPaint runs after Click and a control's own change. It reports
	// that the handler painted the frame, so the page skips the layout.
	ClickPaint func(ctx context.Context, box Box) (bool, error)
	// Press can claim the pointer-down.
	Press func(ctx context.Context, box Box, x, y float64) (bool, error)
	// LongPress handles a hold.
	LongPress func(ctx context.Context, box Box) error
	// KeyDown and KeyUp take lowercase names.
	KeyDown func(ctx context.Context, key string) error
	KeyUp   func(ctx context.Context, key string) error
	Swipe   func(ctx context.Context, dx, dy float64) error
	// DragStart claims a gesture.
	DragStart func(ctx context.Context, box Box) (bool, error)
	// DragMove takes CSS deltas.
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
	// Drop readers work in the callback.
	Drop func(ctx context.Context, files []Drop) error
	// OpenLink takes validated links.
	OpenLink func(ctx context.Context, target *url.URL) error
	// DeepLink takes host URLs.
	DeepLink func(ctx context.Context, target *url.URL) error
}
