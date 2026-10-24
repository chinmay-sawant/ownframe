package csstransition

import (
	"errors"
	"fmt"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
