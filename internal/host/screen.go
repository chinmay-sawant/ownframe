// Package host is the window and the browser page.
// Both talk to a screen. The program supplies that screen.
package host

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Ticker is a screen that wants one call per frame. The window calls Tick
// before it draws the frame. A screen without Tick is skipped.
type Ticker interface {
	Tick(ctx context.Context) error
}

// Screen is one HTML page the hosts can show.
// The page owns the template, the data, and what a click or a key does.
// It composes Boxer with the Pointer, Keyboard, and Clipboard roles plus
// the window and picture surface; selection, PDF, reload, and devtools
// stay optional host interfaces. Screen stays for compat: entry points
// take it while sinks narrow to one role at the call site.
type Screen interface {
	Boxer
	Pointer
	Keyboard
	Clipboard
	Title() string
	Size() (int, int)
	MinSize() (int, int)
	Clamp(width, height int) (int, int)
	SetSize(width, height int)
	Redraw(ctx context.Context) error
	Image() image.Image
	Display() *layout.Display
	PNG() []byte
	Generation() uint64
}
