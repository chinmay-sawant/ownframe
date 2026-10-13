package replay

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
)

// paintItem is one step of a display's paint plan: one op, or a nested group.
type paintItem struct {
	op    *layout.DisplayOp
	child *groupNode
}

// groupNode buffers one blend or isolation group. The root has a nil group.
type groupNode struct {
	group     *layout.DisplayGroup
	items     []paintItem
	box       image.Rectangle
	known     bool
	unbounded bool
}

// planEntry keeps one display's plan. A frame that repaints several rects
// builds the group tree once.
type planEntry struct {
	d    *layout.Display
	root *groupNode
}

var plans struct {
	mu      sync.Mutex
	entries [2]planEntry
	next    int
}

// planFor returns the display's group tree, from a small cache keyed on the
// display pointer. A display is replaced wholesale on redraw, so a pointer
// match is a plan match.
func planFor(d *layout.Display) *groupNode {
	plans.mu.Lock()

	for i := range plans.entries {
		if plans.entries[i].d == d {
			root := plans.entries[i].root
			plans.mu.Unlock()

			return root
		}
	}

	plans.mu.Unlock()

	root := buildPlan(d)
	plans.mu.Lock()

	plans.entries[plans.next] = planEntry{d: d, root: root}
	plans.next = (plans.next + 1) % len(plans.entries)
	plans.mu.Unlock()

	return root
}

// hasGroups reports a display with ops inside a blend or isolation group.
func hasGroups(d *layout.Display) bool {
	for i := range d.Ops {
		if d.Ops[i].Group() != nil {
			return true
		}
	}

	return false
}
