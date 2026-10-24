package csstransition_test

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

// TestPresenceStaysVisibleUntilTheLeavingEnd covers the visibility
// interpolation rule: a leaving element keeps the visible endpoint for every
// interior sample and flips to the hidden endpoint at progress 1.
func TestPresenceStaysVisibleUntilTheLeavingEnd(t *testing.T) {
	t.Parallel()

	ops := []*layout.DisplayOp{{Alpha: 1}, {Alpha: 0}}
	rule, err := csstransition.NewPresence(ops, []float64{1, 0}, []float64{0, 0}, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !rule.Step(nil, .25) {
		t.Fatal("the first sample did not install the visible endpoint")
	}
	if ops[0].Alpha != 1 || ops[1].Alpha != 0 {
		t.Fatalf("visible alphas = %v/%v, want 1/0", ops[0].Alpha, ops[1].Alpha)
	}
	if rule.Step(nil, .9) {
		t.Fatal("an interior sample after the visible endpoint changed it")
	}
	rebuilt := []*layout.DisplayOp{{Alpha: 9}, {Alpha: 9}}
	if err := rule.Retarget(nil, rebuilt); err != nil {
		t.Fatal(err)
	}
	rule.Reapply(nil)
	if rebuilt[0].Alpha != 1 || rebuilt[1].Alpha != 0 {
		t.Fatalf("reapplied alphas = %v/%v, want the visible endpoint",
			rebuilt[0].Alpha, rebuilt[1].Alpha)
	}
	if !rule.Step(nil, 1) {
		t.Fatal("the end of the run did not install the hidden endpoint")
	}
	if rebuilt[0].Alpha != csstransition.PresenceHiddenAlpha ||
		rebuilt[1].Alpha != csstransition.PresenceHiddenAlpha {
		t.Fatalf("hidden alphas = %v/%v, want the hidden endpoint",
			rebuilt[0].Alpha, rebuilt[1].Alpha)
	}
}

// TestPresenceEntersAtTheStart covers the entering direction: hidden at 0 and
// visible at the first interior sample.
func TestPresenceEntersAtTheStart(t *testing.T) {
	t.Parallel()

	ops := []*layout.DisplayOp{{Alpha: 1}}
	rule, err := csstransition.NewPresence(ops, []float64{1}, []float64{0}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !rule.Step(nil, 0) {
		t.Fatal("the zero sample did not install the hidden endpoint")
	}
	if ops[0].Alpha != csstransition.PresenceHiddenAlpha {
		t.Fatalf("alpha at 0 = %v, want hidden", ops[0].Alpha)
	}
	if !rule.Step(nil, .01) {
		t.Fatal("the first interior sample did not install the visible endpoint")
	}
	if ops[0].Alpha != 1 {
		t.Fatalf("alpha after the start = %v, want visible 1", ops[0].Alpha)
	}
}

// TestPresenceRetargetRequiresTheSameOperationSet pins the fit rule.
func TestPresenceRetargetRequiresTheSameOperationSet(t *testing.T) {
	t.Parallel()

	ops := []*layout.DisplayOp{{Alpha: 1}, {Alpha: 1}}
	rule, err := csstransition.NewPresence(ops, []float64{1, 1}, []float64{0, 0}, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := rule.Retarget(nil, ops[:1]); err == nil {
		t.Fatal("a shortened target set was retargeted")
	}
	if err := rule.Retarget(nil, ops); err != nil {
		t.Fatalf("the same target set was refused: %v", err)
	}
}

// TestStackFlipsAtTheMidpoint covers the z-index rule: the integer and the
// paint order flip at 50% of the eased progress, and a sample on the same side
// reports no change.
func TestStackFlipsAtTheMidpoint(t *testing.T) {
	t.Parallel()

	display := &layout.Display{Ops: []layout.DisplayOp{{}, {}}}
	ops := []*layout.DisplayOp{&display.Ops[0], &display.Ops[1]}
	rule, err := csstransition.NewStack(ops,
		[]int{0, 0}, []int{5, 0},
		[]bool{false, false}, []bool{true, false},
		[]int{0, 1}, []int{1, 0})
	if err != nil {
		t.Fatal(err)
	}
	if !rule.Step(display, .25) {
		t.Fatal("the first sample did not install the old stacking")
	}
	if ops[0].ZIndexSet || ops[0].ZIndex != 0 {
		t.Fatalf("z-index at 25%% = %d/%v, want the old unset 0",
			ops[0].ZIndex, ops[0].ZIndexSet)
	}
	if len(display.Order) != 2 || display.Order[0] != 0 || display.Order[1] != 1 {
		t.Fatalf("order at 25%% = %v, want 0,1", display.Order)
	}
	if rule.Step(display, .49) {
		t.Fatal("a sample on the same side reported a change")
	}
	rebuilt := &layout.Display{Ops: []layout.DisplayOp{{}, {}}}
	if err := rule.Retarget(rebuilt, []*layout.DisplayOp{&rebuilt.Ops[0], &rebuilt.Ops[1]}); err != nil {
		t.Fatal(err)
	}
	rule.Reapply(rebuilt)
	if len(rebuilt.Order) != 2 || rebuilt.Order[0] != 0 || rebuilt.Order[1] != 1 {
		t.Fatalf("reapplied order = %v, want the old 0,1", rebuilt.Order)
	}
	if rebuilt.Ops[0].ZIndexSet {
		t.Fatal("reapply wrote the new z-index while the rule was on the old side")
	}
	if !rule.Step(display, .5) {
		t.Fatal("the midpoint did not install the new stacking")
	}
	if !rebuilt.Ops[0].ZIndexSet || rebuilt.Ops[0].ZIndex != 5 {
		t.Fatalf("z-index at the midpoint = %d/%v, want 5",
			rebuilt.Ops[0].ZIndex, rebuilt.Ops[0].ZIndexSet)
	}
	if len(display.Order) != 2 || display.Order[0] != 1 || display.Order[1] != 0 {
		t.Fatalf("order at the midpoint = %v, want 1,0", display.Order)
	}
}

// TestStackRetargetRejectsAMismatchedOrder pins the order fit rule: a display
// whose operation count no longer matches the captured order cancels the run.
func TestStackRetargetRejectsAMismatchedOrder(t *testing.T) {
	t.Parallel()

	display := &layout.Display{Ops: []layout.DisplayOp{{}}}
	ops := []*layout.DisplayOp{&display.Ops[0]}
	rule, err := csstransition.NewStack(ops, []int{0}, []int{1},
		[]bool{false}, []bool{true}, []int{0}, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	short := &layout.Display{Ops: display.Ops[:0]}
	if err := rule.Retarget(short, nil); err == nil {
		t.Fatal("a shorter display was retargeted")
	}
}
