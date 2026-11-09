package cssmotion

import (
	"slices"

	"github.com/chinmay-sawant/blinkless/layout"
)

// emptyIndex is the shared immutable index returned for nil or empty displays.
// It performs 0 allocations and its lookups are all empty.
var emptyIndex = &Index{} //nolint:gochecknoglobals // reason: shared immutable zero index returned by EmptyIndex, allocated once per process.

// EmptyIndex returns the shared immutable empty index with 0 allocations.
func EmptyIndex() *Index {
	return emptyIndex
}

// Index resolves operation ownership for one display list without scanning
// it per box. A bind builds one index for the display it resolves and shares
// it across every animated box and every transition capture, instead of
// scanning the whole operation list once per box per paint class.
//
// The index is read-only after construction. Sampling mutates operation
// paint but never adds, removes, renumbers, or reassigns operations, so one
// index serves every selection the bind makes against its display.
type Index struct {
	display *layout.Display
	// owners lists operation positions per owning box index, in operation
	// order. It is only consulted when ownersValid holds, the same condition
	// Select requires.
	owners [][]int
	// byID maps an operation id to its positions, in operation order.
	byID map[uint64][]int
	// ownersValid reports that every operation has a nearest owner, so an
	// owner-scoped selection can trust the owner map.
	ownersValid bool
}

// NewIndex builds the shared lookup for d. A nil display, or a display with no
// operations, returns a shared immutable empty index with 0 allocs.
func NewIndex(d *layout.Display) *Index {
	if d == nil || len(d.Ops) == 0 {
		return emptyIndex
	}
	x := &Index{display: d}
	x.byID = make(map[uint64][]int, len(d.Ops))
	for i := range d.Ops {
		id := d.Ops[i].ID
		x.byID[id] = append(x.byID[id], i)
	}
	if len(d.OpOwners) != len(d.Ops) {
		return x
	}
	x.ownersValid = true
	x.owners = make([][]int, len(d.Boxes))
	for i := range d.Ops {
		owner := d.OpOwners[i]
		if owner < 0 || owner >= len(d.Boxes) {
			continue
		}
		x.owners[owner] = append(x.owners[owner], i)
	}
	return x
}

// NewKeyframesIndex returns the shared index for a keyframe rule set and display.
// If len(keyframes) == 0, it returns the shared emptyIndex immediately.
func NewKeyframesIndex(keyframes []layout.KeyframesRule, d *layout.Display) *Index {
	if len(keyframes) == 0 {
		return emptyIndex
	}
	return NewIndex(d)
}

// Display returns the display this index was built for, or nil for an empty index.
func (x *Index) Display() *layout.Display {
	if x == nil {
		return nil
	}
	return x.display
}

// NumOps returns the number of operations in the indexed display.
func (x *Index) NumOps() int {
	if x == nil || x.display == nil {
		return 0
	}
	return len(x.display.Ops)
}

// Repoint updates the underlying display pointer when a new display with
// identical operation topology is committed.
func (x *Index) Repoint(d *layout.Display) {
	if x != nil && x != emptyIndex {
		x.display = d
	}
}

// Select returns the operation indices a property of one kind may write,
// exactly as Select does: owner-scoped kinds read the nearest-owner map and
// fall back to the filtered subtree without a trusted index, while a
// transform or an opacity keeps the whole subtree in operation order.
func (x *Index) Select(box layout.Box, boxIndex int, kind Kind) []int {
	if x == nil || x.display == nil {
		return nil
	}
	if kind == KindSubtree {
		return x.subtree(box)
	}
	if boxIndex >= 0 && x.ownersValid {
		var out []int
		if boxIndex < len(x.owners) {
			for _, index := range x.owners[boxIndex] {
				if matches(kind, &x.display.Ops[index]) {
					out = append(out, index)
				}
			}
		}
		return out
	}
	var out []int
	for _, index := range x.subtree(box) {
		if matches(kind, &x.display.Ops[index]) {
			out = append(out, index)
		}
	}
	return out
}

// subtree returns the operation positions the box's subtree carries, in
// operation order, matching OperationIndices including its geometric fallback
// for boxes that predate operation ids.
func (x *Index) subtree(box layout.Box) []int {
	if x == nil || x.display == nil {
		return nil
	}
	d := x.display
	if box.OperationIDs == nil {
		return geometricIndices(d, box)
	}
	var out []int
	for _, id := range box.OperationIDs {
		out = append(out, x.byID[id]...)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// ImageIndex returns the raster an inline SVG paints: the last image
// operation the box owns, with the same owner, id-set, and geometric tiers
// the SVG binder uses. An empty slice means the element painted no SVG.
func (x *Index) ImageIndex(box layout.Box, boxIndex int) []int {
	if x == nil || x.display == nil {
		return nil
	}
	d := x.display
	if boxIndex >= 0 && x.ownersValid && boxIndex < len(x.owners) {
		last := -1
		for _, index := range x.owners[boxIndex] {
			if d.Ops[index].Kind == layout.DisplayOpImage {
				last = index
			}
		}
		if last >= 0 {
			return []int{last}
		}
	}
	if len(box.OperationIDs) > 0 {
		last := -1
		for _, id := range box.OperationIDs {
			for _, index := range x.byID[id] {
				if d.Ops[index].Kind == layout.DisplayOpImage && index > last {
					last = index
				}
			}
		}
		if last >= 0 {
			return []int{last}
		}
		return nil
	}
	last := -1
	for i := range d.Ops {
		if d.Ops[i].Kind != layout.DisplayOpImage {
			continue
		}
		if !Contains(d, box, d.Ops[i]) {
			continue
		}
		last = i
	}
	if last >= 0 {
		return []int{last}
	}
	return nil
}
