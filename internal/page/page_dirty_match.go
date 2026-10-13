package page

// diffByPosition pairs operations by kind and geometry when the two lists
// differ in length. An operation with no partner dirties its bounds, and
// changed counts every unpaired or repainted operation.
func diffByPosition(prev, next *Display) (dirtySet, bool, int) {
	queue := make(map[opKey][]int, len(next.Ops))
	for i := range next.Ops {
		k := keyOf(&next.Ops[i])
		queue[k] = append(queue[k], i)
	}

	seen := make([]bool, len(next.Ops))
	d := newDirtySet(next.Width, next.Height)
	changed := 0

	for i := range prev.Ops {
		k := keyOf(&prev.Ops[i])
		list := queue[k]

		if len(list) == 0 {
			changed++
			d.add(opBounds(&prev.Ops[i], prev))
			if d.full {
				return d, true, changed
			}

			continue
		}

		j := list[0]
		queue[k] = list[1:]
		seen[j] = true

		if sameOp(&prev.Ops[i], &next.Ops[j]) {
			continue
		}

		changed++
		d.add(opBounds(&prev.Ops[i], prev).Union(opBounds(&next.Ops[j], next)))
		if d.full {
			return d, true, changed
		}
	}

	for j := range next.Ops {
		if seen[j] {
			continue
		}

		changed++
		d.add(opBounds(&next.Ops[j], next))
		if d.full {
			return d, true, changed
		}
	}

	return d, false, changed
}
