package window

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

var (
	_ host.Pointer   = (*fakeScreen)(nil)
	_ host.Keyboard  = (*fakeScreen)(nil)
	_ host.Clipboard = (*fakeScreen)(nil)
	_ host.Screen    = (*fakeScreen)(nil)
)

type fakeScreen struct {
	width   int
	height  int
	minW    int
	minH    int
	maxW    int
	maxH    int
	redraws int
	boxes   []layout.Box
}

func (f *fakeScreen) Title() string       { return "fake" }
func (f *fakeScreen) Size() (int, int)    { return f.width, f.height }
func (f *fakeScreen) MinSize() (int, int) { return f.minW, f.minH }
func (f *fakeScreen) SetSize(width, height int) {
	f.width, f.height = f.Clamp(width, height)
}
func (f *fakeScreen) Clamp(width, height int) (int, int) {
	return clampRange(width, f.minW, f.maxW), clampRange(height, f.minH, f.maxH)
}
func (f *fakeScreen) Redraw(context.Context) error {
	f.redraws++

	return nil
}
func clampRange(value, low, high int) int {
	if value < low {
		return low
	}

	if high > 0 && value > high {
		return high
	}

	return value
}
