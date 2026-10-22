package page

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
