package element

// UndoRef holds a stable undo/redo reference attached to an element.
type UndoRef struct {
	Identity string
	Epoch    int
}

// AttachUndo stores an undo reference on this node.
func (n *Node) AttachUndo(ref UndoRef) {
	if n == nil {
		return
	}
	n.Metadata["undo"] = ref
	n.UndoRef = ref
}

// UndoReference returns the attached undo reference.
func (n *Node) UndoReference() (UndoRef, bool) {
	if n == nil {
		return UndoRef{}, false
	}
	v, ok := n.Metadata["undo"]
	if !ok {
		if n.UndoRef.Identity != "" {
			return UndoRef{Identity: n.UndoRef.Identity}, true
		}
		return UndoRef{}, false
	}
	r, ok := v.(UndoRef)
	if !ok {
		return UndoRef{}, false
	}
	return r, true
}
