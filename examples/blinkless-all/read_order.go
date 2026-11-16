package main

import (
	"sort"
	"strconv"
)

// readingSequence returns the item indices in reading order.
//
// Without a declaration the order is the document order. A container whose
// reading-flow is not normal reorders its direct children: source-order keeps
// them as written, flex-visual and grid-rows read them by row and then by
// column, flex-flow and grid-columns read by column and then by row (flex-flow
// follows the longer axis of the children), and grid-order keeps the source
// order because the display list does not carry the order property. Inside a
// reordered container reading-order then sorts the children by number, lowest
// first, and ties keep the flow order. A child is followed by its own
// descendants. The display list lists flex items in layout order, so a flex
// container already arrives in visual order and flex-visual cannot change it.
func (m navModel) readingSequence() []int {
	kids := map[int][]int{}

	var roots []int

	for i, it := range m.Items {
		if it.Parent < 0 || it.Parent >= len(m.Items) {
			roots = append(roots, i)

			continue
		}

		kids[it.Parent] = append(kids[it.Parent], i)
	}

	out := make([]int, 0, len(m.Items))

	var walk func(list []int)

	walk = func(list []int) {
		for _, i := range list {
			out = append(out, i)
			walk(m.readingChildren(i, kids[i]))
		}
	}

	walk(roots)

	return out
}

// readingChildren orders the direct children of one container.
func (m navModel) readingChildren(parent int, children []int) []int {
	flow := m.Items[parent].Declared["reading-flow"]
	if flow == "" || flow == "normal" || len(children) < 2 {
		return children
	}

	ordered := append([]int(nil), children...)

	switch flow {
	case "flex-visual", "grid-rows":
		m.sortVisual(ordered, true)
	case "grid-columns":
		m.sortVisual(ordered, false)
	case "flex-flow":
		m.sortVisual(ordered, m.spreadX(ordered) >= m.spreadY(ordered))
	}

	sort.SliceStable(ordered, func(a, b int) bool {
		return m.readingNumber(ordered[a]) < m.readingNumber(ordered[b])
	})

	return ordered
}

// readingNumber is the reading-order of item i, 0 when not declared.
func (m navModel) readingNumber(i int) int {
	n, err := strconv.Atoi(m.Items[i].Declared["reading-order"])
	if err != nil {
		return 0
	}

	return n
}

// sortVisual orders items by position. rowsFirst reads left to right inside a
// row of items whose tops are within half a row height of each other.
func (m navModel) sortVisual(list []int, rowsFirst bool) {
	sort.SliceStable(list, func(a, b int) bool {
		ia, ib := m.Items[list[a]], m.Items[list[b]]

		majorA, majorB, minorA, minorB := ia.Y, ib.Y, ia.X, ib.X
		band := (ia.H + ib.H) * readBandFraction

		if !rowsFirst {
			majorA, majorB, minorA, minorB = ia.X, ib.X, ia.Y, ib.Y
			band = (ia.W + ib.W) * readBandFraction
		}

		if d := majorA - majorB; d > band || d < -band {
			return majorA < majorB
		}

		return minorA < minorB
	})
}

func (m navModel) spreadX(list []int) float64 {
	return m.spread(list, func(it navItem) float64 { return it.X })
}
func (m navModel) spreadY(list []int) float64 {
	return m.spread(list, func(it navItem) float64 { return it.Y })
}

func (m navModel) spread(list []int, f func(navItem) float64) float64 {
	lo, hi := f(m.Items[list[0]]), f(m.Items[list[0]])
	for _, i := range list {
		v := f(m.Items[i])
		lo, hi = min(lo, v), max(hi, v)
	}

	return hi - lo
}

// readTopLayer says whether an element is drawn in the top layer. An open
// popover or dialog is; a closed one that is still animating out stays there
// while overlay is auto, and leaves at once when overlay is none.
func readTopLayer(open bool, overlay string) bool {
	return open || overlay == "auto"
}

// readBackgroundShift is how far a background image sits from where it would
// sit with scroll, in CSS pixels along y. A fixed layer stays on the viewport,
// so it moves against the page scroll; a local layer moves with the box's own
// scroll; a scroll layer does not move.
func readBackgroundShift(attachment string, pageScroll, boxScroll float64) float64 {
	switch attachment {
	case "fixed":
		return pageScroll
	case "local":
		return -boxScroll
	}

	return 0
}

// readBandFraction is the share of the mean box size that counts as one row.
const readBandFraction = 0.25
