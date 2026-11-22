package masking

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

const (
	quarterTurnDeg      = 90
	halfTurnDeg         = 180
	threeQuarterTurnDeg = 270
	fullTurnDeg         = 360
)

// TextMatrix returns the map a text or bullet run paints through when it has
// no baked transform: the stored projection of a projected run, or the
// quarter-turn glyph rotation vertical writing emits. The matrix pivots on the
// run's baseline origin (op.X, op.Y) in points, the pivot the software painter
// rotates glyphs around, so replay, replay bounds, and dirty bounds all place
// the ink where a full software paint would.
//
// ok is false when the run paints upright with no projection, and when it
// carries a rotation the display list does not replay (see
// ReplayableTextRotation); such a run draws unrotated here and the page keeps
// the bitmap fallback for it.
func TextMatrix(op *layout.DisplayOp) (layout.Matrix2D, bool) {
	if op == nil || op.XformSet || (op.Kind != layout.DisplayOpText && op.Kind != layout.DisplayOpBullet) {
		return layout.Matrix2D{}, false
	}
	m := op.XformValue()
	projected := m != (layout.Matrix2D{}) && !m.IsIdentity()
	if op.RotateDeg != 0 {
		if projected || op.FakeOblique {
			return layout.Matrix2D{}, false
		}
		var ok bool
		if m, ok = quarterTurn(op.RotateDeg); !ok {
			return layout.Matrix2D{}, false
		}
	} else if !projected {
		return layout.Matrix2D{}, false
	}
	m.E = op.X - m.A*op.X - m.C*op.Y
	m.F = op.Y - m.B*op.X - m.D*op.Y

	return m, true
}

// ReplayableTextRotation reports whether a text run's glyph rotation can
// replay on the display list: no rotation, or a whole quarter turn on an
// upright, unprojected run with no baked transform and no synthetic oblique.
// Any other rotation keeps the software painter.
func ReplayableTextRotation(op *layout.DisplayOp) bool {
	if op == nil || op.RotateDeg == 0 {
		return true
	}
	_, ok := TextMatrix(op)

	return ok
}

// quarterTurn returns the exact rotation matrix for a whole multiple of 90
// degrees, clockwise in the y-down canvas like the software painter's
// cos/sin matrix. Exact coefficients keep glyph pixels on the grid that a
// floating-point cos(90) would nudge off it.
func quarterTurn(deg float32) (layout.Matrix2D, bool) {
	turn := math.Mod(float64(deg), fullTurnDeg)
	if turn < 0 {
		turn += fullTurnDeg
	}
	if math.Mod(turn, quarterTurnDeg) != 0 {
		return layout.Matrix2D{}, false
	}
	switch turn {
	case quarterTurnDeg:
		return layout.Matrix2D{A: 0, B: 1, C: -1, D: 0}, true
	case halfTurnDeg:
		return layout.Matrix2D{A: -1, B: 0, C: 0, D: -1}, true
	case threeQuarterTurnDeg:
		return layout.Matrix2D{A: 0, B: -1, C: 1, D: 0}, true
	}

	return layout.Matrix2D{}, false
}
