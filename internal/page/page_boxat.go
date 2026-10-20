package page

// boxSpot is one point query against the box index: the innermost box under
// the point and the innermost id-bearing box under it.
type boxSpot struct {
	box Box
	ok  bool
	id  string
}

// scanBoxes walks document-order boxes once and settles both winners, so
// Click, Press, and Hover cannot disagree on overlapping boxes.
func scanBoxes(boxes []Box, x, y float64) boxSpot {
	spot := boxSpot{}
	for _, b := range boxes {
		if x < b.X || y < b.Y || x > b.X+b.W || y > b.Y+b.H {
			continue
		}
		spot.box, spot.ok = b, true
		if b.ID != "" {
			spot.id = b.ID
		}
	}
	return spot
}

// boxAt answers the point query against the page's box index. hit and
// boxIDAt delegate here.
func (p *Page) boxAt(x, y float64) boxSpot {
	return scanBoxes(p.boxes, x, y)
}
