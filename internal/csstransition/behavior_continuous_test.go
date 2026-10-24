package csstransition_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func TestTransitionBehaviorKeepsContinuousPaintInterpolation(t *testing.T) {
	oldDisplay, oldBox := colorFrame(t, "#000000")
	newDisplay, box := colorFrame(t, "#ffffff")
	if len(box.Transitions) != 1 || box.Transitions[0].Behavior != "normal" {
		t.Fatalf("Blinkless continuous transition metadata = %#v", box.Transitions)
	}
	indices := cssmotion.OperationIndices(newDisplay, box)
	from := cssmotion.Capture(oldDisplay, oldBox)
	if len(indices) != 1 || len(from) != len(indices) {
		t.Fatalf("old/new continuous paint counts = %d/%d", len(from), len(indices))
	}
	start := time.Unix(3, 0)
	r, err := csstransition.Bind(newDisplay, box, "background-color", indices, from, start)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Step(start.Add(500 * time.Millisecond)) {
		t.Fatal("continuous background-color transition stopped at its midpoint")
	}
	got := cssmotion.Read(&newDisplay.Ops[indices[0]])
	if got.R <= 0 || got.R >= 1 || got.G <= 0 || got.G >= 1 || got.B <= 0 || got.B >= 1 {
		t.Fatalf("continuous midpoint paint = %+v, want interpolated channels", got)
	}
}
