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
