package motion

import (
	"fmt"
	"image"
	"math"
	"slices"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// The reasons a property change can be reported as blocked. They name the
// missing capability, so Probe().Motion explains a snap instead of pretending
// a transition ran.
const (
	displayEntryReason = "display: an entering element needs the retained display path (phase 04)"
	displayExitReason  = "display: an exiting element keeps its pixels in a layer (phase 04)"
	orderReason        = "order: the engine has no order override, so the integer snaps"
	gridLineReason     = "grid line: the engine has no grid placement override, so the integer snaps"
)

// presenceExitReason names an exiting visibility or content-visibility change
// whose new display kept no operation to fade. Without an outgoing layer the
// run cannot keep the old pixels on screen.
func presenceExitReason(property string) string {
	return fmt.Sprintf("%s: an exiting presence change has no destination operation; "+
		"the outgoing layer is phase 04", property)
}

// discreteCapture is one non-interpolated style change. rule installs the
// endpoint the eased progress selects; targetKey identifies the same
// destination across an unrelated publish so the clock survives it.
type discreteCapture struct {
	rule      csstransition.Discrete
	targetKey string
	indices   []int
}

// discreteEntry is one live non-interpolated run.
type discreteEntry struct {
	key       transitionKey
	rule      csstransition.Discrete
	targetKey string
	record    layout.Transition
	start     time.Time
	// wrote reports that the from side is on screen. Until the delay ends the
	// sample pass parks instead of reinstalling it.
	wrote   bool
	sampled bool
	active  bool
}

// transitionDiscreteProperty reports whether a property needs the dedicated
// discrete rules instead of the paint renderer. Position stays on the paint
// path, where the existing halfway X/Y flip lives.
func transitionDiscreteProperty(property string) bool {
	switch property {
	case "visibility", "content-visibility", "z-index":
		return true
	}

	return false
}

// transitionReportProperty reports a property whose change the capture can
// detect but cannot transition yet. It produces no renderer; the change snaps
// and the reason lands in Stats.
func transitionReportProperty(property string) bool {
	switch property {
	case "display", "order",
		"grid-column-start", "grid-column-end",
		"grid-row-start", "grid-row-end":
		return true
	}

	return false
}

// captureDiscreteChange builds the rule for a presence or stacking change. ok
// is false when the property did not change, when the record does not opt in
// with allow-discrete, or when the display holds no operation the rule could
// write.
func (r *Runtime) captureDiscreteChange(previous *Display, old boxRef, next *Display,
	current boxRef, property string, record layout.Transition) (TransitionCapture, bool) {
	if record.DurationMS <= 0 || record.Behavior != "allow-discrete" {
		return TransitionCapture{}, false
	}

	switch property {
	case "visibility", "content-visibility":
		return r.capturePresence(previous, old, next, current, property)
	case "z-index":
		return r.captureStack(previous, old, next, current, property)
	}

	return TransitionCapture{}, false
}

// capturePresence builds a visibility or content-visibility run from the
// operation sets the two displays carry. A hidden element loses its text and
// image operations, so their count is the only presence signal the display
// exposes. An exiting run whose new display kept nothing cannot be painted
// without an outgoing layer and is counted as blocked.
func (r *Runtime) capturePresence(previous *Display, old boxRef, next *Display,
	current boxRef, property string) (TransitionCapture, bool) {
	oldIndices := cssmotion.Select(previous, old.box, old.index, cssmotion.KindSubtree)
	newIndices := cssmotion.Select(next, current.box, current.index, cssmotion.KindSubtree)
	if len(oldIndices) == len(newIndices) {
		return TransitionCapture{}, false
	}
	if len(newIndices) == 0 {
		r.transitionCapture.block(presenceExitReason(property))

		return TransitionCapture{}, false
	}

	entering := len(newIndices) > len(oldIndices)
	ops := make([]*layout.DisplayOp, len(newIndices))
	visible := make([]float64, len(newIndices))
	hidden := make([]float64, len(newIndices))
	for i, index := range newIndices {
		op := &next.Ops[index]
		ops[i] = op
		hidden[i] = csstransition.PresenceHiddenAlpha
		visible[i] = op.Alpha
		if !entering && i < len(oldIndices) {
			// An exiting run keeps the alpha that is on screen, which is the
			// old display's value at the same position in the element's
			// operation list. A redraw renumbers operation IDs, so position is
			// the only stable correspondence the display exposes.
			visible[i] = previous.Ops[oldIndices[i]].Alpha
		}
	}

	rule, err := csstransition.NewPresence(ops, visible, hidden, entering,
		property == "content-visibility")
	if err != nil {
		r.stats.TransitionError = err.Error()
		r.transitionCapture.skipped++

		return TransitionCapture{}, false
	}

	return TransitionCapture{
		kind:    transitionDiscrete,
		indices: newIndices,
		discrete: &discreteCapture{
			rule:      rule,
			targetKey: presenceTargetKey(property, entering, visible, hidden),
			indices:   newIndices,
		},
	}, true
}

// captureStack builds a z-index run. The operation set must stay the same, so
// the z values are compared by position; the paint order is captured alongside
// them and flipped at the midpoint.
func (r *Runtime) captureStack(previous *Display, old boxRef, next *Display,
	current boxRef, property string) (TransitionCapture, bool) {
	oldIndices := cssmotion.Select(previous, old.box, old.index, cssmotion.KindSubtree)
	newIndices := cssmotion.Select(next, current.box, current.index, cssmotion.KindSubtree)
	if len(oldIndices) != len(newIndices) ||
		!sameOpKinds(previous, oldIndices, next, newIndices) {
		return TransitionCapture{}, false
	}

	fromZ := make([]int, len(newIndices))
	toZ := make([]int, len(newIndices))
	fromSet := make([]bool, len(newIndices))
	toSet := make([]bool, len(newIndices))
	ops := make([]*layout.DisplayOp, len(newIndices))
	changed := !slices.Equal(previous.Order, next.Order)
	for i, index := range newIndices {
		oldOp := &previous.Ops[oldIndices[i]]
		op := &next.Ops[index]
		ops[i] = op
		fromZ[i], fromSet[i] = oldOp.ZIndex, oldOp.ZIndexSet
		toZ[i], toSet[i] = op.ZIndex, op.ZIndexSet
		if fromZ[i] != toZ[i] || fromSet[i] != toSet[i] {
			changed = true
		}
	}
	if !changed {
		return TransitionCapture{}, false
	}

	fromOrder := slices.Clone(previous.Order)
	toOrder := slices.Clone(next.Order)
	rule, err := csstransition.NewStack(ops, fromZ, toZ, fromSet, toSet, fromOrder, toOrder)
	if err != nil {
		r.stats.TransitionError = err.Error()
		r.transitionCapture.skipped++

		return TransitionCapture{}, false
	}

	return TransitionCapture{
		kind:    transitionDiscrete,
		indices: newIndices,
		discrete: &discreteCapture{
			rule:      rule,
			targetKey: stackTargetKey(property, fromZ, toZ, fromSet, toSet, fromOrder, toOrder),
			indices:   newIndices,
		},
	}, true
}

// discreteTargetUnchanged reports that a live rule already heads for the
// captured destination. An unrelated publish keeps the clock instead of
// restarting the run from the side it reached.
func (r *Runtime) discreteTargetUnchanged(key transitionKey, capture *discreteCapture) bool {
	if capture == nil {
		return false
	}
	for i := range r.discrete {
		entry := &r.discrete[i]
		if entry.key == key {
			return entry.targetKey == capture.targetKey
		}
	}

	return false
}

// noteReportChange records a detected change of a property whose transition
// path is not implemented. It never builds a renderer.
func (r *Runtime) noteReportChange(previous *Display, old boxRef, next *Display,
	current boxRef, property string) {
	switch property {
	case "display":
		// Both boxes are in their displays, so display did not change.
	case "order":
		if siblingRank(previous, old.index) == siblingRank(next, current.index) {
			return
		}
		r.transitionCapture.block(orderReason)
	case "grid-column-start", "grid-column-end", "grid-row-start", "grid-row-end":
		if math.Abs(old.box.X-current.box.X) <= transitionPaintEpsilon &&
			math.Abs(old.box.Y-current.box.Y) <= transitionPaintEpsilon {
			return
		}
		r.transitionCapture.block(gridLineReason)
	}
}

// noteDisplayExits reports an element the new display dropped while it declared
// a display transition. The committed display no longer holds its operations,
// so keeping its pixels is the phase 04 layer's job.
func (r *Runtime) noteDisplayExits(previous, next *Display) {
	if previous == nil || next == nil {
		return
	}

	nextBoxes := indexBoxes(next.Boxes)
	for key, ref := range indexBoxes(previous.Boxes) {
		if _, ok := nextBoxes[key]; ok {
			continue
		}
		if _, ok := transitionRecords(ref.box)["display"]; ok {
			r.transitionCapture.block(displayExitReason)
		}
	}
}

// siblingRank counts the earlier boxes that share a parent. An order change
// reorders the box list, so a different rank is the display signal for it.
func siblingRank(d *Display, index int) int {
	if d == nil || index < 0 || index >= len(d.Boxes) {
		return -1
	}
	parent := d.Boxes[index].ParentIndex
	rank := 0
	for i := range d.Boxes {
		if i == index || d.Boxes[i].ParentIndex != parent {
			continue
		}
		if i < index {
			rank++
		}
	}

	return rank
}

// presenceTargetKey identifies one presence destination.
func presenceTargetKey(property string, entering bool, visible, hidden []float64) string {
	return fmt.Sprintf("%s:presence:%t:%v:%v", property, entering, visible, hidden)
}

// stackTargetKey identifies one stacking destination, including the paint order
// the run installs at the midpoint.
func stackTargetKey(property string, fromZ, toZ []int, fromSet, toSet []bool,
	fromOrder, toOrder []int) string {
	return fmt.Sprintf("%s:stack:%v:%v:%v:%v:%v:%v", property,
		fromZ, toZ, fromSet, toSet, fromOrder, toOrder)
}

// discreteBounds is the union of the operations a discrete rule writes.
func discreteBounds(d *Display, rule csstransition.Discrete) (box image.Rectangle) {
	if rule == nil {
		return box
	}
	for _, op := range rule.Ops() {
		if op == nil {
			continue
		}
		box = box.Union(dirty.OpBounds(op, d))
	}

	return box
}

// discreteActive reports whether a discrete run still needs a frame.
func (r *Runtime) discreteActive() bool {
	for i := range r.discrete {
		if !r.discrete[i].sampled || r.discrete[i].active {
			return true
		}
	}

	return false
}

// nextDiscreteFrame is the next time a discrete run needs a sample. A run
// inside its delay parks until the delay ends; a moving run asks for the next
// frame.
func (r *Runtime) nextDiscreteFrame(now time.Time) (time.Time, bool) {
	var wake time.Time
	found := false
	add := func(at time.Time) {
		if !found || at.Before(wake) {
			wake, found = at, true
		}
	}
	for i := range r.discrete {
		entry := &r.discrete[i]
		if !entry.sampled {
			add(now)

			continue
		}
		if !entry.active {
			continue
		}
		at := now
		if end := entry.start.Add(milliseconds(entry.record.DelayMS)); now.Before(end) {
			at = end
		}
		add(at)
	}

	return wake, found
}
