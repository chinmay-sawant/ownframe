package host

import "context"

// WheelHandler is a screen that can scroll a region under the cursor. The
// coordinates are CSS pixels in page space. Returning true consumes the
// wheel so the window does not move the page.
type WheelHandler interface {
	Wheel(ctx context.Context, x, y, dx, dy float64) (bool, error)
}
