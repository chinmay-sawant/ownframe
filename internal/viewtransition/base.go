package viewtransition

import "github.com/chinmay-sawant/ownframe"

func baseOps(d *ownframe.Display, groups []group) []int {
	hidden := make(map[int]bool)
	for _, g := range groups {
		for _, i := range g.newOps {
			hidden[i] = true
		}
	}
	var out []int
	for i := range d.Ops {
		if !hidden[i] {
			out = append(out, i)
		}
	}
	return out
}

func orderedSubset(d *ownframe.Display, index map[int]int) []int {
	order := d.Order
	if len(order) == 0 {
		order = make([]int, len(d.Ops))
		for i := range order {
			order[i] = i
		}
	}
	out := make([]int, 0, len(index))
	for _, old := range order {
		if v, ok := index[old]; ok {
			out = append(out, v)
		}
	}
	return out
}

func subset(d *ownframe.Display, keep []int) *ownframe.Display {
	out := *d
	out.Ops = make([]ownframe.DisplayOp, 0, len(keep))
	out.OpOwners = make([]int, 0, len(keep))
	index := make(map[int]int, len(keep))
	for _, old := range keep {
		index[old] = len(out.Ops)
		out.Ops = append(out.Ops, d.Ops[old].Clone())
		out.OpOwners = append(out.OpOwners, ownerAt(d, old))
	}
	out.Order = orderedSubset(d, index)
	out.Boxes = subsetBoxes(d.Boxes, out.Ops)
	return &out
}

func lerpRect(a, b Rect, p float64) Rect {
	f := func(x, y float64) float64 { return x + (y-x)*p }
	return Rect{f(a.X, b.X), f(a.Y, b.Y), f(a.W, b.W), f(a.H, b.H)}
}
