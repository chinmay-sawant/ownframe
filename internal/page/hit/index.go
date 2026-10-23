package hit

import "github.com/chinmay-sawant/blinkless/layout"

// Index maps each box id to the last box laid out with that id, reusing
// an existing map when one is available. An empty id is indexed like any
// other so an anonymous box still answers a lookup for "".
func Index(ix map[string]layout.Box, boxes []layout.Box) map[string]layout.Box {
	if ix == nil {
		ix = make(map[string]layout.Box, len(boxes))
	} else {
		clear(ix)
	}

	for _, b := range boxes {
		ix[b.ID] = b
	}

	return ix
}

// ByID returns the last box laid out with that id, or a box that carries
// only the id. The lookup is one map read.
func ByID(ix map[string]layout.Box, id string) layout.Box {
	if b, ok := ix[id]; ok {
		return b
	}

	return layout.Box{ID: id}
}

// Has reports whether the index has a box for id, in one map read.
func Has(ix map[string]layout.Box, id string) bool {
	_, ok := ix[id]

	return ok
}
