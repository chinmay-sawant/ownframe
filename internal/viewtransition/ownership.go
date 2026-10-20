package viewtransition

import "github.com/chinmay-sawant/ownframe"

func captureOps(d *ownframe.Display, root int, names map[string]node) []int {
	owners := make(map[int]bool)
	for i := range d.Boxes {
		if inside(d.Boxes, i, root) && !nestedCapture(d.Boxes, i, root, names) {
			owners[i] = true
		}
	}
	indices := make(map[int]int)
	for i := range d.Ops {
		owner := ownerAt(d, i)
		if owners[owner] {
			indices[i] = i
		}
	}
	return orderedSubset(d, indices)
}

func nestedCapture(boxes []ownframe.Box, i, root int, names map[string]node) bool {
	for p, steps := i, 0; p >= 0 && p < len(boxes) && steps < len(boxes); steps++ {
		if p != root {
			for _, n := range names {
				if n.index == p {
					return true
				}
			}
		}
		if p == root {
			return false
		}
		p = boxes[p].ParentIndex
	}
	return false
}

func ownerAt(d *ownframe.Display, op int) int {
	if len(d.OpOwners) == len(d.Ops) {
		return d.OpOwners[op]
	}
	return ownerFromIDs(d, d.Ops[op].ID)
}

func ownerFromIDs(d *ownframe.Display, id uint64) int {
	best, depth := -1, -1
	for i, b := range d.Boxes {
		if !hasID(b.OperationIDs, id) {
			continue
		}
		level := boxDepth(d.Boxes, i)
		if level > depth {
			best, depth = i, level
		}
	}
	return best
}

func hasID(ids []uint64, id uint64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func boxDepth(boxes []ownframe.Box, i int) int {
	depth := 0
	for p, steps := boxes[i].ParentIndex, 0; p >= 0 && p < len(boxes) && steps < len(boxes); steps++ {
		depth++
		p = boxes[p].ParentIndex
	}
	return depth
}
