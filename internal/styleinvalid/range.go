package styleinvalid

// DirtyRange records a contiguous range of style dirty boxes (by index into
// a flattened box list) so page stats can report it.
type DirtyRange struct {
	Start, End int
}

// RangesFromChanges converts changes to index ranges using boxIndex. Changes
// without an index are dropped.
func RangesFromChanges(changes []Change, boxIndex map[string]int, size int) []DirtyRange {
	if size <= 0 {
		return nil
	}
	seen := make([]bool, size)
	for _, c := range changes {
		if i, ok := boxIndex[c.BoxID]; ok && i >= 0 && i < size {
			seen[i] = true
		}
	}
	var ranges []DirtyRange
	i := 0
	for i < size {
		if !seen[i] {
			i++
			continue
		}
		start := i
		for i < size && seen[i] {
			i++
		}
		ranges = append(ranges, DirtyRange{Start: start, End: i})
	}
	return ranges
}
