package page

// newBoxIndex maps each box id to the last box laid out with that id, so a
// lookup is one map read instead of a walk over every box. setBoxes rebuilds
// it whenever p.boxes is replaced, and that is the only place boxes change.
// An empty id is indexed like any other so an anonymous box still answers a
// lookup for "".
func newBoxIndex(boxes []Box) map[string]Box {
	ix := make(map[string]Box, len(boxes))

	for _, b := range boxes {
		ix[b.ID] = b
	}

	return ix
}
