package pagetest

import "github.com/chinmay-sawant/ownframe"

// BoxByID returns the last box whose id matches.
func BoxByID(p *ownframe.Page, id string) (ownframe.Box, bool) {
	return lookup(p, func(b ownframe.Box) bool { return b.ID == id })
}

// BoxByAction returns the last box whose data-action matches. An empty action
// matches nothing.
func BoxByAction(p *ownframe.Page, action string) (ownframe.Box, bool) {
	if action == "" {
		return ownframe.Box{}, false
	}

	return lookup(p, func(b ownframe.Box) bool { return b.Action == action })
}

// lookup scans the committed display's boxes. It reads the display slice
// directly instead of Page.Boxes, which copies every box and makes a lookup on
// a large page cost more than the click it is preparing.
func lookup(p *ownframe.Page, match func(ownframe.Box) bool) (ownframe.Box, bool) {
	if p == nil {
		return ownframe.Box{}, false
	}
	var found ownframe.Box
	ok := false
	if display := p.Probe().Display(); display != nil {
		for _, b := range display.Boxes {
			if match(b) {
				found = b
				ok = true
			}
		}
		return found, ok
	}
	// A bitmap-fallback page has no retained display; Boxes still answers.
	for _, b := range p.Boxes() {
		if match(b) {
			found = b
			ok = true
		}
	}
	return found, ok
}
