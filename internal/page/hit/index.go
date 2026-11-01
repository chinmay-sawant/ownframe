package hit

import "github.com/chinmay-sawant/blinkless/layout"

// Index points each id at the last box laid out with that id, reusing the
// map and the boxes' storage. An empty id is indexed like any other.
func Index(ix map[string]*layout.Box, boxes []layout.Box) map[string]*layout.Box {
	const defaultIndexCap = 16

	if ix == nil {
		capHint := defaultIndexCap
		if len(boxes) < defaultIndexCap {
			capHint = len(boxes)
		}
		ix = make(map[string]*layout.Box, capHint)
	} else {
		clear(ix)
	}

	for i := range boxes {
		ix[boxes[i].ID] = &boxes[i]
	}

	return ix
}

// ByID returns the last box laid out with that id, or a box that carries
// only the id. The lookup is one map read.
func ByID(ix map[string]*layout.Box, id string) layout.Box {
	if b, ok := ix[id]; ok {
		return *b
	}

	return layout.Box{ID: id}
}

// Has reports whether the index has a box for id, in one map read.
func Has(ix map[string]*layout.Box, id string) bool {
	_, ok := ix[id]

	return ok
}
