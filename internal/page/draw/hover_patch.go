package draw

import (
	"github.com/chinmay-sawant/blinkless/layout"
)

// Hover patch budget: the subtrees a hover change re-places must cover at
// most half the page boxes. A wider change is close enough to a full layout
// that two patches plus a full diff would cost more than laying the page out
// again, so the caller falls back to the full layout.
const hoverPatchShare = 2

// hoverPatchSides is the most roots one hover change patches: one subtree
// losing the hover and one gaining it.
const hoverPatchSides = 2

// HoverPatchRoots returns the ids whose subtrees a hover-only change must
// re-lay out: on each side of the change, the child of the lowest common
// ancestor that contains the hovered element. Patching those subtrees covers
// the elements whose :hover match changed, their descendants, and the
// ancestors between them and the common ancestor, whose own :hover rules also
// changed. Everything above the common ancestor matched :hover in both
// frames, so its paint is already correct.
//
// An empty side (entering from or leaving to no element) patches the hovered
// element itself, or its parent when the hovered element is a leaf: a leaf
// hover paints its ancestors (the row behind the cell), so the parent scope
// covers both. Ancestors above the chosen scope matched :hover in neither
// frame when leaving, and only in the new frame when entering; the patch
// leaves them stale, so a page whose sheets style those top containers with
// :hover must keep the full layout. In practice those containers are layout
// shells without :hover paint, and the budget keeps a near-page patch on the
// full path. A missing box, no common ancestor, or subtrees wider than the
// patch budget falls back to the full layout.
//
// Sibling combinators after :hover and :has() can style elements outside
// those subtrees; HoverPatchRisky guards those patterns and the caller falls
// back to the full layout when the sheets use them.
//
// ok is false when the change cannot be scoped: a missing box, no common
// ancestor, or subtrees wider than the patch budget.
func HoverPatchRoots(boxes []layout.Box, oldID, newID string) ([]string, bool) {
	if oldID == newID {
		return nil, false
	}

	if oldID == "" || newID == "" {
		return hoverEnterLeaveRoots(boxes, oldID, newID)
	}

	index := make(map[string]int, len(boxes))
	for i := range boxes {
		if boxes[i].ID != "" {
			if _, seen := index[boxes[i].ID]; !seen {
				index[boxes[i].ID] = i
			}
		}
	}

	oldIdx, ok := index[oldID]
	if !ok {
		return nil, false
	}

	newIdx, ok := index[newID]
	if !ok {
		return nil, false
	}

	oldChain := hoverAncestors(boxes, oldIdx)
	newChain := hoverAncestors(boxes, newIdx)
	if oldChain == nil || newChain == nil {
		return nil, false
	}

	lca := hoverLCA(oldChain, newChain)
	if lca < 0 {
		return nil, false
	}

	roots := make([]int, 0, hoverPatchSides)
	if root, ok := hoverSideRoot(oldChain, lca); ok {
		roots = append(roots, root)
	}

	if root, ok := hoverSideRoot(newChain, lca); ok && (len(roots) == 0 || root != roots[0]) {
		roots = append(roots, root)
	}

	if len(roots) == 0 {
		return nil, false
	}

	if !hoverPatchFits(boxes, roots) {
		return nil, false
	}

	ids := make([]string, 0, len(roots))
	for _, root := range roots {
		ids = append(ids, boxes[root].ID)
	}

	return ids, true
}

// hoverEnterLeaveRoots returns the single root an entering or leaving hover
// must re-lay out: the hovered element itself, or its parent when the
// hovered element is a leaf. A leaf hover changes its ancestors' :hover
// paint (the row behind the price), so patching the leaf alone would miss
// the row; patching the parent covers both. ok is false when the hovered id
// is missing, has no id-bearing scope, or that scope is wider than the
// patch budget.
func hoverEnterLeaveRoots(boxes []layout.Box, oldID, newID string) ([]string, bool) {
	hovered := newID
	if hovered == "" {
		hovered = oldID
	}

	if hovered == "" {
		return nil, false
	}

	index := make(map[string]int, len(boxes))
	for i := range boxes {
		if boxes[i].ID != "" {
			if _, seen := index[boxes[i].ID]; !seen {
				index[boxes[i].ID] = i
			}
		}
	}

	start, ok := index[hovered]
	if !ok {
		return nil, false
	}

	// A leaf hover paints its ancestors (the row behind the cell), so start
	// from the parent; otherwise start from the hovered element itself.
	scope := start
	if hoverIsLeaf(boxes, start) {
		scope = boxes[start].ParentIndex
	}

	// Walk up to the nearest id-bearing scope.
	for scope >= 0 && scope < len(boxes) && boxes[scope].ID == "" {
		scope = boxes[scope].ParentIndex
	}

	if scope < 0 || scope >= len(boxes) || boxes[scope].ID == "" {
		return nil, false
	}

	if !hoverPatchFits(boxes, []int{scope}) {
		return nil, false
	}

	return []string{boxes[scope].ID}, true
}

// hoverIsLeaf reports whether no box names idx as its parent.
func hoverIsLeaf(boxes []layout.Box, idx int) bool {
	for i := range boxes {
		if boxes[i].ParentIndex == idx {
			return false
		}
	}

	return true
}

// hoverAncestors returns the box chain from idx up to the outermost ancestor,
// or nil when the chain leaves the slice (a cycle or a dangling parent).
func hoverAncestors(boxes []layout.Box, idx int) []int {
	chain := []int{idx}
	seen := map[int]bool{idx: true}

	for range len(boxes) + 1 {
		parent := boxes[chain[len(chain)-1]].ParentIndex
		if parent < 0 {
			return chain
		}

		if parent >= len(boxes) || seen[parent] {
			return nil
		}

		seen[parent] = true
		chain = append(chain, parent)
	}

	return nil
}

// hoverLCA returns the deepest box both chains share, or -1. Each chain runs
// from the element up, so the first new-chain member in the old set wins.
func hoverLCA(oldChain, newChain []int) int {
	inOld := make(map[int]bool, len(oldChain))
	for _, idx := range oldChain {
		inOld[idx] = true
	}

	for _, idx := range newChain {
		if inOld[idx] {
			return idx
		}
	}

	return -1
}

// hoverSideRoot returns the chain member just below the common ancestor: the
// topmost element whose :hover match changed on that side. ok is false when
// the hovered element is the common ancestor itself, whose match did not
// change because it contains the other side's element too.
func hoverSideRoot(chain []int, lca int) (int, bool) {
	for i, idx := range chain {
		if idx == lca {
			if i == 0 {
				return 0, false
			}

			return chain[i-1], true
		}
	}

	return 0, false
}

// hoverPatchFits reports whether the subtrees under roots together cover at
// most half the page boxes.
func hoverPatchFits(boxes []layout.Box, roots []int) bool {
	if len(boxes) == 0 {
		return false
	}

	inRoot := make(map[int]bool, len(roots))
	for _, root := range roots {
		inRoot[root] = true
	}

	count := 0

	for i := range boxes {
		for idx := i; ; {
			if inRoot[idx] {
				count++

				break
			}

			parent := boxes[idx].ParentIndex
			if parent < 0 || parent >= len(boxes) {
				break
			}

			idx = parent
		}
	}

	return count > 0 && count*hoverPatchShare <= len(boxes)
}
