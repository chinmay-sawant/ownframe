package uia

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

var (
	// ErrNotAvailable means the element is gone or its document was replaced.
	ErrNotAvailable = errors.New("uia: element not available")
	// ErrNotEnabled means the element is disabled.
	ErrNotEnabled = errors.New("uia: element not enabled")
	// ErrNotSupported means the element does not offer the operation.
	ErrNotSupported = errors.New("uia: operation not supported")
	// ErrReadOnly means the value cannot be set.
	ErrReadOnly = errors.New("uia: value is read-only")
)

// ActionTimeout bounds how long a client call waits for the page loop.
const ActionTimeout = 3 * time.Second

// Performer runs one request on the page-owning loop and waits for the result.
type Performer func(context.Context, host.AccessibilityRequest) error

// Direction is a UIA NavigateDirection.
type Direction int32

const (
	Parent Direction = iota
	NextSibling
	PreviousSibling
	FirstChild
	LastChild
)

// Geometry places page content on the screen in physical pixels.
type Geometry struct {
	OriginX, OriginY float64
	// Scale converts page CSS pixels to physical pixels.
	Scale         float64
	Width, Height float64
}

// Rect is a screen rectangle in physical pixels.
type Rect struct{ X, Y, W, H float64 }

// Property names the property-change events the provider reports.
type Property uint8

const (
	PropName Property = iota + 1
	PropValue
	PropToggle
	PropEnabled
	PropSelected
)

// PropertyChange records one changed property on a surviving node.
type PropertyChange struct {
	ID       host.AccessibilityNodeID
	Property Property
}

// Change describes what an Update altered, for event raising.
type Change struct {
	// First is true for the first snapshot of a provider.
	First bool
	// Document is true when the document token changed. Every node is new.
	Document bool
	// Structure is true when nodes were added, removed, or reordered.
	Structure bool
	// Props lists property changes on nodes present before and after.
	Props []PropertyChange
	// Focus is the newly focused node, or empty when focus did not move.
	Focus host.AccessibilityNodeID
}

// Provider holds the current UIA view and runs client operations. All
// methods are safe for concurrent use: UIA clients call from arbitrary
// threads while the render loop publishes.
type Provider struct {
	perform Performer

	mu      sync.RWMutex
	snap    *Snapshot
	geom    Geometry
	ids     map[host.AccessibilityNodeID]int32
	nextID  int32
	stopped bool
}

// NewProvider returns a provider that dispatches actions through perform.
func NewProvider(perform Performer) *Provider {
	return &Provider{perform: perform, ids: map[host.AccessibilityNodeID]int32{}, geom: Geometry{Scale: 1}}
}

// SetGeometry updates the content placement. Call it whenever the window
// moves, resizes, or changes display scale.
func (p *Provider) SetGeometry(g Geometry) {
	if g.Scale <= 0 {
		g.Scale = 1
	}
	p.mu.Lock()
	p.geom = g
	p.mu.Unlock()
}

// Stop makes every later operation fail with ErrNotAvailable.
func (p *Provider) Stop() {
	p.mu.Lock()
	p.stopped = true
	p.mu.Unlock()
}

// Update replaces the snapshot with tree and reports the difference.
func (p *Provider) Update(tree host.AccessibilityTree) Change {
	next := BuildSnapshot(tree)
	p.mu.Lock()
	defer p.mu.Unlock()
	prev := p.snap
	ch := Change{First: prev == nil}
	if prev == nil || prev.Document != next.Document {
		ch.Document = true
		ch.Structure = true
		p.ids = make(map[host.AccessibilityNodeID]int32, len(next.order))
		p.nextID = 0
	}
	for _, id := range next.order {
		rid, ok := p.ids[id]
		if !ok {
			p.nextID++
			rid = p.nextID
			p.ids[id] = rid
		}
		next.byID[id].RuntimeID = rid
	}
	if !ch.Document {
		ch.Structure, ch.Props = diff(prev, next)
	}
	if next.Focused != "" && (prev == nil || prev.Focused != next.Focused) {
		ch.Focus = next.Focused
	}
	p.snap = next
	return ch
}

func diff(prev, next *Snapshot) (bool, []PropertyChange) {
	structure := len(prev.order) != len(next.order) || prev.Root != next.Root
	var props []PropertyChange
	for i, id := range next.order {
		a, b := prev.byID[id], next.byID[id]
		if a == nil {
			structure = true
			continue
		}
		if !structure && prev.order[i] != id {
			structure = true
		}
		if a.Parent != b.Parent || !sameIDs(a.Children, b.Children) {
			structure = true
		}
		if a.Name != b.Name {
			props = append(props, PropertyChange{id, PropName})
		}
		if a.Value != b.Value {
			props = append(props, PropertyChange{id, PropValue})
		}
		if a.Toggled != b.Toggled {
			props = append(props, PropertyChange{id, PropToggle})
		}
		if a.Enabled != b.Enabled {
			props = append(props, PropertyChange{id, PropEnabled})
		}
		if a.Selected != b.Selected {
			props = append(props, PropertyChange{id, PropSelected})
		}
	}
	return structure, props
}

func sameIDs(a, b []host.AccessibilityNodeID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Document returns the current document token, or zero before the first Update.
func (p *Provider) Document() uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.snap == nil {
		return 0
	}
	return p.snap.Document
}

// Root returns the root node ID, or empty when there is no tree.
func (p *Provider) Root() host.AccessibilityNodeID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.snap == nil || p.stopped {
		return ""
	}
	return p.snap.Root
}

// Focused returns the focused node ID, or empty.
func (p *Provider) Focused() host.AccessibilityNodeID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.snap == nil || p.stopped {
		return ""
	}
	return p.snap.Focused
}

// Element returns a copy of the element, or ErrNotAvailable.
func (p *Provider) Element(id host.AccessibilityNodeID) (Element, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, _, err := p.lookup(id)
	if err != nil {
		return Element{}, err
	}
	out := *e
	out.Children = append([]host.AccessibilityNodeID(nil), e.Children...)
	return out, nil
}

func (p *Provider) lookup(id host.AccessibilityNodeID) (*Element, uint64, error) {
	if p.snap == nil || p.stopped {
		return nil, 0, ErrNotAvailable
	}
	e, ok := p.snap.byID[id]
	if !ok {
		return nil, 0, ErrNotAvailable
	}
	return e, p.snap.Document, nil
}

// Navigate returns the neighbor of id in direction. It returns false when
// there is none. The root has no parent: the window provider owns it.
func (p *Provider) Navigate(id host.AccessibilityNodeID, dir Direction) (host.AccessibilityNodeID, bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, _, err := p.lookup(id)
	if err != nil {
		return "", false, err
	}
	switch dir {
	case Parent:
		if id == p.snap.Root || e.Parent == "" {
			return "", false, nil
		}
		return e.Parent, true, nil
	case FirstChild:
		if len(e.Children) == 0 {
			return "", false, nil
		}
		return e.Children[0], true, nil
	case LastChild:
		if len(e.Children) == 0 {
			return "", false, nil
		}
		return e.Children[len(e.Children)-1], true, nil
	case NextSibling, PreviousSibling:
		if id == p.snap.Root || e.Parent == "" {
			return "", false, nil
		}
		parent := p.snap.byID[e.Parent]
		for i, c := range parent.Children {
			if c != id {
				continue
			}
			j := i + 1
			if dir == PreviousSibling {
				j = i - 1
			}
			if j < 0 || j >= len(parent.Children) {
				return "", false, nil
			}
			return parent.Children[j], true, nil
		}
	}
	return "", false, nil
}

// Bounds returns the element rectangle in screen pixels. The second result
// is false when the element has no geometry. The root falls back to the whole
// content area.
func (p *Provider) Bounds(id host.AccessibilityNodeID) (Rect, bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, _, err := p.lookup(id)
	if err != nil {
		return Rect{}, false, err
	}
	g := p.geom
	if !e.HasBounds {
		if id == p.snap.Root && g.Width > 0 && g.Height > 0 {
			return Rect{g.OriginX, g.OriginY, g.Width, g.Height}, true, nil
		}
		return Rect{}, false, nil
	}
	return Rect{
		X: g.OriginX + e.Bounds.X*g.Scale,
		Y: g.OriginY + e.Bounds.Y*g.Scale,
		W: math.Max(0, e.Bounds.W*g.Scale),
		H: math.Max(0, e.Bounds.H*g.Scale),
	}, true, nil
}

// ElementAt returns the element under the screen point (x, y), in the pixels
// Bounds uses. It returns the deepest element whose rectangle contains the
// point, and the later one in document order when two at the same depth
// overlap, as a later sibling paints over an earlier one. The root is never
// returned: UI Automation then falls back to the window itself.
func (p *Provider) ElementAt(x, y float64) (host.AccessibilityNodeID, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.snap == nil || p.stopped {
		return "", false
	}

	g := p.geom
	best, bestDepth := host.AccessibilityNodeID(""), -1

	for _, id := range p.snap.order {
		e := p.snap.byID[id]
		if id == p.snap.Root || !e.HasBounds {
			continue
		}

		left := g.OriginX + e.Bounds.X*g.Scale
		top := g.OriginY + e.Bounds.Y*g.Scale
		right := left + math.Max(0, e.Bounds.W*g.Scale)
		bottom := top + math.Max(0, e.Bounds.H*g.Scale)

		if x < left || x >= right || y < top || y >= bottom {
			continue
		}

		if depth := p.depthLocked(id); depth >= bestDepth {
			best, bestDepth = id, depth
		}
	}

	return best, best != ""
}

// depthLocked counts the ancestors of id. The caller holds p.mu.
func (p *Provider) depthLocked(id host.AccessibilityNodeID) int {
	depth := 0

	for e := p.snap.byID[id]; e != nil && e.Parent != ""; e = p.snap.byID[e.Parent] {
		depth++
		if depth > len(p.snap.order) { // a cycle in a malformed tree
			break
		}
	}

	return depth
}

// SetFocus moves page focus to the element.
func (p *Provider) SetFocus(id host.AccessibilityNodeID) error {
	return p.run(id, func(e *Element) error {
		if !e.canFocus {
			return ErrNotSupported
		}
		return nil
	}, host.AccessibilityFocus, "")
}

// Invoke activates a button or link.
func (p *Provider) Invoke(id host.AccessibilityNodeID) error {
	return p.run(id, func(e *Element) error {
		if !e.Invoke {
			return ErrNotSupported
		}
		return enabled(e)
	}, host.AccessibilityActivate, "")
}

// Toggle flips a checkbox.
func (p *Provider) Toggle(id host.AccessibilityNodeID) error {
	return p.run(id, func(e *Element) error {
		if !e.Toggle {
			return ErrNotSupported
		}
		return enabled(e)
	}, host.AccessibilityActivate, "")
}

// Select selects a radio button. Selecting a selected one is a no-op.
func (p *Provider) Select(id host.AccessibilityNodeID) error {
	err := p.run(id, func(e *Element) error {
		if !e.SelectionItem || !e.canActivate {
			return ErrNotSupported
		}
		if err := enabled(e); err != nil {
			return err
		}
		if e.Selected {
			return errDone
		}
		return nil
	}, host.AccessibilityActivate, "")
	if errors.Is(err, errDone) {
		return nil
	}
	return err
}

// errDone short-circuits run when the operation has nothing left to do.
var errDone = errors.New("uia: nothing to do")

// SetValue replaces the value of an editable, non-password control.
func (p *Provider) SetValue(id host.AccessibilityNodeID, value string) error {
	return p.run(id, func(e *Element) error {
		if !e.ValuePattern {
			return ErrNotSupported
		}
		if !e.canSet {
			return ErrReadOnly
		}
		return enabled(e)
	}, host.AccessibilitySetValue, value)
}

func enabled(e *Element) error {
	if !e.Enabled {
		return ErrNotEnabled
	}
	return nil
}

// run checks the element under the read lock, then dispatches without it so
// a slow page loop never blocks navigation or property reads.
func (p *Provider) run(id host.AccessibilityNodeID, check func(*Element) error, kind host.AccessibilityActionKind, value string) error {
	p.mu.RLock()
	e, doc, err := p.lookup(id)
	if err == nil {
		err = check(e)
	}
	p.mu.RUnlock()
	if err != nil {
		return err
	}
	if p.perform == nil {
		return host.ErrAccessibilityUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), ActionTimeout)
	defer cancel()
	return p.perform(ctx, host.AccessibilityRequest{Document: doc, Node: id, Kind: kind, Value: value})
}
