package element

func createPatches(n *Node) PatchList {
	if n == nil {
		return nil
	}
	var patches PatchList
	patches = append(patches, Patch{Op: OpCreate, Node: n})
	for _, c := range n.Children {
		patches = append(patches, createPatches(c)...)
	}
	return patches
}

func deletePatches(n *Node) PatchList {
	if n == nil {
		return nil
	}
	var patches PatchList
	for _, c := range n.Children {
		patches = append(patches, deletePatches(c)...)
	}
	patches = append(patches, Patch{Op: OpDelete, Node: n})
	n.Dispose()
	return patches
}

func cloneAttrs(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneMeta(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneNode(n *Node) *Node {
	if n == nil {
		return nil
	}
	out := NewNode(n.ID, n.Kind)
	out.Key = n.Key
	out.Attrs = cloneAttrs(n.Attrs)
	out.Metadata = cloneMeta(n.Metadata)
	for _, c := range n.Children {
		child := cloneNode(c)
		out.Append(child)
	}
	return out
}

func collectRetained(n *Node, out *[]*Node) {
	if n == nil {
		return
	}
	*out = append(*out, n)
	for _, c := range n.Children {
		collectRetained(c, out)
	}
}

func collectDisposed(n *Node, out *[]*Node) {
	if n == nil {
		return
	}
	for _, c := range n.Children {
		collectDisposed(c, out)
	}
	*out = append(*out, n)
}
