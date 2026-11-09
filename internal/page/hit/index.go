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

// SameIDs reports whether a and b name the same ids in order. The page uses
// it to gate the index rebuild: an unchanged id set repoints without a
// clear. Pairwise equality implies the same id multiset, so duplicates and
// anonymous boxes stay correct where a map-content comparison would not.
func SameIDs(a, b []layout.Box) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}

// Repoint repoints an index whose key set already matches boxes, without
// clearing it, so the map keeps its capacity and buckets. The caller
// guarantees the match with SameIDs against the list the index was built
// from; a nil map builds a fresh index instead.
func Repoint(ix map[string]*layout.Box, boxes []layout.Box) map[string]*layout.Box {
	if ix == nil {
		return Index(nil, boxes)
	}
	for i := range boxes {
		ix[boxes[i].ID] = &boxes[i]
	}
	return ix
}

// Has reports whether the index has a box for id, in one map read.
func Has(ix map[string]*layout.Box, id string) bool {
	_, ok := ix[id]

	return ok
}
