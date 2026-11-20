package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
	"math"
	"slices"
	"sort"
	"strconv"
	"time"
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
	rule      Discrete
	targetKey string
	indices   []int
}

// discreteEntry is one live non-interpolated run.
type discreteEntry struct {
	key       transitionKey
	rule      Discrete
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
	current boxRef, property string, record layout.Transition,
	prevIndex, nextIndex *Index) (TransitionCapture, bool) {
	if record.DurationMS <= 0 || record.Behavior != "allow-discrete" {
		return TransitionCapture{}, false
	}

	switch property {
	case "visibility", "content-visibility":
		return r.capturePresence(previous, old, next, current, property, prevIndex, nextIndex)
	case "z-index":
		return r.captureStack(previous, old, next, current, property, prevIndex, nextIndex)
	}

	return TransitionCapture{}, false
}

// capturePresence builds a visibility or content-visibility run from the
// operation sets the two displays carry. A hidden element loses its text and
// image operations, so their count is the only presence signal the display
// exposes. An exiting run whose new display kept nothing cannot be painted
// without an outgoing layer and is counted as blocked.
func (r *Runtime) capturePresence(previous *Display, old boxRef, next *Display,
	current boxRef, property string,
	prevIndex, nextIndex *Index) (TransitionCapture, bool) {
	oldIndices := prevIndex.Select(old.box, old.index, KindSubtree)
	newIndices := nextIndex.Select(current.box, current.index, KindSubtree)
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
		hidden[i] = PresenceHiddenAlpha
		visible[i] = op.Alpha
		if !entering && i < len(oldIndices) {
			// An exiting run keeps the alpha that is on screen, which is the
			// old display's value at the same position in the element's
			// operation list. A redraw renumbers operation IDs, so position is
			// the only stable correspondence the display exposes.
			visible[i] = previous.Ops[oldIndices[i]].Alpha
		}
	}

	rule, err := NewPresence(ops, visible, hidden, entering,
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
	current boxRef, property string,
	prevIndex, nextIndex *Index) (TransitionCapture, bool) {
	oldIndices := prevIndex.Select(old.box, old.index, KindSubtree)
	newIndices := nextIndex.Select(current.box, current.index, KindSubtree)
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
	rule, err := NewStack(ops, fromZ, toZ, fromSet, toSet, fromOrder, toOrder)
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
func (r *Runtime) noteDisplayExits(previous, next *Display, filter map[string]struct{}) {
	if previous == nil || next == nil {
		return
	}

	nextBoxes := indexBoxes(next.Boxes, filter)
	for key, ref := range indexBoxes(previous.Boxes, filter) {
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
func discreteBounds(d *Display, rule Discrete) (box image.Rectangle) {
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

// stackFlip is the eased progress where an integer property flips.
const stackFlip = 0.5

// Discrete is one non-interpolated transition. The runtime samples it with the
// eased progress of its record; Step installs the value that progress selects
// and reports whether the display changed. Ops returns the operations the rule
// writes, so the runtime can damage their painted area and re-resolve them
// after a display commit. Reapply installs the value the rule has already
// selected on a display a later publish built, so a redraw does not flash the
// committed endpoint for a frame.
type Discrete interface {
	Step(d *layout.Display, progress float64) bool
	Reapply(d *layout.Display)
	Ops() []*layout.DisplayOp
	Retarget(d *layout.Display, ops []*layout.DisplayOp) error
}

// errDiscreteTargets is returned when a display hands a rule a different
// operation set. A half-written target set looks worse than a snap, so the
// caller drops the run.
var errDiscreteTargets = errors.New("csstransition: discrete target set changed")

// presenceState is the endpoint a presence rule has installed.
type presenceState int8

const (
	presenceUnknown presenceState = iota
	presenceHidden
	presenceVisible
)

// PresenceHiddenAlpha is the alpha a presence rule writes for the hidden
// endpoint. The engine treats an operation alpha of zero as "no alpha scale",
// the same as one, so a literal zero would paint an opaque element. One
// thousandth is below the 8-bit paint floor and leaves the operation invisible.
const PresenceHiddenAlpha = 1e-3

// Presence is a visibility or content-visibility transition. Visibility keeps
// the element visible for the whole interval when either endpoint is visible:
// an element that enters is hidden at progress 0 and visible after it, and an
// element that leaves stays visible until progress 1 and is hidden at the end.
// Content-visibility is a plain discrete property and flips at the halfway
// point instead.
//
// The rule writes the operation alpha. A hidden element's own operations are
// the ones the new display kept, so the visible endpoint is the alpha the
// commit carries; the hidden endpoint is PresenceHiddenAlpha. Visible is
// indexed like Ops.
type Presence struct {
	ops      []*layout.DisplayOp
	visible  []float64
	hidden   []float64
	entering bool
	midpoint bool
	state    presenceState
}

// NewPresence builds a presence rule over the operations the new display kept.
// visible and hidden are the endpoint alphas per operation. entering reports
// that the element is becoming visible, which flips at the start instead of
// the end. midpoint selects the discrete 50% flip content-visibility uses
// instead of the visible-through-the-run rule visibility uses. A hidden
// endpoint of zero becomes PresenceHiddenAlpha, because the engine reads a
// zero operation alpha as opaque.
func NewPresence(ops []*layout.DisplayOp, visible, hidden []float64,
	entering, midpoint bool) (*Presence, error) {
	if len(ops) == 0 {
		return nil, errors.New("csstransition: presence needs at least one operation")
	}
	if len(visible) != len(ops) || len(hidden) != len(ops) {
		return nil, fmt.Errorf("csstransition: presence has %d operations and %d/%d alpha values",
			len(ops), len(visible), len(hidden))
	}
	hidden = append([]float64(nil), hidden...)
	for i := range hidden {
		if hidden[i] <= 0 {
			hidden[i] = PresenceHiddenAlpha
		}
	}

	return &Presence{
		ops:      ops,
		visible:  visible,
		hidden:   hidden,
		entering: entering,
		midpoint: midpoint,
	}, nil
}

// Step installs the endpoint the eased progress selects and reports whether it
// changed. A run that has already installed that endpoint does nothing.
func (p *Presence) Step(d *layout.Display, progress float64) bool {
	if p == nil || len(p.ops) == 0 {
		return false
	}

	next := presenceHidden
	if presenceShown(progress, p.entering, p.midpoint) {
		next = presenceVisible
	}
	if next == p.state {
		return false
	}
	p.state = next

	alpha := p.hidden
	if next == presenceVisible {
		alpha = p.visible
	}
	p.write(alpha)

	return true
}

// Reapply installs the endpoint the rule has already selected. A publish
// rebuilds the display from the declared styles, so a redraw in the middle of
// a presence run would otherwise flash the visible endpoint.
func (p *Presence) Reapply(_ *layout.Display) {
	if p == nil || p.state == presenceUnknown {
		return
	}
	alpha := p.hidden
	if p.state == presenceVisible {
		alpha = p.visible
	}
	p.write(alpha)
}

// write installs one endpoint alpha per operation.
func (p *Presence) write(alpha []float64) {
	for i, op := range p.ops {
		if op == nil {
			continue
		}
		op.Alpha = clampAlpha(alpha[i])
	}
}

// Ops returns the operations the rule writes.
func (p *Presence) Ops() []*layout.DisplayOp {
	if p == nil {
		return nil
	}

	return p.ops
}

// Retarget points the rule at the operations a later display published for the
// same element. The operation count must match, because the endpoint alphas
// are indexed by operation position.
func (p *Presence) Retarget(_ *layout.Display, ops []*layout.DisplayOp) error {
	if p == nil || len(ops) != len(p.ops) {
		return errDiscreteTargets
	}
	p.ops = ops

	return nil
}

// presenceShown reports whether a presence transition shows the element at the
// sampled progress. A midpoint rule uses the discrete 50% boundary. Otherwise
// the visible value holds for every interior sample; only the hidden endpoint
// is the value the run leaves or starts from.
func presenceShown(progress float64, entering, midpoint bool) bool {
	if midpoint {
		if entering {
			return progress >= stackFlip
		}

		return progress < stackFlip
	}
	if entering {
		return progress > 0
	}

	return progress < 1
}

// Stack is a z-index transition. The integer flips at 50% of the eased
// progress, and the display paint order flips with it, so an element that
// changes its stacking level paints above its neighbours on both sides of the
// midpoint rather than only after the run ends.
//
// FromOrder and ToOrder are the display paint orders captured at the commit.
// HasOrder is false when the two orders cannot describe the same operation
// set, in which case only the integer moves and the committed paint order
// stays.
type Stack struct {
	ops            []*layout.DisplayOp
	fromZ, toZ     []int
	fromSet, toSet []bool
	fromOrder      []int
	toOrder        []int
	hasOrder       bool
	side           bool
	wrote          bool
}

// NewStack builds a z-index rule. from/to report each operation's z-index and
// whether the value was set; fromOrder and toOrder are the old and new display
// paint orders. An order pair that does not cover the same length is ignored.
func NewStack(ops []*layout.DisplayOp, fromZ, toZ []int, fromSet, toSet []bool,
	fromOrder, toOrder []int) (*Stack, error) {
	if len(ops) == 0 {
		return nil, errors.New("csstransition: stack needs at least one operation")
	}
	if len(fromZ) != len(ops) || len(toZ) != len(ops) ||
		len(fromSet) != len(ops) || len(toSet) != len(ops) {
		return nil, fmt.Errorf("csstransition: stack target set does not match %d operations", len(ops))
	}

	return &Stack{
		ops:       ops,
		fromZ:     fromZ,
		toZ:       toZ,
		fromSet:   fromSet,
		toSet:     toSet,
		fromOrder: fromOrder,
		toOrder:   toOrder,
		hasOrder:  len(fromOrder) == len(toOrder) && len(fromOrder) > 0,
	}, nil
}

// Step installs the z-index and paint order the eased progress selects and
// reports whether it changed. The rule writes once per side, so a steady
// stacking level damages nothing between frames.
func (s *Stack) Step(d *layout.Display, progress float64) bool {
	if s == nil || len(s.ops) == 0 {
		return false
	}

	side := progress >= stackFlip
	if s.wrote && side == s.side {
		return false
	}
	s.side, s.wrote = side, true
	s.install(d)

	return true
}

// Reapply installs the stacking side the rule has already selected. A publish
// rebuilds the display from the declared styles, so a redraw in the middle of a
// z-index run would otherwise paint the new stacking for a frame.
func (s *Stack) Reapply(d *layout.Display) {
	if s == nil || !s.wrote {
		return
	}
	s.install(d)
}

// install writes the current z-index and paint order.
func (s *Stack) install(d *layout.Display) {
	values, set := s.fromZ, s.fromSet
	if s.side {
		values, set = s.toZ, s.toSet
	}
	for i, op := range s.ops {
		if op == nil {
			continue
		}
		op.ZIndex, op.ZIndexSet = values[i], set[i]
	}
	if s.hasOrder && d != nil {
		if s.side {
			d.Order = s.toOrder
		} else {
			d.Order = s.fromOrder
		}
	}
}

// Ops returns the operations the rule writes.
func (s *Stack) Ops() []*layout.DisplayOp {
	if s == nil {
		return nil
	}

	return s.ops
}

// Retarget points the rule at the operations and paint order a later display
// published. A different operation count or an order that no longer indexes
// the display drops the run, because the committed order is the only one the
// runtime can describe.
func (s *Stack) Retarget(d *layout.Display, ops []*layout.DisplayOp) error {
	if s == nil || len(ops) != len(s.ops) {
		return errDiscreteTargets
	}
	if s.hasOrder && (!orderIndexes(d, s.fromOrder) || !orderIndexes(d, s.toOrder)) {
		return errDiscreteTargets
	}
	s.ops = ops

	return nil
}

// orderIndexes reports whether every index in order addresses an operation of
// d and the order covers the whole list.
func orderIndexes(d *layout.Display, order []int) bool {
	if d == nil || len(order) != len(d.Ops) {
		return false
	}
	for _, index := range order {
		if index < 0 || index >= len(d.Ops) {
			return false
		}
	}

	return true
}

// clampAlpha keeps a presence endpoint in the range the painters accept. NaN
// reads as opaque, the way the renderer treats it.
func clampAlpha(value float64) float64 {
	if math.IsNaN(value) {
		return 1
	}
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}

	return value
}

// discretePaintProperty reports the discrete properties that still ride the
// paint renderer. Position moves the operations, so it flips the captured X/Y
// at the halfway point; presence and stacking need their own rules above.
func discretePaintProperty(property string) bool {
	return property == "position"
}

// geometryEpsilon is the change in a CSS length at which a geometry transition
// is worth another engine relayout. A sub-pixel sample keeps the previous
// request, so a slow run does not spend the per-frame budget on a value the
// display cannot show.
const geometryEpsilon = 1e-3

// geometryCapture is one layout length at the two commits. The value is the
// element's border-box extent for width and height, and the inset from the
// parent box origin for top and left.
type geometryCapture struct {
	from, to float64
}

// geometryEntry is one live layout-length transition. It shares the relayout
// budget and element queue with the layout keyframe driver, so a run reaches
// the engine's override path at most once per frame.
type geometryEntry struct {
	key    transitionKey
	from   float64
	to     float64
	record layout.Transition
	start  time.Time
	// animated is the layout-tier properties the box's animations declare at
	// the last bind or adoption. A property in this set wins the cascade, so
	// the run is dropped without reparsing the keyframes every frame.
	animated map[string]struct{}
	// reversingStart and reversingFactor are the CSS reversal state, kept for
	// an interrupt that returns to where the run started.
	reversingStart  float64
	reversingFactor float64
	// applied is the value the last emitted request used, so a run that does
	// not move this frame emits nothing.
	applied    float64
	hasApplied bool
	// sampled reports that the run has reached the engine at least once.
	sampled bool
	// finished marks the entry whose final destination has been emitted. It
	// stays until the next layout pass can prune it, so a failed apply can be
	// retried by ReturnLayout.
	finished bool
}

// transitionGeometryProperty reports whether a property is a layout length the
// display exposes through a box. The display carries the border box, not the
// used style, so width and height are exact for a box with no padding or
// border and the inset is measured from the parent border-box origin.
func transitionGeometryProperty(property string) bool {
	switch property {
	case "width", "height", "top", "left":
		return true
	}

	return false
}

// geometryAnimationNames returns the layout-tier keyframe properties that write
// the same used value a geometry transition would. An animation on any of them
// wins the cascade, so the transition never binds.
func geometryAnimationNames(property string) []string {
	switch property {
	case "width":
		return []string{"width", "inline-size"}
	case "height":
		return []string{"height", "block-size"}
	case "top":
		return []string{"top", "inset", "inset-block-start"}
	case "left":
		return []string{"left", "inset", "inset-inline-start"}
	}

	return nil
}

// layoutAnimatedProperties collects the layout-tier properties the box's
// animations declare. The keyframe parser is the same one the layout binder
// uses, so a property it accepts is the property the player will sample.
func layoutAnimatedProperties(d *Display, box layout.Box) map[string]struct{} {
	if d == nil || len(box.Animations) == 0 {
		return nil
	}

	var out map[string]struct{}
	for i := range box.Animations {
		animation := box.Animations[i]
		if animation.Name == "" || animation.Name == "none" {
			continue
		}
		frames, err := LayoutKeyframes(d.Keyframes, animation.Name)
		if err != nil {
			continue
		}
		for f := range frames {
			for v := range frames[f].Layout {
				if out == nil {
					out = make(map[string]struct{})
				}
				out[frames[f].Layout[v].Property] = struct{}{}
			}
		}
	}

	return out
}

// layoutAnimationOwns reports whether any of the layout properties that write
// the same used value is animated on the box.
func layoutAnimationOwns(animated map[string]struct{}, property string) bool {
	if len(animated) == 0 {
		return false
	}
	for _, name := range geometryAnimationNames(property) {
		if _, ok := animated[name]; ok {
			return true
		}
	}

	return false
}

// animationOwnsReason names the precedence skip for one geometry property.
func animationOwnsReason(property string) string {
	return fmt.Sprintf("%s: an animation owns the property, so the transition is skipped", property)
}

// duplicateIDReason names a geometry change the engine cannot address because
// two boxes share the element id.
func duplicateIDReason(property string) string {
	return fmt.Sprintf("%s: a repeated element id cannot be relaid out", property)
}

// captureGeometryChange builds one layout-length capture from the two boxes.
// changed is false when the used length is the same on both displays, which the
// caller still uses to cancel a run heading elsewhere.
//
// Width and height use the border box directly. Top and left are expressed
// against the new box's parent origin, so a parent that moves in the same
// commit still interpolates the element's visual position from the old box to
// the new one.
func captureGeometryChange(old boxRef, next *Display, current boxRef,
	property string) (TransitionCapture, bool) {
	from, to := geometryEndpoints(old, next, current, property)
	capture := TransitionCapture{
		kind:     transitionGeometry,
		geometry: geometryCapture{from: from, to: to},
	}

	return capture, math.Abs(to-from) > transitionPaintEpsilon
}

// geometryEndpoints returns the two lengths one property interpolates.
func geometryEndpoints(old boxRef, next *Display, current boxRef,
	property string) (from, to float64) {
	switch property {
	case "width":
		return old.box.W, current.box.W
	case "height":
		return old.box.H, current.box.H
	case "top":
		origin := geometryOrigin(next, current.box, true)

		return old.box.Y - origin, current.box.Y - origin
	case "left":
		origin := geometryOrigin(next, current.box, false)

		return old.box.X - origin, current.box.X - origin
	}

	return 0, 0
}

// geometryOrigin returns the parent border-box origin a top or left override is
// expressed against. A top-level box uses the page origin. The engine's used
// style is not in the display, so a parent padding or border is not part of the
// reference; the approximation is reported in the phase notes.
func geometryOrigin(d *Display, box layout.Box, vertical bool) float64 {
	if d == nil || box.ParentIndex < 0 || box.ParentIndex >= len(d.Boxes) {
		return 0
	}
	parent := d.Boxes[box.ParentIndex]
	if vertical {
		return parent.Y
	}

	return parent.X
}

// liveGeometry returns the running layout-length entry for one element and
// property, or nil.
func (r *Runtime) liveGeometry(key transitionKey) *geometryEntry {
	for i := range r.geometry {
		if r.geometry[i].key == key {
			return &r.geometry[i]
		}
	}

	return nil
}

// geometryTargetUnchanged reports that a live run already heads for the
// captured destination, so an unrelated publish must keep its clock.
func (r *Runtime) geometryTargetUnchanged(key transitionKey, to float64) bool {
	entry := r.liveGeometry(key)

	return entry != nil && math.Abs(entry.to-to) <= transitionPaintEpsilon
}

// syncGeometryQueue adds every live geometry element that is not already in
// the shared layout queue. The queue is rebuilt from the registry instead of
// requeued inside the budget loop, so one element cannot be sampled twice in
// one frame.
func (r *Runtime) syncGeometryQueue() {
	if len(r.geometry) == 0 {
		return
	}

	ids := make([]string, 0, len(r.geometry))
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.finished || entry.key.element.id == "" {
			continue
		}
		ids = append(ids, entry.key.element.id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		r.layout.enqueue(id)
	}
}

// pruneGeometry drops the entries whose final destination already reached the
// engine. A ReturnLayout on the same frame clears the finished mark first, so a
// refused final patch is not lost.
func (r *Runtime) pruneGeometry() {
	if len(r.geometry) == 0 {
		return
	}

	kept := r.geometry[:0]
	for i := range r.geometry {
		if r.geometry[i].finished {
			continue
		}
		kept = append(kept, r.geometry[i])
	}
	r.geometry = kept
}

// returnGeometry clears the finished mark for one element after a request the
// page could not apply, so the next layout pass emits its final value again.
func (r *Runtime) returnGeometry(id string) {
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.key.element.id == id && entry.finished {
			entry.finished = false
		}
	}
}

// geometryOverrides samples every live layout-length run on one element and
// appends the override list the engine needs.
func (r *Runtime) geometryOverrides(id string, now time.Time,
	buf []LayoutOverride) []LayoutOverride {
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.key.element.id != id || entry.finished {
			continue
		}
		if layoutAnimationOwns(entry.animated, entry.key.property) {
			// The animation writes this used value later in the frame. Drop the
			// run rather than sampling a value that is overwritten.
			entry.finished = true
			r.stats.TransitionsCancelled++
			continue
		}

		progress, active := ProgressFor(entry.record, entry.start, now)
		value := entry.to
		if active {
			value = entry.from + (entry.to-entry.from)*progress
		}
		entry.sampled = true
		if !active {
			entry.finished = true
		}
		if entry.hasApplied && math.Abs(value-entry.applied) <= geometryEpsilon {
			continue
		}
		entry.applied, entry.hasApplied = value, true
		buf = append(buf, LayoutOverride{
			Property: entry.key.property,
			Value:    cssPixels(value),
		})
	}

	return buf
}

// geometryActive reports whether a layout-length run still needs a frame.
func (r *Runtime) geometryActive() bool {
	for i := range r.geometry {
		if !r.geometry[i].finished {
			return true
		}
	}

	return false
}

// nextGeometryFrame is the next time a layout-length run needs the engine. A
// run inside its delay parks until the delay ends; a moving run asks for the
// next frame.
func (r *Runtime) nextGeometryFrame(now time.Time) (time.Time, bool) {
	var wake time.Time
	found := false
	for i := range r.geometry {
		entry := &r.geometry[i]
		if entry.finished {
			continue
		}
		at := now
		if end := entry.start.Add(milliseconds(entry.record.DelayMS)); entry.sampled && now.Before(end) {
			at = end
		}
		if !found || at.Before(wake) {
			wake, found = at, true
		}
	}

	return wake, found
}

// cssPixels formats one length in the form the engine's override relayout
// accepts. A value that rounds to zero prints as 0px rather than a negative
// zero or a long tail.
func cssPixels(value float64) string {
	if math.Abs(value) < geometryEpsilon {
		value = 0
	}

	return strconv.FormatFloat(value, 'f', -1, 64) + "px"
}
