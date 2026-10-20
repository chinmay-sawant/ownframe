package cssmotion

import "github.com/chinmay-sawant/ownframe"

// OperationIndices captures operations owned by a box. Older Blinkless boxes
// without ownership IDs use the geometric fallback in Contains.
func OperationIndices(d *ownframe.Display, box ownframe.Box) []int {
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

func geometricIndices(d *ownframe.Display, box ownframe.Box) []int {
	var out []int
	for i := range d.Ops {
		if Contains(d, box, d.Ops[i]) {
			out = append(out, i)
		}
	}
	return out
}
