package window

import "github.com/chinmay-sawant/blinkless/layout"

// devHover records the box under the cursor without calling the page.
func (s *shell) devHover(px, py float64) {
	box, ok := s.boxAt(px, py)
	s.dev.hovered, s.dev.haveHov = box, ok
}

// devBoxAt returns the last box that contains x, y, the innermost element.
func devBoxAt(boxes []layout.Box, x, y float64) (layout.Box, bool) {
	var found layout.Box
	ok := false

	for _, b := range boxes {
		if x < b.X || y < b.Y || x > b.X+b.W || y > b.Y+b.H {
			continue
		}

		found = b
		ok = true
	}

	return found, ok
}

// devSameBox reports whether two picks name the same element.
func devSameBox(a, b layout.Box) bool {
	if a.ID != "" || b.ID != "" {
		return a.ID == b.ID
	}

	return a.Tag == b.Tag && a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}
