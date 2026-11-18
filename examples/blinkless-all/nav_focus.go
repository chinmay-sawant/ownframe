package main

import (
	"math"
	"strings"
)

// navDir is a direction key.
type navDir int

const (
	navUp navDir = iota
	navDown
	navLeft
	navRight
)

func (d navDir) prop() string {
	return [...]string{"nav-up", "nav-down", "nav-left", "nav-right"}[d]
}

// scrollStep is how far a scroll action moves per key press, in CSS pixels.
const scrollStep = 40

// navResult is what one arrow key press does.
type navResult struct {
	// Target is the item that gets focus, or -1.
	Target int
	// Action is "focus" or "scroll".
	Action string
	// Via says how the target was found: "nav-" + dir, "spatial" or "none".
	Via string
	// ScrollDX and ScrollDY are set when Action is "scroll".
	ScrollDX, ScrollDY float64
}

// Move resolves one arrow key press from item from. focusable says which
// items can hold focus.
//
// The order is: spatial-navigation-action on the nearest ancestor-or-self
// (scroll turns the key into a scroll step); then an explicit nav-<dir> on the
// item (#id jumps to that id when it exists); then spatial search, limited to
// the nearest spatial-navigation-contain: contain container, using the
// nearest spatial-navigation-function (normal or grid).
func (m navModel) Move(from int, dir navDir, focusable func(navItem) bool) navResult {
	none := navResult{Target: -1, Action: "focus", Via: "none"}
	if from < 0 || from >= len(m.Items) {
		return none
	}

	if act, _ := m.declaredUp(from, "spatial-navigation-action"); act == "scroll" {
		r := navResult{Target: -1, Action: "scroll", Via: "spatial-navigation-action"}

		switch dir {
		case navUp:
			r.ScrollDY = -scrollStep
		case navDown:
			r.ScrollDY = scrollStep
		case navLeft:
			r.ScrollDX = -scrollStep
		case navRight:
			r.ScrollDX = scrollStep
		}

		return r
	}

	if t := m.explicitTarget(from, dir); t >= 0 {
		return navResult{Target: t, Action: "focus", Via: dir.prop()}
	}

	container := -1
	if v, at := m.declaredUp(from, "spatial-navigation-contain"); v == "contain" {
		container = at
	}

	fn, _ := m.declaredUp(from, "spatial-navigation-function")
	best, bestScore := -1, math.Inf(1)

	for i, c := range m.Items {
		if i == from || !focusable(c) {
			continue
		}

		if container >= 0 && !m.isUnder(i, container) {
			continue
		}

		if score, ok := navScore(m.Items[from], c, dir, fn == "grid"); ok && score < bestScore {
			best, bestScore = i, score
		}
	}

	if best < 0 {
		return none
	}

	return navResult{Target: best, Action: "focus", Via: "spatial"}
}

// explicitTarget reads nav-<dir> on item from. It returns -1 for auto, for an
// unknown id, and for a value that names a frame the model does not have.
func (m navModel) explicitTarget(from int, dir navDir) int {
	v := m.Items[from].Declared[dir.prop()]
	if v == "" || v == "auto" {
		return -1
	}

	fields := strings.Fields(v)
	if len(fields) == 0 || !strings.HasPrefix(fields[0], "#") {
		return -1
	}

	// A frame name other than current or root points at a different
	// browsing context, which this host does not have.
	if len(fields) > 1 && fields[1] != "current" && fields[1] != "root" {
		return -1
	}

	return m.ByID(fields[0][1:])
}

// navScore rates candidate c as a target of a key press from f. Lower is
// better. It returns false when c is not in that direction. The score is the
// gap along the key axis plus twice the sideways offset of the centres. In
// grid mode c must also overlap f on the other axis, so only the same row or
// column is reachable.
func navScore(f, c navItem, dir navDir, grid bool) (float64, bool) {
	const eps = 0.5

	var gap, side float64

	var overlap bool

	switch dir {
	case navRight:
		gap = c.X - (f.X + f.W)
		side = math.Abs((c.Y + c.H/2) - (f.Y + f.H/2))
		overlap = c.Y < f.Y+f.H && f.Y < c.Y+c.H
	case navLeft:
		gap = f.X - (c.X + c.W)
		side = math.Abs((c.Y + c.H/2) - (f.Y + f.H/2))
		overlap = c.Y < f.Y+f.H && f.Y < c.Y+c.H
	case navDown:
		gap = c.Y - (f.Y + f.H)
		side = math.Abs((c.X + c.W/2) - (f.X + f.W/2))
		overlap = c.X < f.X+f.W && f.X < c.X+c.W
	case navUp:
		gap = f.Y - (c.Y + c.H)
		side = math.Abs((c.X + c.W/2) - (f.X + f.W/2))
		overlap = c.X < f.X+f.W && f.X < c.X+c.W
	}

	if gap < -eps || (grid && !overlap) {
		return 0, false
	}

	return math.Max(gap, 0) + 2*side, true
}
