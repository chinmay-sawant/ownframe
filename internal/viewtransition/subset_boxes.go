package viewtransition

import "github.com/chinmay-sawant/ownframe"

func subsetBoxes(boxes []ownframe.Box, ops []ownframe.DisplayOp) []ownframe.Box {
	ids := make(map[uint64]bool, len(ops))
	for _, op := range ops {
		ids[op.ID] = true
	}
	out := make([]ownframe.Box, len(boxes))
	for i, box := range boxes {
		out[i] = box
		if box.OperationIDs == nil {
			continue
		}
		out[i].OperationIDs = nil
		for _, id := range box.OperationIDs {
			if ids[id] {
				out[i].OperationIDs = append(out[i].OperationIDs, id)
			}
		}
		if out[i].OperationIDs == nil {
			out[i].OperationIDs = []uint64{}
		}
	}
	return out
}
