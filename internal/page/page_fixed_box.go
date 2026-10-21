package page

// fixedBoxes returns the boxes the window pins to the viewport, such as a
// fixed header. The engine marks the whole operation range of a fixed box,
// so that range can hold an operation an ancestor owns, such as the canvas
// fill. The first operation of the range belongs to the fixed box itself,
// and an operation an ancestor owns belongs to no fixed box at all.
func fixedBoxes(d *Display) []Box {
	if d == nil {
		return nil
	}
	seen := map[int]bool{}
	var out []Box
	owner := -1
	for i, op := range d.Ops {
		if !op.Fixed {
			owner = -1
			continue
		}
		next := opOwner(d, i)
		if next < 0 || next >= len(d.Boxes) {
			continue
		}
		if owner < 0 {
			owner = next
		} else if !insideBox(d.Boxes, next, owner) && !aboveBox(d.Boxes, next, owner) {
			owner = next
		}
		if seen[owner] {
			continue
		}
		seen[owner] = true
		out = append(out, d.Boxes[owner])
	}
	return out
}

// opOwner returns the box that owns an operation, or -1.
func opOwner(d *Display, index int) int {
	if index < 0 || index >= len(d.OpOwners) {
		return -1
	}
	return d.OpOwners[index]
}
