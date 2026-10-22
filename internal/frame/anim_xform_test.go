package frame

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// near compares two floats within a small epsilon.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestScalePatchesMatrix(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := ScaleTo(d, []*page.DisplayOp{op}, 2, 100*time.Millisecond, Linear)

	a.Step(t0)

	if !op.XformSet {
		t.Fatal("no transform set")
	}

	a.Step(t0.Add(50 * time.Millisecond))

	// The card is 10..30 x 10..20 points, center (20,15). At 1.5x about
	// the center: E = 20 - 1.5*20 = -10, F = 15 - 1.5*15 = -7.5.
	m := op.Transform()
	if !near(m.A, 1.5) || !near(m.D, 1.5) || !near(m.E, -10) || !near(m.F, -7.5) {
		t.Fatalf("matrix = %+v", m)
	}

	cx, cy := m.Apply(20, 15)
	if !near(cx, 20) || !near(cy, 15) {
		t.Fatalf("center moved to %v,%v", cx, cy)
	}
}

func TestScaleComposesBakedTransform(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()

	var zero layout.DisplayOp
	base := zero.Transform()
	base.E = 5
	op.SetXform(base)

	a := ScaleTo(d, []*page.DisplayOp{op}, 2, time.Millisecond, Linear)
	a.Step(t0)
	a.Step(t0.Add(time.Millisecond))

	// Scale 2 about the center (20,15) outside a 5-point translate:
	// E = -20 + 2*5 = -10, F = -15 + 2*0 = -15.
	m := op.Transform()
	if !near(m.A, 2) || !near(m.E, -10) || !near(m.F, -15) {
		t.Fatalf("matrix = %+v", m)
	}
}

func TestRotateTurnsAboutCenter(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := RotateTo(d, []*page.DisplayOp{op}, 90, time.Millisecond, Linear)
	a.Step(t0)
	a.Step(t0.Add(time.Millisecond))

	m := op.Transform()

	cx, cy := m.Apply(20, 15)
	if !near(cx, 20) || !near(cy, 15) {
		t.Fatalf("center moved to %v,%v", cx, cy)
	}

	if !near(m.B, 1) || !near(m.C, -1) {
		t.Fatalf("matrix = %+v", m)
	}
}
