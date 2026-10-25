package page

import "github.com/chinmay-sawant/ownframe/internal/page/hit"

// hitItem is the hit list entry stored on the page. It is hit.Item so the
// page can pass the slice into package hit without copying it.
type hitItem = hit.Item

// boxSpot is one point query against the box index: the innermost box under
// the point and the innermost id-bearing box under it.
type boxSpot struct {
	box Box
	ok  bool
	id  string
}

func spotFrom(s hit.Spot) boxSpot {
	return boxSpot{box: s.Box, ok: s.OK, id: s.ID}
}

// boxAt answers the point query against the page's boxes. Click, Press,
// and Hover all read it, so overlapping boxes have one winner.
func (p *Page) boxAt(x, y float64) boxSpot {
	if len(p.hits) > 0 {
		return spotFrom(hit.ScanList(p.hits, x, y, p.scroll.Windowing.X, p.scroll.Windowing.Y))
	}

	return spotFrom(hit.Scan(p.boxes, x, y))
}

// hitBox returns the box under a content point.
func (p *Page) hitBox(x, y float64) (Box, bool) {
	spot := p.boxAt(x, y)

	return spot.box, spot.ok
}

// isFixed reports whether the window pins the box to the viewport.
func (p *Page) isFixed(box Box) bool {
	return hit.IsFixed(p.hits, box)
}

// fixedBoxes returns the boxes the window pins to the viewport, such as a
// fixed header. The first operation of a fixed range belongs to that box.
func fixedBoxes(d *Display) []Box {
	return hit.FixedBoxes(d)
}
