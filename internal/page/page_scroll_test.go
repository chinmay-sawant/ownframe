package page

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestScrollRequests(t *testing.T) {
	t.Parallel()

	p := &Page{}

	p.ScrollTo(10, 20)
	req, ok := p.TakeScroll()
	if !ok || req.X != 10 || req.Y != 20 || !req.Absolute {
		t.Fatalf("scroll to %+v %v", req, ok)
	}

	if _, ok := p.TakeScroll(); ok {
		t.Fatal("scroll not cleared")
	}

	p.ScrollBy(1, 2)
	p.ScrollBy(3, 4)
	req, ok = p.TakeScroll()
	if !ok || req.X != 4 || req.Y != 6 || req.Absolute {
		t.Fatalf("scroll by %+v %v", req, ok)
	}

	p.ScrollTo(5, 6)
	p.ScrollBy(1, 1)
	req, ok = p.TakeScroll()
	if !ok || req.X != 6 || req.Y != 7 || !req.Absolute {
		t.Fatalf("scroll replace %+v %v", req, ok)
	}
}

func TestScrollOffsetRoundTrip(t *testing.T) {
	t.Parallel()

	p := &Page{}
	p.SetScrollOffset(3, 9)
	x, y := p.ScrollOffset()
	if x != 3 || y != 9 || p.scroll.Windowing.X != 3 || p.scroll.Windowing.Y != 9 {
		t.Fatalf("offset %d %d window %d %d", x, y, p.scroll.Windowing.X, p.scroll.Windowing.Y)
	}
}

func TestScrollNewAllowsPan(t *testing.T) {
	p, err := New(Config{HTML: `<p>scroll</p>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if !p.AllowScroll() || p.ViewportPinZ() != 0 {
		t.Fatalf("allow %v pin %d", p.AllowScroll(), p.ViewportPinZ())
	}
	p.SetAllowScroll(false)
	p.SetViewportPinZ(4)
	if p.AllowScroll() || p.ViewportPinZ() != 4 {
		t.Fatalf("set allow %v pin %d", p.AllowScroll(), p.ViewportPinZ())
	}
}

func TestScrollNilPage(t *testing.T) {
	t.Parallel()

	var p *Page
	if !p.AllowScroll() {
		t.Fatal("nil AllowScroll")
	}
	p.SetAllowScroll(false)
	p.SetTouchScrollOptions(TouchScrollOptions{Sensitivity: 2})
	if p.ScrollbarMode() != host.ScrollbarAuto {
		t.Fatal("nil scrollbar")
	}
	if got := p.TouchScrollOptions(); got.Sensitivity != 1 || got.Deceleration != 5 {
		t.Fatalf("nil options %+v", got)
	}
}
