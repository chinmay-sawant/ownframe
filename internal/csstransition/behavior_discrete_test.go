package csstransition_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func TestTransitionBehaviorAllowDiscreteSwitchesPositionAtMidpoint(t *testing.T) {
	oldDisplay, oldBox := positionFrame(t, "allow-discrete", "static")
	newDisplay, newBox := positionFrame(t, "allow-discrete", "absolute")
	if len(newBox.Transitions) != 1 || newBox.Transitions[0].Behavior != "allow-discrete" {
		t.Fatalf("Blinkless transition metadata = %#v", newBox.Transitions)
	}
	indices := cssmotion.OperationIndices(newDisplay, newBox)
	from := cssmotion.Capture(oldDisplay, oldBox)
	ops := cssmotion.Resolve(newDisplay, indices)
	if len(indices) != 1 || len(from) != len(indices) || len(ops) != len(indices) {
		t.Fatalf("old/new paint counts = %d/%d", len(from), len(ops))
	}
	to := cssmotion.Read(ops[0])
	if from[0].X == to.X {
		t.Fatal("position change did not move the Blinkless paint operation")
	}
	start := time.Unix(1, 0)
	r, err := csstransition.Bind(newDisplay, newBox, "position", indices, from, start)
	if err != nil {
		t.Fatal(err)
	}
	op := ops[0]
	if !r.Step(start.Add(499*time.Millisecond)) || cssmotion.Read(op) != from[0] {
		t.Fatal("allow-discrete changed the operation before 50% progress")
	}
	if !r.Step(start.Add(500*time.Millisecond)) || cssmotion.Read(op) != to {
		t.Fatalf("midpoint paint = %+v, want destination %+v", cssmotion.Read(op), to)
	}
	if r.Step(start.Add(time.Second)) {
		t.Fatal("completed discrete transition remains active")
	}
}
