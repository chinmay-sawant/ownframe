package replay

import (
	"image"
	"sync"
	"sync/atomic"

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
	group          *layout.DisplayGroup
	items          []paintItem
	box            image.Rectangle
	maskBox        image.Rectangle
	backdropBox    image.Rectangle
	backdropMark   *layout.DisplayOp
	backdrop       backdropFilter
	backdropRadius float64
	maskTex        *ebiten.Image
	maskMu         sync.Mutex
	known          bool
	hasMask        bool
	hasBackdropBox bool
	hasBackdrop    bool
	unbounded      bool
}

// planEntry keeps one display's plan. A frame that repaints several rects
// builds the group tree once. plan is the chunk-aware paint order built on
// demand from the same tree: the first tile paint lazily folds it, so a frame
// that never touches chunks pays no fold cost. gen is the display shape the
// plans were built from (see displayFingerprint); an element splice changes
// the shape and rebuilds.
type planEntry struct {
	d    *layout.Display
	gen  uint64
	root *groupNode
	plan *chunkPlan
}

// groupPlanBuilds counts group-tree builds. Tests prove plan reuse by
// observing no build across identical repaints, not by timing.
var groupPlanBuilds atomic.Int64

// groupPlanBuildCount reports group-tree builds.
func groupPlanBuildCount() int64 {
	return groupPlanBuilds.Load()
}

type planCache struct {
	mu      sync.Mutex
	entries [2]planEntry
	next    int
}

// planFor returns the display's group tree, from a small cache keyed on the
// display pointer and its generation fingerprint. A display is replaced
// wholesale on redraw, so a pointer match with the same shape is a plan
// match; an element splice changes lengths or canvas size, changes the
// fingerprint, and rebuilds. The chunk plan stays nil until the first tile
// paint folds it lazily from the same tree. Call on the draw path only; the
// display must not be mutated concurrently. Tick patches ops on the same
// loop, so never share a Page across goroutines. The lock stays held across
// a miss build, so two concurrent callers serialize instead of building
// twice and evicting.
func (renderer *Renderer) planFor(d *layout.Display) *groupNode {
	gen := displayFingerprint(d)
	renderer.plans.mu.Lock()
	defer renderer.plans.mu.Unlock()

	for i := range renderer.plans.entries {
		if renderer.plans.entries[i].d == d && renderer.plans.entries[i].gen == gen {
			return renderer.plans.entries[i].root
		}
	}

	root := buildPlan(d)
	groupPlanBuilds.Add(1)

	disposePlan(renderer.plans.entries[renderer.plans.next].root)
	renderer.plans.entries[renderer.plans.next] = planEntry{d: d, gen: gen, root: root}
	renderer.plans.next = (renderer.plans.next + 1) % len(renderer.plans.entries)

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
