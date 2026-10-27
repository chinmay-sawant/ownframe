package host_test

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

type stubFront struct {
	display *layout.Display
	img     image.Image
	w, h    int
	gen     uint64
}

func (s stubFront) Display() *layout.Display { return s.display }
func (s stubFront) Image() image.Image       { return s.img }
func (s stubFront) Size() (int, int)         { return s.w, s.h }
func (s stubFront) Generation() uint64       { return s.gen }

func TestSelectFrontPreference(t *testing.T) {
	t.Parallel()

	idle := host.SelectFront(stubFront{w: 3, h: 4, gen: 1})
	if idle.Display != nil || idle.Image != nil || idle.Fallback || idle.Width != 3 || idle.Height != 4 || idle.Generation != 1 {
		t.Fatalf("idle %+v", idle)
	}

	d := &layout.Display{Width: 8, Height: 9}
	list := host.SelectFront(stubFront{display: d, img: image.NewNRGBA(image.Rect(0, 0, 1, 1)), gen: 2})
	if list.Display != d || list.Image != nil || list.Fallback || list.Width != 8 || list.Height != 9 || list.Generation != 2 {
		t.Fatalf("list %+v", list)
	}

	img := image.NewNRGBA(image.Rect(0, 0, 5, 6))
	fb := host.SelectFront(stubFront{img: img, gen: 3})
	if fb.Display != nil || fb.Image != img || !fb.Fallback || fb.Width != 5 || fb.Height != 6 || fb.Generation != 3 {
		t.Fatalf("fallback %+v", fb)
	}
}
