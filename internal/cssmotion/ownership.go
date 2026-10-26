package cssmotion

import (
	"slices"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Kind names the operations one paint-property family writes. A colour belongs
// to the element that paints it, so selection is by nearest owner; a transform
// or an opacity moves or fades the element's whole subtree.
type Kind uint8

// The operation sets a property family may write.
const (
	// KindFill is a background, accent, or rule fill: the owner's FillRect
	// operations.
	KindFill Kind = iota + 1
	// KindInk is text ink: the owner's Text operations.
	KindInk
	// KindStroke is a border or an outline: the owner's StrokeRect and Line
	// operations.
	KindStroke
	// KindSubtree is a transform or an opacity: every operation the box's
	// subtree carries, because a moved parent moves its children.
	KindSubtree
)

// Select returns the operation indices a property of one kind may write.
// boxIndex is the box's position in d.Boxes, or -1 when the caller cannot
// trust it; an owner-scoped kind then falls back to the box's subtree.
//
// A colour only writes the operations whose nearest owner is this box
// (d.OpOwners), so a parent's background colour cannot recolour a child's card
// and cannot touch its own label text. Transform and opacity keep the subtree
// the box's OperationIDs describe.
func Select(d *layout.Display, box layout.Box, boxIndex int, kind Kind) []int {
	if d == nil {
		return nil
	}
	if kind == KindSubtree {
		return OperationIndices(d, box)
	}
	if boxIndex >= 0 && len(d.OpOwners) == len(d.Ops) {
		var out []int
		for i := range d.Ops {
			if d.OpOwners[i] == boxIndex && matches(kind, &d.Ops[i]) {
				out = append(out, i)
			}
		}
		return out
	}
	// No trusted owner index: keep the property on the box's subtree, filtered
	// to the kinds it paints. The caller counts the fallback.
	var out []int
	for _, index := range OperationIndices(d, box) {
		if matches(kind, &d.Ops[index]) {
			out = append(out, index)
		}
	}
	return out
}

// matches reports whether one operation belongs to a property kind. A kind
// outside these sets is not a paint field a keyframe can write: an image and a
// grid run carry no colour, and a background never paints text.
//
// A corner-shape path carries both a fill and a border under one op kind, so
// the path's Stroke flag decides: a background recolours the fill path and a
// border or outline recolours the stroke path. Without this, a corner-shaped
// element would keep its declared colours while its rectangular neighbours
// moved.
func matches(kind Kind, op *layout.DisplayOp) bool {
	if op.Kind == layout.DisplayOpPath {
		corner, ok := op.CornerPathValue()
		if !ok || corner == nil {
			return false
		}

		switch kind {
		case KindFill:
			return !corner.Stroke
		case KindStroke:
			return corner.Stroke
		}

		return false
	}

	switch kind {
	case KindFill:
		return op.Kind == layout.DisplayOpFillRect
	case KindInk:
		return op.Kind == layout.DisplayOpText
	case KindStroke:
		return op.Kind == layout.DisplayOpStrokeRect || op.Kind == layout.DisplayOpLine
	}
	return false
}

// OperationIndices captures operations owned by a box. Older Blinkless boxes
// without ownership IDs use the geometric fallback in Contains.
func OperationIndices(d *layout.Display, box layout.Box) []int {
	if d == nil {
		return nil
	}
	if box.OperationIDs == nil {
		return geometricIndices(d, box)
	}
	ids := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		ids[id] = struct{}{}
	}
	var out []int
	for i := range d.Ops {
		if _, ok := ids[d.Ops[i].ID]; ok {
			out = append(out, i)
		}
	}
	return out
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

// NewIndex builds the shared lookup for d. A nil display builds an empty
// index whose selections are all empty, matching Select on nil.
func NewIndex(d *layout.Display) *Index {
	x := &Index{display: d}
	if d == nil {
		return x
	}
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
	if x == nil {
		return nil
	}
	d := x.display
	if d == nil {
		return nil
	}
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

func geometricIndices(d *layout.Display, box layout.Box) []int {
	var out []int
	for i := range d.Ops {
		if Contains(d, box, d.Ops[i]) {
			out = append(out, i)
		}
	}
	return out
}
