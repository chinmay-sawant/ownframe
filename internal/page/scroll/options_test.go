package scroll

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestScrollTouchClamp(t *testing.T) {
	t.Parallel()

	var s State
	if s.Sensitivity() != sensDefault || s.Deceleration() != decelDefault {
		t.Fatalf("zero %+v", s.touch)
	}
	if (*State)(nil).Sensitivity() != sensDefault || (*State)(nil).Deceleration() != decelDefault {
		t.Fatal("nil")
	}

	s.SetTouch(Touch{Sensitivity: 0.1, Deceleration: 20})
	if s.Sensitivity() != sensMin || s.Deceleration() != decelMax {
		t.Fatalf("clamp %v %v", s.Sensitivity(), s.Deceleration())
	}

	s.SetTouch(Touch{Sensitivity: 0, Deceleration: -1})
	if s.Sensitivity() != sensDefault || s.Deceleration() != decelDefault {
		t.Fatalf("non-positive %v %v", s.Sensitivity(), s.Deceleration())
	}

	s.SetTouch(Touch{Sensitivity: math.NaN(), Deceleration: math.Inf(1)})
	if s.Sensitivity() != sensDefault || s.Deceleration() != decelDefault {
		t.Fatalf("non-finite %v %v", s.Sensitivity(), s.Deceleration())
	}

	s.SetTouch(Touch{Sensitivity: 0.5, Deceleration: 0.5})
	if s.Sensitivity() != 0.5 || s.Deceleration() != decelMin {
		t.Fatalf("partial %v %v", s.Sensitivity(), s.Deceleration())
	}
}

func TestScrollReadyAllowAndPin(t *testing.T) {
	t.Parallel()

	var zero State
	if zero.Allow() || zero.Bar() != host.ScrollbarAuto || zero.PinZ() != 0 {
		t.Fatalf("zero allow %v bar %v pin %d", zero.Allow(), zero.Bar(), zero.PinZ())
	}

	s := New(host.ScrollbarHidden)
	if !s.Allow() || s.Bar() != host.ScrollbarHidden || s.Sensitivity() != sensDefault {
		t.Fatalf("new allow %v bar %v sens %v", s.Allow(), s.Bar(), s.Sensitivity())
	}
	s.SetAllow(false)
	s.SetPinZ(4)
	if s.Allow() || s.PinZ() != 4 {
		t.Fatalf("set allow %v pin %d", s.Allow(), s.PinZ())
	}
}
