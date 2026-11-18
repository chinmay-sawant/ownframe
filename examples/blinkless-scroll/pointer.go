package main

// PointerProgress is where the pointer sits inside the box on the axis:
// 0 at the near edge, 1 at the far edge. inside is false when the pointer is
// outside the box on that axis; progress then clamps to the nearest edge.
func PointerProgress(px, py float64, box Rect, axis Axis) (progress float64, inside bool) {
	p := px
	if axis == AxisY {
		p = py
	}

	start, size := box.start(axis), box.size(axis)
	if size <= 0 {
		return 0, false
	}

	progress = clamp01((p - start) / size)

	return progress, p >= start && p <= start+size
}

// NearestPointerBox picks the node whose pointer timeline the pointer is
// driving: the deepest node under the pointer that declares a pointer
// timeline, skipping inert nodes. It returns -1 when none matches.
func NearestPointerBox(nodes []Node, px, py float64) int {
	best := -1

	for i := range nodes {
		if !hasKind(nodes[i], KindPointer) || InertAt(nodes, i) || !contains(nodes[i].Rect, px, py) {
			continue
		}

		if best < 0 || depth(nodes, i) >= depth(nodes, best) {
			best = i
		}
	}

	return best
}

func hasKind(n Node, kind string) bool {
	for _, tl := range n.Timelines {
		if tl.Kind == kind {
			return true
		}
	}

	return false
}

func contains(r Rect, x, y float64) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func depth(nodes []Node, i int) int {
	d := 0
	for p := nodes[i].Parent; p >= 0 && p < len(nodes); p = nodes[p].Parent {
		d++
	}

	return d
}
