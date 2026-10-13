package element

// reconcileChildren matches old and new children.
func reconcileChildren(parent *Node, oldKids, newKids []*Node) ([]*Node, PatchList, []*Node) {
	retained := make([]*Node, len(newKids))
	consumed := make([]bool, len(oldKids))
	var patches PatchList
	var disposed []*Node
	for i, nk := range newKids {
		idx := findMatch(oldKids, consumed, nk)
		if idx >= 0 {
			consumed[idx] = true
			oldNode := oldKids[idx]
			props := changedProps(oldNode, nk)
			copyMerge(oldNode, nk)
			childChildren, childPatches, childDisposed := reconcileChildren(oldNode, oldNode.Children, nk.Children)
			oldNode.Children = childChildren
			retained[i] = oldNode
			oldNode.Parent = parent
			if idx != i {
				patches = append(patches, Patch{Op: OpMove, Node: oldNode, Props: []string{"index"}})
			}
			if len(props) > 0 {
				patches = append(patches, Patch{Op: OpUpdate, Node: oldNode, Props: props})
			}
			patches = append(patches, childPatches...)
			disposed = append(disposed, childDisposed...)
		} else {
			retained[i] = nk
			nk.Parent = parent
			patches = append(patches, createPatches(nk)...)
		}
	}
	for j, ok := range oldKids {
		if consumed[j] || ok == nil {
			continue
		}
		var d []*Node
		collectDisposed(ok, &d)
		disposed = append(disposed, d...)
		patches = append(patches, deletePatches(ok)...)
	}
	return retained, patches, disposed
}

func findMatch(oldKids []*Node, consumed []bool, want *Node) int {
	for j, ok := range oldKids {
		if consumed[j] || ok == nil {
			continue
		}
		if keysMatch(ok, want) {
			return j
		}
	}
	return -1
}
