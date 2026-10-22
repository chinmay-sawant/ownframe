package replay

import (
	"image"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
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
	maskBox   image.Rectangle
	maskTex   *ebiten.Image
	maskMu    sync.Mutex
	known     bool
	hasMask   bool
	unbounded bool
}

// planEntry keeps one display's plan. A frame that repaints several rects
// builds the group tree once. plan is the chunk-aware paint order built on
// demand from the same tree.
type planEntry struct {
	d    *layout.Display
	root *groupNode
	plan *chunkPlan
}

var plans struct {
	mu      sync.Mutex
	entries [2]planEntry
	next    int
}

// planFor returns the display's group tree, from a small cache keyed on the
// display pointer. A display is replaced wholesale on redraw, so a pointer
// match is a plan match. Call on the draw path only; the display must not be
// mutated concurrently. Tick patches ops on the same loop, so never share a
// Page across goroutines. The lock stays held across a miss build, so two
// concurrent callers serialize instead of building twice and evicting.
func planFor(d *layout.Display) *groupNode {
	plans.mu.Lock()
	defer plans.mu.Unlock()

	for i := range plans.entries {
		if plans.entries[i].d == d {
			return plans.entries[i].root
		}
	}

	root := buildPlan(d)

	disposePlan(plans.entries[plans.next].root)
	plans.entries[plans.next] = planEntry{d: d, root: root}
	plans.next = (plans.next + 1) % len(plans.entries)

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
