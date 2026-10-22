package page

// hitItem is one box in the hit list, with the space its rectangle lives in.
// A box the window pins to the viewport is tested in viewport coordinates,
// because the window paints it there while the content below scrolls under it.
type hitItem struct {
	box   Box
	fixed bool
}

// hitList builds the hit list in document order, so the innermost box under
// a point still wins inside one layer. A box the window pins to the viewport
// is marked, because the window paints that layer last and a click has to
// land on the box the user sees on top.
func hitList(boxes []Box, fixed []Box) []hitItem {
	out := make([]hitItem, 0, len(boxes))
	if len(fixed) == 0 {
		for _, b := range boxes {
			out = append(out, hitItem{box: b})
		}
		return out
	}
	pinned := make(map[boxKey]bool, len(fixed))
	for _, b := range fixed {
		pinned[boxKeyOf(b)] = true
	}
	for i, b := range boxes {
		out = append(out, hitItem{box: b, fixed: pinnedBox(boxes, i, pinned)})
	}
	return out
}

// boxKey identifies a box inside the page's box list. An anonymous box has
// no id to match on, so its tag, its parent, and its rectangle name it.
type boxKey struct {
	id     string
	tag    string
	parent int
	x, y   float64
	w, h   float64
}

// boxKeyOf returns the key of a box.
func boxKeyOf(b Box) boxKey {
	return boxKey{
		id:     b.ID,
		tag:    b.Tag,
		parent: b.ParentIndex,
		x:      b.X,
		y:      b.Y,
		w:      b.W,
		h:      b.H,
	}
}

// pinnedBox reports whether the box is painted with the fixed layer: it is
// the fixed box itself, or it sits inside one, because the engine paints
// the whole subtree of a fixed element with it. The window paints that
// layer last, so such a box covers the content below it.
func pinnedBox(boxes []Box, index int, pinned map[boxKey]bool) bool {
	for i := index; i >= 0 && i < len(boxes); {
		if pinned[boxKeyOf(boxes[i])] {
			return true
		}
		next := boxes[i].ParentIndex
		if next >= i {
			return false
		}
		i = next
	}
	return false
}

// insideBox reports whether the box is the other box or sits inside it.
func insideBox(boxes []Box, box, other int) bool {
	for i := box; i >= 0 && i < len(boxes); {
		if i == other {
			return true
		}
		next := boxes[i].ParentIndex
		if next >= i {
			return false
		}
		i = next
	}
	return false
}

// aboveBox reports whether the box is the other box or an ancestor of it.
func aboveBox(boxes []Box, box, other int) bool {
	return insideBox(boxes, other, box)
}

// scanHit returns the box under a point. The window paints the fixed layer
// last, so a fixed box wins over the content it covers and the click lands
// on the box the user sees on top. Inside one layer the innermost box in
// document order wins.
func (p *Page) scanHit(x, y float64) boxSpot {
	vx, vy := x-float64(p.scroll.Windowing.X), y-float64(p.scroll.Windowing.Y)
	content := boxSpot{}
	pinned := boxSpot{}
	for _, item := range p.hits {
		px, py := x, y
		spot := &content
		if item.fixed {
			px, py = vx, vy
			spot = &pinned
		}
		if px < item.box.X || px > item.box.X+item.box.W ||
			py < item.box.Y || py > item.box.Y+item.box.H {
			continue
		}
		spot.box = item.box
		spot.ok = true
		if item.box.ID != "" {
			spot.id = item.box.ID
		}
	}
	if pinned.ok {
		return pinned
	}

	return content
}

// hitBox returns the box under a content point.
func (p *Page) hitBox(x, y float64) (Box, bool) {
	if len(p.hits) == 0 {
		return hit(p.boxes, x, y)
	}
	spot := p.scanHit(x, y)

	return spot.box, spot.ok
}

// isFixed reports whether the window pins the box to the viewport.
func (p *Page) isFixed(box Box) bool {
	if box.ID == "" {
		return false
	}
	for _, item := range p.hits {
		if item.fixed && item.box.ID == box.ID {
			return true
		}
	}
	return false
}
