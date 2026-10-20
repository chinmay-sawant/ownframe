package viewtransition

import "github.com/chinmay-sawant/ownframe"

// Display combines the base scene with the sampled transition paint layers.
func (f Frame) Display() *ownframe.Display {
	if f.Base == nil {
		return nil
	}
	out := *f.Base
	out.Ops = make([]ownframe.DisplayOp, 0, len(f.Base.Ops))
	for _, op := range f.Base.Ops {
		out.Ops = append(out.Ops, op.Clone())
	}
	out.Order = append([]int(nil), f.Base.Order...)
	out.OpOwners = append([]int(nil), f.Base.OpOwners...)
	out.Boxes = append([]ownframe.Box(nil), f.Base.Boxes...)
	for i := range out.Boxes {
		out.Boxes[i].OperationIDs = append([]uint64(nil), f.Base.Boxes[i].OperationIDs...)
	}
	var id uint64
	for _, op := range out.Ops {
		if op.ID > id {
			id = op.ID
		}
	}
	for _, group := range f.Groups {
		appendOps(&out, &id, group.Old, nil)
		appendOps(&out, &id, group.New, group.newOwners)
	}
	return &out
}

func appendOps(d *ownframe.Display, id *uint64, ops []ownframe.DisplayOp, owners []int) {
	for i, op := range ops {
		*id = *id + 1
		op.ID = *id
		index := len(d.Ops)
		d.Ops = append(d.Ops, op.Clone())
		d.Order = append(d.Order, index)
		owner := -1
		if i < len(owners) {
			owner = owners[i]
		}
		d.OpOwners = append(d.OpOwners, owner)
		if owner >= 0 && owner < len(d.Boxes) {
			d.Boxes[owner].OperationIDs = append(d.Boxes[owner].OperationIDs, *id)
		}
	}
}
