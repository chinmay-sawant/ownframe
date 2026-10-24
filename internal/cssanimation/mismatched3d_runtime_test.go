package cssanimation_test

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestMismatchedThreeDimensionalKeyframesBindAndRestamp is the regression for
// the former bind error: rotateX(0) to rotateY(90) has different function
// lists, and both are 3D. CSS interpolates that pair through the four-by-four
// decomposition, so the bind must succeed and the sample must reach the
// engine's restamp pass instead of failing the page.
func TestMismatchedThreeDimensionalKeyframesBindAndRestamp(t *testing.T) {
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`animation-name:turn;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:rotateX(0deg)}to{transform:rotateY(90deg)}}`)

	player, indices, base, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if got := cssmotion.Read(&animated.display.Ops[indices[0]]).Matrix; got != base[0].Matrix {
		t.Fatalf("mismatched 3D keyframe wrote matrix %+v, want the base %+v", got, base[0].Matrix)
	}
	if !player.RestampPending() {
		t.Fatal("a mismatched 3D keyframe did not reach the restamp pass")
	}
	values := player.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if !strings.HasPrefix(values[0].Transform, "matrix3d(") {
		t.Fatalf("restamp transform = %q, want matrix3d()", values[0].Transform)
	}
}

// TestMismatchedMatrix3DKeyframesBindAndRestamp covers the matrix3d pair, which
// CSS also converts to four-by-four matrices before interpolating.
func TestMismatchedMatrix3DKeyframesBindAndRestamp(t *testing.T) {
	const identity = `matrix3d(1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1)`
	const quarterTurn = `matrix3d(0,0,-1,0,0,1,0,0,1,0,0,0,0,0,0,1)` // rotateY(90deg)
	animated := layoutTarget(t, `<div id="target"></div>`,
		targetRules+`animation-name:turn;animation-duration:1s;animation-fill-mode:both;`+
			`animation-timing-function:linear;}`+
			`@keyframes turn{from{transform:`+identity+`}to{transform:`+quarterTurn+`}}`)

	player, _, _, start := bindMatrixTarget(t, animated)
	player.Step(start.Add(500 * time.Millisecond))

	if !player.RestampPending() {
		t.Fatal("a mismatched matrix3d pair did not reach the restamp pass")
	}
}
