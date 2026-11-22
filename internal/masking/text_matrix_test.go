package masking

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestTextMatrixHandlesQuarterAndFullTurns(t *testing.T) {
	t.Parallel()

	cases := []struct {
		deg  float32
		want layout.Matrix2D
		ok   bool
	}{
		{360, layout.IdentityMatrix(), true},
		{-360, layout.IdentityMatrix(), true},
		{720, layout.IdentityMatrix(), true},
		{90, layout.Matrix2D{A: 0, B: 1, C: -1, D: 0}, true},
		{180, layout.Matrix2D{A: -1, B: 0, C: 0, D: -1}, true},
		{270, layout.Matrix2D{A: 0, B: -1, C: 1, D: 0}, true},
		{-90, layout.Matrix2D{A: 0, B: -1, C: 1, D: 0}, true},
		{45, layout.Matrix2D{}, false},
	}

	for _, tc := range cases {
		op := &layout.DisplayOp{
			Kind:      layout.DisplayOpText,
			X:         10,
			Y:         20,
			RotateDeg: tc.deg,
		}
		if got := ReplayableTextRotation(op); got != tc.ok {
			t.Fatalf("deg=%v: ReplayableTextRotation = %v, want %v", tc.deg, got, tc.ok)
		}
		m, ok := TextMatrix(op)
		if ok != tc.ok {
			t.Fatalf("deg=%v: TextMatrix ok = %v, want %v", tc.deg, ok, tc.ok)
		}
		if !ok {
			continue
		}
		if m.A != tc.want.A || m.B != tc.want.B || m.C != tc.want.C || m.D != tc.want.D {
			t.Fatalf("deg=%v: linear = (%v,%v,%v,%v), want (%v,%v,%v,%v)",
				tc.deg, m.A, m.B, m.C, m.D, tc.want.A, tc.want.B, tc.want.C, tc.want.D)
		}
		// Pivot (op.X, op.Y) must be invariant under TextMatrix.
		px, py := m.Apply(op.X, op.Y)
		if px != op.X || py != op.Y {
			t.Fatalf("deg=%v: pivot mapped to (%v,%v), want (%v,%v)", tc.deg, px, py, op.X, op.Y)
		}
	}
}
