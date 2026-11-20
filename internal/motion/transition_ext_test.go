package motion_test

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"testing"
	"time"
)

func TestTransitionBehaviorKeepsContinuousPaintInterpolation(t *testing.T) {
	oldDisplay, oldBox := colorFrame(t, "#000000")
	newDisplay, box := colorFrame(t, "#ffffff")
	if len(box.Transitions) != 1 || box.Transitions[0].Behavior != "normal" {
		t.Fatalf("Blinkless continuous transition metadata = %#v", box.Transitions)
	}
	indices := motion.OperationIndices(newDisplay, box)
	from := motion.Capture(oldDisplay, oldBox)
	if len(indices) != 1 || len(from) != len(indices) {
		t.Fatalf("old/new continuous paint counts = %d/%d", len(from), len(indices))
	}
	start := time.Unix(3, 0)
	r, err := motion.BindTransition(newDisplay, box, "background-color", indices, from, start)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Step(start.Add(500 * time.Millisecond)) {
		t.Fatal("continuous background-color transition stopped at its midpoint")
	}
	got := motion.Read(&newDisplay.Ops[indices[0]])
	if got.R <= 0 || got.R >= 1 || got.G <= 0 || got.G >= 1 || got.B <= 0 || got.B >= 1 {
		t.Fatalf("continuous midpoint paint = %+v, want interpolated channels", got)
	}
}

func TestTransitionBehaviorAllowDiscreteSwitchesPositionAtMidpoint(t *testing.T) {
	oldDisplay, oldBox := positionFrame(t, "allow-discrete", "static")
	newDisplay, newBox := positionFrame(t, "allow-discrete", "absolute")
	if len(newBox.Transitions) != 1 || newBox.Transitions[0].Behavior != "allow-discrete" {
		t.Fatalf("Blinkless transition metadata = %#v", newBox.Transitions)
	}
	indices := motion.OperationIndices(newDisplay, newBox)
	from := motion.Capture(oldDisplay, oldBox)
	ops := motion.Resolve(newDisplay, indices)
	if len(indices) != 1 || len(from) != len(indices) || len(ops) != len(indices) {
		t.Fatalf("old/new paint counts = %d/%d", len(from), len(ops))
	}
	to := motion.Read(ops[0])
	if from[0].X == to.X {
		t.Fatal("position change did not move the Blinkless paint operation")
	}
	start := time.Unix(1, 0)
	r, err := motion.BindTransition(newDisplay, newBox, "position", indices, from, start)
	if err != nil {
		t.Fatal(err)
	}
	op := ops[0]
	if !r.Step(start.Add(499*time.Millisecond)) || motion.Read(op) != from[0] {
		t.Fatal("allow-discrete changed the operation before 50% progress")
	}
	if !r.Step(start.Add(500*time.Millisecond)) || motion.Read(op) != to {
		t.Fatalf("midpoint paint = %+v, want destination %+v", motion.Read(op), to)
	}
	if r.Step(start.Add(time.Second)) {
		t.Fatal("completed discrete transition remains active")
	}
}

func positionFrame(t *testing.T, behavior, position string) (*ownframe.Display, ownframe.Box) {
	t.Helper()
	html := fmt.Sprintf(`<style>
html,body{margin:0}#target{width:50px;height:20px;background:#f00;
position:%s;left:80px;top:70px;transition-property:position;
transition-duration:1000ms;transition-timing-function:linear;
transition-behavior:%s}
</style><div id="target"></div>`, position, behavior)
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return p.Probe().Display(), box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return nil, ownframe.Box{}
}

func colorFrame(t *testing.T, color string) (*ownframe.Display, ownframe.Box) {
	t.Helper()
	html := `<style>#target{width:50px;height:20px;background-color:` + color + `;
transition-property:background-color;transition-duration:1000ms;
transition-timing-function:linear;transition-behavior:normal}</style><div id="target"></div>`
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return p.Probe().Display(), box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return nil, ownframe.Box{}
}

func TestTransitionBehaviorNormalSkipsDiscretePositionTransition(t *testing.T) {
	oldDisplay, oldBox := positionFrame(t, "normal", "static")
	newDisplay, newBox := positionFrame(t, "normal", "absolute")
	if len(newBox.Transitions) != 1 || newBox.Transitions[0].Behavior != "normal" {
		t.Fatalf("Blinkless transition metadata = %#v", newBox.Transitions)
	}
	indices := motion.OperationIndices(newDisplay, newBox)
	from := motion.Capture(oldDisplay, oldBox)
	if len(indices) != 1 || len(from) != 1 {
		t.Fatalf("old/new position paint counts = %d/%d", len(from), len(indices))
	}
	r, err := motion.BindTransition(newDisplay, newBox, "position", indices, from, time.Unix(2, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.Step(time.Unix(2, 0).Add(500 * time.Millisecond)) {
		t.Fatal("normal behavior kept a discrete transition active")
	}
	if got, want := motion.Read(&newDisplay.Ops[indices[0]]), motion.Capture(newDisplay, newBox)[0]; got != want {
		t.Fatalf("normal behavior paint = %+v, want destination %+v", got, want)
	}
}

func TestVisibilityTransitionHasNoMatchedPaintOperation(t *testing.T) {
	visible := visibilityFrame(t, "visible")
	hidden := visibilityFrame(t, "hidden")
	visibleBox := targetBox(t, visible)
	box := targetBox(t, hidden)
	if len(box.Transitions) != 1 || box.Transitions[0].Behavior != "allow-discrete" {
		t.Fatalf("Blinkless visibility metadata = %#v", box.Transitions)
	}
	from := motion.Capture(visible.Probe().Display(), visibleBox)
	indices := motion.OperationIndices(hidden.Probe().Display(), box)
	if len(from) == 0 || len(indices) != 0 {
		t.Fatalf("visible/hidden paint counts = %d/%d", len(from), len(indices))
	}
	if _, err := motion.BindTransition(hidden.Probe().Display(), box, "visibility", indices, from, time.Time{}); err == nil {
		t.Fatal("visibility transition bound without a matched paint operation")
	}
}

func visibilityFrame(t *testing.T, visibility string) *ownframe.Page {
	t.Helper()
	html := `<style>#target{width:40px;height:20px;visibility:` + visibility + `;
transition-property:visibility;transition-duration:1s;transition-behavior:allow-discrete}</style><div id="target">visible text</div>`
	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 100, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

func targetBox(t *testing.T, p *ownframe.Page) ownframe.Box {
	t.Helper()
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			return box
		}
	}
	t.Fatal("Blinkless display has no target box")
	return ownframe.Box{}
}

// colorCaptures returns a color transition's display and box, the captured
// indices, the start paint of the old display, and the destination paint of
// the new one.
func colorCaptures(t *testing.T) (*ownframe.Display, ownframe.Box, []int,
	[]motion.Paint, []motion.Paint) {
	t.Helper()

	oldDisplay, oldBox := colorFrame(t, "#000000")
	newDisplay, box := colorFrame(t, "#ffffff")

	return newDisplay, box, motion.OperationIndices(newDisplay, box),
		motion.Capture(oldDisplay, oldBox), motion.Capture(newDisplay, box)
}

// TestBindToKeepsTheSizeAfterAResize checks that a rebind after the layout
// resized the operation keeps the operation's current size instead of
// writing the captured size back.
func TestBindToKeepsTheSizeAfterAResize(t *testing.T) {
	display, box, indices, from, to := colorCaptures(t)
	ops := motion.Resolve(display, indices)

	resized := to[0]
	resized.W = 120
	motion.Write(ops[0], resized)

	start := time.Unix(8, 0)
	r, err := motion.BindTransitionTo(display, box, "background-color", indices, from, to, start)
	if err != nil {
		t.Fatal(err)
	}
	r.Step(start.Add(time.Second))

	if got := motion.Read(ops[0]); got.W != 120 {
		t.Fatalf("the rebind wrote the captured width back: %+v", got)
	}
}

// TestBindToAnimatesToTheCapturedDestination checks that a rebind with an
// explicit destination animates toward the captured paint. The operation
// carries a mid-transition value, so the destination is the captured paint
// and not the value the display holds.
func TestBindToAnimatesToTheCapturedDestination(t *testing.T) {
	display, box, indices, from, to := colorCaptures(t)
	if len(indices) != 1 || len(from) != 1 || len(to) != 1 {
		t.Fatalf("paint counts = %d/%d/%d", len(indices), len(from), len(to))
	}

	ops := motion.Resolve(display, indices)
	motion.Write(ops[0], motion.Blend(from[0], to[0], .5))

	start := time.Unix(5, 0)
	r, err := motion.BindTransitionTo(display, box, "background-color", indices, from, to, start)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Step(start.Add(500 * time.Millisecond)) {
		t.Fatal("the transition stopped at its midpoint")
	}

	mid := motion.Read(ops[0])
	if mid.R <= 0 || mid.R >= 1 {
		t.Fatalf("midpoint paint = %+v, want an interpolated channel", mid)
	}

	r.Step(start.Add(time.Second))
	if got := motion.Read(ops[0]); got != to[0] {
		t.Fatalf("final paint = %+v, want the captured destination %+v", got, to[0])
	}
}

// TestBindToRefusesAMismatchedDestination checks that a destination paint
// count that does not match the operations is refused.
func TestBindToRefusesAMismatchedDestination(t *testing.T) {
	display, box, indices, from, _ := colorCaptures(t)
	final := append(append([]motion.Paint{}, from...), from...)
	if _, err := motion.BindTransitionTo(display, box, "background-color", indices, from, final, time.Unix(6, 0)); err == nil {
		t.Fatal("a destination paint count that does not match was accepted")
	}
}

// twoElementFrame lays out a target element and a second element, so a
// caller can point the captured indices at another element's operation.
func twoElementFrame(t *testing.T) (*ownframe.Display, ownframe.Box, ownframe.Box) {
	t.Helper()

	html := `<style>#target{width:50px;height:20px;background:#f00;
transition-property:background-color;transition-duration:1000ms}</style>
<div id="target"></div><div id="other" style="width:30px;height:10px;background:#0f0"></div>`

	p, err := ownframe.New(ownframe.Config{HTML: html, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	var target, other ownframe.Box
	for _, box := range p.Boxes() {
		switch box.ID {
		case "target":
			target = box
		case "other":
			other = box
		}
	}
	if target.ID == "" || other.ID == "" {
		t.Fatal("the display has no target or other box")
	}

	return p.Probe().Display(), target, other
}

// TestBindToRefusesOperationsOutsideTheBox checks that a rebind whose
// indices resolve to another element's operations is refused.
func TestBindToRefusesOperationsOutsideTheBox(t *testing.T) {
	display, target, other := twoElementFrame(t)
	indices := motion.OperationIndices(display, other)
	from := motion.Capture(display, other)
	to := motion.Capture(display, other)
	if len(indices) == 0 {
		t.Fatal("the other element has no operation")
	}

	_, err := motion.BindTransitionTo(display, target, "background-color", indices, from, to, time.Unix(7, 0))
	if err == nil {
		t.Fatal("a rebind onto another element's operations was accepted")
	}
}

// TestPresenceStaysVisibleUntilTheLeavingEnd covers the visibility
// interpolation rule: a leaving element keeps the visible endpoint for every
// interior sample and flips to the hidden endpoint at progress 1.
func TestPresenceStaysVisibleUntilTheLeavingEnd(t *testing.T) {
	t.Parallel()

	ops := []*layout.DisplayOp{{Alpha: 1}, {Alpha: 0}}
	rule, err := motion.NewPresence(ops, []float64{1, 0}, []float64{0, 0}, false, false)
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
	if rebuilt[0].Alpha != motion.PresenceHiddenAlpha ||
		rebuilt[1].Alpha != motion.PresenceHiddenAlpha {
		t.Fatalf("hidden alphas = %v/%v, want the hidden endpoint",
			rebuilt[0].Alpha, rebuilt[1].Alpha)
	}
}

// TestPresenceEntersAtTheStart covers the entering direction: hidden at 0 and
// visible at the first interior sample.
func TestPresenceEntersAtTheStart(t *testing.T) {
	t.Parallel()

	ops := []*layout.DisplayOp{{Alpha: 1}}
	rule, err := motion.NewPresence(ops, []float64{1}, []float64{0}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !rule.Step(nil, 0) {
		t.Fatal("the zero sample did not install the hidden endpoint")
	}
	if ops[0].Alpha != motion.PresenceHiddenAlpha {
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
	rule, err := motion.NewPresence(ops, []float64{1, 1}, []float64{0, 0}, false, false)
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
	rule, err := motion.NewStack(ops,
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
	rule, err := motion.NewStack(ops, []int{0}, []int{1},
		[]bool{false}, []bool{true}, []int{0}, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	short := &layout.Display{Ops: display.Ops[:0]}
	if err := rule.Retarget(short, nil); err == nil {
		t.Fatal("a shorter display was retargeted")
	}
}
