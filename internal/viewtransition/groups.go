package viewtransition

func setParents(groups []group, old, next map[string]node) {
	for i := range groups {
		g := &groups[i]
		nodes := old
		if g.info.NewPresent {
			nodes = next
		}
		byIndex := make(map[int]string, len(nodes))
		for name, item := range nodes {
			byIndex[item.index] = name
		}
		n, ok := nodes[g.info.Name]
		if !ok {
			continue
		}
		for p, steps := n.box.ParentIndex, 0; p >= 0 && p < len(n.boxes) && steps < len(n.boxes); steps++ {
			name, named := byIndex[p]
			if named && groupPresent(groups, name) && acceptsParent(g.info.Mode, name, nodes[name].box.ViewTransitionGroup) {
				g.info.Parent = name
				break
			}
			p = n.boxes[p].ParentIndex
		}
	}
}

func groupPresent(groups []group, name string) bool {
	for _, g := range groups {
		if g.info.Name == name {
			return true
		}
	}
	return false
}

func acceptsParent(targetMode, name, parentMode string) bool {
	switch targetMode {
	case "nearest":
		return true
	case "normal", "contain":
		return mode(parentMode) != "normal"
	default:
		return name == targetMode
	}
}
