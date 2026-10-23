package pagetest

import "github.com/chinmay-sawant/ownframe"

// BoxByID returns the last box whose id matches.
func BoxByID(p *ownframe.Page, id string) (ownframe.Box, bool) {
	return lookup(p, func(b ownframe.Box) bool { return b.ID == id })
}

// BoxByAction returns the last box whose data-action matches.
func BoxByAction(p *ownframe.Page, action string) (ownframe.Box, bool) {
	return lookup(p, func(b ownframe.Box) bool { return b.Action == action })
}

func lookup(p *ownframe.Page, match func(ownframe.Box) bool) (ownframe.Box, bool) {
	if p == nil {
		return ownframe.Box{}, false
	}
	var found ownframe.Box
	ok := false
	for _, b := range p.Boxes() {
		if match(b) {
			found = b
			ok = true
		}
	}
	return found, ok
}
