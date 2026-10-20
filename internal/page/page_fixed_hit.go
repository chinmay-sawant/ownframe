package page

// hitItem is one box in the hit list, with the space its rectangle lives in.
// A box the window pins to the viewport is tested in viewport coordinates,
// because the window paints it there while the content below scrolls under it.
type hitItem struct {
	box   Box
	fixed bool
}

// fixedBoxes returns the boxes the window pins to the viewport, such as a
// fixed header, in document order.
func fixedBoxes(d *Display) []Box {
	if d == nil {
		return nil
	}
	seen := map[int]bool{}
	var out []Box
	for i, op := range d.Ops {
		if !op.Fixed {
			continue
		}
		owner := -1
		if i < len(d.OpOwners) {
			owner = d.OpOwners[i]
		}
		if owner < 0 || owner >= len(d.Boxes) || seen[owner] {
			continue
		}
		seen[owner] = true
		out = append(out, d.Boxes[owner])
	}
	return out
}

// hitList builds the hit list in document order, so the innermost box under
// a point still wins when a fixed box overlaps the content.
func hitList(boxes []Box, fixed []Box) []hitItem {
	pinned := make(map[string]bool, len(fixed))
	for _, b := range fixed {
		if b.ID != "" {
			pinned[b.ID] = true
		}
	}
	out := make([]hitItem, 0, len(boxes))
	for _, b := range boxes {
		out = append(out, hitItem{box: b, fixed: pinned[b.ID]})
	}
	return out
}

// scanHit returns the innermost box under a point.
func (p *Page) scanHit(x, y float64) boxSpot {
	vx, vy := x-float64(p.windowing.x), y-float64(p.windowing.y)
	spot := boxSpot{}
	for _, item := range p.hits {
		px, py := x, y
		if item.fixed {
			px, py = vx, vy
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
	return spot
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
