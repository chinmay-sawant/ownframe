package main

import (
	"github.com/chinmay-sawant/ownframe"
)

// patchCostOps returns the operations one element owns, in display order. It
// uses the operation ids the element box carries, the rule the page itself
// uses to find the element's own operations.
func patchCostOps(d *ownframe.Display, id string) []ownframe.DisplayOp {
	var out []ownframe.DisplayOp
	for _, b := range d.Boxes {
		if b.ID != id {
			continue
		}
		own := make(map[uint64]bool, len(b.OperationIDs))
		for _, op := range b.OperationIDs {
			own[op] = true
		}
		for _, op := range d.Ops {
			if own[op.ID] {
				out = append(out, op)
			}
		}
		return out
	}
	return out
}

// patchCostOtherOps returns every operation the element does not own, in
// display order.
func patchCostOtherOps(d *ownframe.Display, id string) []ownframe.DisplayOp {
	own := map[uint64]bool{}
	for _, b := range d.Boxes {
		if b.ID != id {
			continue
		}
		for _, op := range b.OperationIDs {
			own[op] = true
		}
	}
	var out []ownframe.DisplayOp
	for _, op := range d.Ops {
		if !own[op.ID] {
			out = append(out, op)
		}
	}
	return out
}

// patchCostBox returns the border box of one element.
func patchCostBox(d *ownframe.Display, id string) (ownframe.Box, bool) {
	for _, b := range d.Boxes {
		if b.ID == id {
			return b, true
		}
	}
	return ownframe.Box{}, false
}
