package element

// ReconcileResult carries a reused root and its patches.
type ReconcileResult struct {
	Root     *Node
	Patches  PatchList
	Retained []*Node
	Disposed []*Node
}

// Reconcile reuses matching instances across two trees.
func Reconcile(oldRoot, newRoot *Node) ReconcileResult {
	res := ReconcileResult{}
	if oldRoot == nil {
		res.Root = newRoot
		res.Patches = createPatches(newRoot)
		collectRetained(newRoot, &res.Retained)
		return res
	}
	if newRoot == nil {
		collectDisposed(oldRoot, &res.Disposed)
		res.Root = nil
		res.Patches = deletePatches(oldRoot)
		return res
	}
	if !keysMatch(oldRoot, newRoot) {
		collectDisposed(oldRoot, &res.Disposed)
		res.Root = newRoot
		res.Patches = append(deletePatches(oldRoot), createPatches(newRoot)...)
		collectRetained(newRoot, &res.Retained)
		return res
	}
	props := changedProps(oldRoot, newRoot)
	copyMerge(oldRoot, newRoot)
	var childPatches PatchList
	var disposed []*Node
	oldRoot.Children, childPatches, disposed = reconcileChildren(oldRoot, oldRoot.Children, newRoot.Children)
	res.Root = oldRoot
	if len(props) > 0 {
		res.Patches = append(PatchList{{Op: OpUpdate, Node: oldRoot, Props: props}}, childPatches...)
	} else {
		res.Patches = childPatches
	}
	res.Retained = []*Node{}
	collectRetained(oldRoot, &res.Retained)
	res.Disposed = append(res.Disposed, disposed...)
	return res
}

// Diff returns patches without mutating ownership expectations.
func Diff(oldRoot, newRoot *Node) PatchList {
	return Reconcile(oldRoot, newRoot).Patches
}

func keysMatch(a, b *Node) bool {
	if a == nil || b == nil {
		return false
	}
	ak, bk := KeyOf(a), KeyOf(b)
	if ak != "" && ak == bk {
		return true
	}
	return a.ID != "" && a.ID == b.ID
}

func copyMerge(dst, src *Node) {
	if dst == nil || src == nil {
		return
	}
	dst.Key = src.Key
	dst.Kind = src.Kind
	dst.Attrs = cloneAttrs(src.Attrs)
	dst.Metadata = cloneMeta(src.Metadata)
}
