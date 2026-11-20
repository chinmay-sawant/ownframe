package page

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"testing"
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

func TestScrollbarModeConfig(t *testing.T) {
	p, err := New(Config{
		HTML: `<p>scroll</p>`, Width: 100, Height: 100,
		Scrollbar: host.ScrollbarHidden,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.ScrollbarMode(); got != host.ScrollbarHidden {
		t.Fatalf("ScrollbarMode() = %v", got)
	}
}

func TestTouchScrollOptionsDefaultAndClamp(t *testing.T) {
	p := &Page{}
	if got := p.TouchScrollOptions(); got.Sensitivity != 1 || got.Deceleration != 5 {
		t.Fatalf("default options = %+v", got)
	}
	p.SetTouchScrollOptions(TouchScrollOptions{Sensitivity: .1, Deceleration: 20})
	if got := p.TouchScrollOptions(); got.Sensitivity != .25 || got.Deceleration != 12 {
		t.Fatalf("clamped options = %+v", got)
	}
}

func TestTouchZoomConfigAndOverride(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if !p.TouchZoomAllowed() {
		t.Fatal("touch zoom should remain enabled by default")
	}
	p.SetTouchZoomAllowed(false)
	if p.TouchZoomAllowed() {
		t.Fatal("runtime override did not disable touch zoom")
	}
	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableTouchZoom: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.TouchZoomAllowed() {
		t.Fatal("config did not disable touch zoom")
	}
}

func TestWheelZoomConfig(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if !p.WheelZoomAllowed() {
		t.Fatal("wheel zoom should remain enabled by default")
	}

	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableWheelZoom: true})
	if err != nil {
		t.Fatal(err)
	}

	if p.WheelZoomAllowed() {
		t.Fatal("config did not disable wheel zoom")
	}
}

func TestKeyZoomConfig(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if !p.KeyZoomAllowed() {
		t.Fatal("key zoom should remain enabled by default")
	}

	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableKeyZoom: true})
	if err != nil {
		t.Fatal(err)
	}

	if p.KeyZoomAllowed() {
		t.Fatal("config did not disable key zoom")
	}

	if !p.WheelZoomAllowed() {
		t.Fatal("key zoom config also disabled wheel zoom")
	}
}

func TestPageWheelHandler(t *testing.T) {
	t.Parallel()

	p := &Page{}
	if handled, err := p.Wheel(context.Background(), 0, 0, 0, 1); handled || err != nil {
		t.Fatalf("no handler = %v %v", handled, err)
	}

	var got struct {
		x, y, dx, dy float64
	}

	p.handlers.Wheel = func(_ context.Context, x, y, dx, dy float64) (bool, error) {
		got.x, got.y, got.dx, got.dy = x, y, dx, dy

		return true, nil
	}

	handled, err := p.Wheel(context.Background(), 12, 34, 5, 6)
	if err != nil || !handled {
		t.Fatalf("handler = %v %v", handled, err)
	}

	if got.x != 12 || got.y != 34 || got.dx != 5 || got.dy != 6 {
		t.Fatalf("args = %+v", got)
	}
}
