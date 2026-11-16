package main

// ResolveTimeline finds the timeline a name means for the node at index from.
//
// With timeline-scope off, a name is visible to the declaring node and its
// descendants, and the nearest declaration wins. A node with timeline-scope
// all, or naming the name, hoists every declaration of that name below it, so
// any descendant of that node (a sibling of the declarer, for example) finds
// it too. When several nodes below the scope node declare the name, the first
// in document order wins.
func ResolveTimeline(nodes []Node, from int, name string) (Timeline, int, bool) {
	for at := from; at >= 0 && at < len(nodes); at = nodes[at].Parent {
		if tl, ok := declared(nodes[at], name); ok {
			return tl, at, true
		}

		if nodes[at].Scope.covers(name) {
			for i := range nodes {
				if i == at || !isAncestor(nodes, at, i) {
					continue
				}

				if tl, ok := declared(nodes[i], name); ok {
					return tl, i, true
				}
			}
		}
	}

	return Timeline{}, -1, false
}

func declared(n Node, name string) (Timeline, bool) {
	for _, tl := range n.Timelines {
		if tl.Name == name {
			return tl, true
		}
	}

	return Timeline{}, false
}

// isAncestor reports whether nodes[anc] is a strict ancestor of nodes[i].
func isAncestor(nodes []Node, anc, i int) bool {
	for p := nodes[i].Parent; p >= 0 && p < len(nodes); p = nodes[p].Parent {
		if p == anc {
			return true
		}
	}

	return false
}
