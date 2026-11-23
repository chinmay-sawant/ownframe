package replay

import "github.com/chinmay-sawant/blinkless/layout"

// CachePlan builds and caches the group plan for d without painting. The
// window uses it in tests to prove a new display generation evicts a plan
// even when the display pointer is recycled.
func (renderer *Renderer) CachePlan(d *layout.Display) {
	if renderer == nil || d == nil {
		return
	}
	renderer.planFor(d)
}

// PlanCached reports whether a group plan for d is currently cached.
func (renderer *Renderer) PlanCached(d *layout.Display) bool {
	if renderer == nil || d == nil {
		return false
	}
	renderer.plans.mu.Lock()
	defer renderer.plans.mu.Unlock()

	for i := range renderer.plans.entries {
		if renderer.plans.entries[i].d == d && renderer.plans.entries[i].root != nil {
			return true
		}
	}

	return false
}
