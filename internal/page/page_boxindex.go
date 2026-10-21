package page

// fillBoxIndex maps each box id to the last box laid out with that id, reusing
// an existing map when one is available. An empty id is indexed like any
// other so an anonymous box still answers a lookup for "".
func fillBoxIndex(ix map[string]Box, boxes []Box) map[string]Box {
	if ix == nil {
		ix = make(map[string]Box, len(boxes))
	} else {
		clear(ix)
	}

	for _, b := range boxes {
		ix[b.ID] = b
	}

	return ix
}
