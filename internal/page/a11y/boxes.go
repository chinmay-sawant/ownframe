package a11y

import "github.com/chinmay-sawant/blinkless/layout"

// UniqueBoxes indexes boxes by id and counts duplicates.
func UniqueBoxes(boxes []layout.Box) (map[string]layout.Box, map[string]int) {
	byID, counts := map[string]layout.Box{}, map[string]int{}
	for _, box := range boxes {
		if box.ID != "" {
			byID[box.ID] = box
			counts[box.ID]++
		}
	}
	return byID, counts
}
