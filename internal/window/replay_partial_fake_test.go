package window

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// dirtyScreen is a fakeScreen that keeps a display list and reports a
// scripted damage list, the way a page with TakeDirtyRects does.
type dirtyScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	layouts int
	rects   []image.Rectangle
	ok      bool
	takes   int
	ticking bool
	frame   bool
}

func (d *dirtyScreen) FrameDirty() bool { return d.frame }

func (d *dirtyScreen) Display() *layout.Display { return d.display }
func (d *dirtyScreen) Generation() uint64       { return d.gen }
func (d *dirtyScreen) Ticking() bool            { return d.ticking }

func (d *dirtyScreen) TakeDirtyRects() ([]image.Rectangle, bool) {
	d.takes++

	return d.rects, d.ok
}

// click stands in for one counter click: the page lays out once and reports
// one changed box.
func (d *dirtyScreen) click(rect image.Rectangle) {
	d.gen++
	d.layouts++
	d.rects = []image.Rectangle{rect}
	d.ok = true
}

func newDirtyShell(screen *dirtyScreen) *shell {
	return &shell{
		app: screen, ctx: context.Background(),
		display: screen.display, seq: screen.gen,
		screenW: screen.width, screenH: screen.height,
	}
}

// oldScreen is a page from before the damage list: it reports one rect
// through TakeDirty only, so the window wraps it as a one-element list.
type oldScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	rect    image.Rectangle
	ok      bool
	takes   int
}

func (o *oldScreen) Display() *layout.Display { return o.display }
func (o *oldScreen) Generation() uint64       { return o.gen }

func (o *oldScreen) TakeDirty() (image.Rectangle, bool) {
	o.takes++

	return o.rect, o.ok
}
