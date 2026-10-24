package cssmotion

import "github.com/chinmay-sawant/blinkless/layout"

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
			if d.OpOwners[i] == boxIndex && matches(kind, d.Ops[i].Kind) {
				out = append(out, i)
			}
		}
		return out
	}
	// No trusted owner index: keep the property on the box's subtree, filtered
	// to the kinds it paints. The caller counts the fallback.
	var out []int
	for _, index := range OperationIndices(d, box) {
		if matches(kind, d.Ops[index].Kind) {
			out = append(out, index)
		}
	}
	return out
}

// matches reports whether one operation kind belongs to a property kind. A
// kind outside these sets is not a paint field a keyframe can write: an image
// and a grid run carry no colour, and a background never paints text.
func matches(kind Kind, op layout.DisplayKind) bool {
	switch kind {
	case KindFill:
		return op == layout.DisplayOpFillRect
	case KindInk:
		return op == layout.DisplayOpText
	case KindStroke:
		return op == layout.DisplayOpStrokeRect || op == layout.DisplayOpLine
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

func geometricIndices(d *layout.Display, box layout.Box) []int {
	var out []int
	for i := range d.Ops {
		if Contains(d, box, d.Ops[i]) {
			out = append(out, i)
		}
	}
	return out
}
