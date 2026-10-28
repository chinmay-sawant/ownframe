package window

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestZoomByClamps checks the keyboard zoom steps and the range clamp.
func TestZoomByClamps(t *testing.T) {
	s := &shell{pageZoom: 1}

	s.zoomBy(zoomStep)
	if math.Abs(s.zoom()-zoomStep) > 1e-9 {
		t.Fatalf("zoom after one step = %v", s.zoom())
	}

	for i := 0; i < 100; i++ {
		s.zoomBy(zoomStep)
	}

	if s.zoom() != maxZoom {
		t.Fatalf("zoom at the top = %v", s.zoom())
	}

	for i := 0; i < 200; i++ {
		s.zoomBy(1 / zoomStep)
	}

	if s.zoom() != minZoom {
		t.Fatalf("zoom at the bottom = %v", s.zoom())
	}
}

// TestZoomKey checks that only the command modifier consumes =, -, and 0.
func TestZoomKey(t *testing.T) {
	s := &shell{pageZoom: 1}

	if !s.zoomKey(ebiten.KeyEqual, true, modifiers{Control: true}) {
		t.Fatal("Control+= was not consumed")
	}

	if s.zoomKey(ebiten.KeyEqual, true, modifiers{}) {
		t.Fatal("plain = was consumed")
	}

	before := s.zoom()
	s.zoomKey(ebiten.KeyMinus, true, modifiers{Control: true})

	if s.zoom() >= before {
		t.Fatalf("Control+- zoom = %v, before %v", s.zoom(), before)
	}

	s.zoomKey(ebiten.Key0, true, modifiers{Control: true})

	if s.zoom() != 1 {
		t.Fatalf("zoom reset = %v", s.zoom())
	}
}

type noWheelZoomScreen struct{ fakeScreen }

func (noWheelZoomScreen) WheelZoomAllowed() bool { return false }

type noKeyZoomScreen struct{ fakeScreen }

func (noKeyZoomScreen) KeyZoomAllowed() bool { return false }

func TestDisabledWheelZoomStaysStill(t *testing.T) {
	off := &shell{app: &noWheelZoomScreen{}, pageZoom: 1}
	if off.wheelMayZoom(1) {
		t.Fatal("a page that refuses wheel zoom still accepted a notch")
	}

	if off.zoom() != 1 {
		t.Fatalf("zoom = %v", off.zoom())
	}

	on := &shell{app: &fakeScreen{}, pageZoom: 1}
	if !on.wheelMayZoom(1) {
		t.Fatal("the default page rejected wheel zoom")
	}
}

func TestDisabledKeyZoomStaysStill(t *testing.T) {
	off := &shell{app: &noKeyZoomScreen{}, pageZoom: 1}
	mods := modifiers{Control: true}
	keys := []ebiten.Key{
		ebiten.KeyEqual,
		ebiten.KeyKPAdd,
		ebiten.KeyMinus,
		ebiten.KeyKPSubtract,
		ebiten.Key0,
		ebiten.KeyKP0,
	}

	for _, key := range keys {
		if !off.zoomKey(key, true, mods) {
			t.Fatalf("Control+%v was not consumed", key)
		}
	}

	if off.zoom() != 1 {
		t.Fatalf("zoom = %v", off.zoom())
	}

	plain := &shell{pageZoom: 1}
	plain.zoomKey(ebiten.KeyEqual, true, mods)

	if plain.zoom() == 1 {
		t.Fatal("a page without the policy rejected Control+=")
	}
}

// TestWheelZoomFactor checks trackpad deltas scale by the same step.
func TestWheelZoomFactor(t *testing.T) {
	if got := wheelZoomFactor(0); got != 1 {
		t.Fatalf("wheelZoomFactor(0) = %v", got)
	}

	if got := wheelZoomFactor(1); math.Abs(got-zoomStep) > 1e-9 {
		t.Fatalf("wheelZoomFactor(1) = %v", got)
	}

	half := wheelZoomFactor(0.5)

	if math.Abs(half*half-zoomStep) > 1e-9 {
		t.Fatalf("half notch = %v", half)
	}
}

// TestZoomCombinesPinch checks the page zoom multiplies into the pinch scale.
func TestZoomCombinesPinch(t *testing.T) {
	s := &shell{pageZoom: 2}
	s.fingers.zoom = 1.5

	if got := s.zoom(); math.Abs(got-3) > 1e-9 {
		t.Fatalf("combined zoom = %v", got)
	}
}
