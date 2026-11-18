package main

// InertAt reports whether the node at i is inert: it or any ancestor has
// interactivity inert.
func InertAt(nodes []Node, i int) bool {
	for at := i; at >= 0 && at < len(nodes); at = nodes[at].Parent {
		if nodes[at].Inert {
			return true
		}
	}

	return false
}

// HitTest returns the deepest, latest node under the point that is not inert,
// or -1. An inert node is invisible to the pointer, and so are its
// descendants, so a click goes to whatever sits behind them in document
// order.
func HitTest(nodes []Node, x, y float64) int {
	best := -1

	for i := range nodes {
		if InertAt(nodes, i) || !contains(nodes[i].Rect, x, y) {
			continue
		}

		best = i
	}

	return best
}

// NextFocus returns the next focusable, non-inert node after from (use -1 to
// start), wrapping once. dir is +1 or -1. It returns -1 when none exists.
func NextFocus(nodes []Node, from, dir int) int {
	n := len(nodes)
	if n == 0 || (dir != 1 && dir != -1) {
		return -1
	}

	at := from
	for range n {
		at = ((at+dir)%n + n) % n
		if nodes[at].Focusable && !InertAt(nodes, at) {
			return at
		}
	}

	return -1
}
