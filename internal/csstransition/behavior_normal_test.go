package csstransition_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func TestTransitionBehaviorNormalSkipsDiscretePositionTransition(t *testing.T) {
	oldDisplay, oldBox := positionFrame(t, "normal", "static")
	newDisplay, newBox := positionFrame(t, "normal", "absolute")
	if len(newBox.Transitions) != 1 || newBox.Transitions[0].Behavior != "normal" {
		t.Fatalf("Blinkless transition metadata = %#v", newBox.Transitions)
	}
	indices := cssmotion.OperationIndices(newDisplay, newBox)
	from := cssmotion.Capture(oldDisplay, oldBox)
	if len(indices) != 1 || len(from) != 1 {
		t.Fatalf("old/new position paint counts = %d/%d", len(from), len(indices))
	}
	r, err := csstransition.Bind(newDisplay, newBox, "position", indices, from, time.Unix(2, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.Step(time.Unix(2, 0).Add(500 * time.Millisecond)) {
		t.Fatal("normal behavior kept a discrete transition active")
	}
	if got, want := cssmotion.Read(&newDisplay.Ops[indices[0]]), cssmotion.Capture(newDisplay, newBox)[0]; got != want {
		t.Fatalf("normal behavior paint = %+v, want destination %+v", got, want)
	}
}
