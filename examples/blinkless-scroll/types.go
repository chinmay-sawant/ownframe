package main

// Plain data the host engine works on. Nothing here imports ownframe or a
// window, so every rule in this program runs in a unit test with numbers only.
// adapter.go is the single place that turns blinkless layout boxes into these
// types.

// Rect is a border box in CSS pixels.
type Rect struct{ X, Y, W, H float64 }

// Axis is "x" or "y".
type Axis string

const (
	AxisX Axis = "x"
	AxisY Axis = "y"
)

// start returns the box edge on the axis, size the box length on it.
func (r Rect) start(a Axis) float64 {
	if a == AxisX {
		return r.X
	}

	return r.Y
}

func (r Rect) size(a Axis) float64 {
	if a == AxisX {
		return r.W
	}

	return r.H
}

// Viewport is the page scroller: how far it is scrolled, how big the window
// is, and how big the content is.
type Viewport struct {
	ScrollX, ScrollY float64
	W, H             float64
	ContentW         float64
	ContentH         float64
}

func (v Viewport) offset(a Axis) float64 {
	if a == AxisX {
		return v.ScrollX
	}

	return v.ScrollY
}

func (v Viewport) size(a Axis) float64 {
	if a == AxisX {
		return v.W
	}

	return v.H
}

func (v Viewport) content(a Axis) float64 {
	if a == AxisX {
		return v.ContentW
	}

	return v.ContentH
}

// Timeline kinds.
const (
	KindScroll  = "scroll"
	KindView    = "view"
	KindPointer = "pointer"
)

// Timeline is one named timeline declared by a node.
type Timeline struct {
	Name       string
	Kind       string
	Axis       Axis
	SubjectID  string
	InsetStart float64
	InsetEnd   float64
}

// ScopeKind says which names a node hoists with timeline-scope.
type ScopeKind int

const (
	ScopeNone ScopeKind = iota
	ScopeAll
	ScopeNames
)

// Scope is the timeline-scope of one node.
type Scope struct {
	Kind  ScopeKind
	Names []string
}

func (s Scope) covers(name string) bool {
	switch s.Kind {
	case ScopeAll:
		return true
	case ScopeNames:
		for _, n := range s.Names {
			if n == name {
				return true
			}
		}
	}

	return false
}

// RangeBound is one end of a trigger range. It mirrors the engine's bound:
// an optional range name, then an optional percent or length offset.
type RangeBound struct {
	Normal     bool
	Auto       bool
	Name       string
	Percent    float64
	Length     float64
	HasPercent bool
	HasLength  bool
}

// Node is one element box plus the facts the host needs about it.
type Node struct {
	ID        string
	Parent    int // index of the parent node, -1 for a top-level node
	Rect      Rect
	Timelines []Timeline
	Scope     Scope
	// Inert is the computed interactivity value of this node. It is inherited
	// by the engine already; hit-testing still walks ancestors so a hand-built
	// tree behaves the same.
	Inert     bool
	Focusable bool
}
