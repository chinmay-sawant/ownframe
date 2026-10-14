package displaypatch

import "image"

// Op is one display operation. ID is the operation's stable identity.
type Op struct {
	ID          int
	Bounds      image.Rectangle
	Hash        uint64
	Transformed image.Rectangle
}

// Dominates reports whether a's bounds contain b's bounds.
func (a Op) Dominates(b Op) bool {
	return b.Bounds.In(a.Bounds)
}

// PrecomputeBounds returns the operation's bounds, transformed by dx,dy.
func (o Op) PrecomputeBounds(dx, dy int) image.Rectangle {
	return o.Bounds.Add(image.Pt(dx, dy))
}

// PatchResult summarizes an operation-level diff.
type PatchResult struct {
	Added     []Op
	Removed   []Op
	Changed   []Op
	Unchanged int
}

// Diff compares two operation slices by ID and hash.
func Diff(oldOps, newOps []Op) PatchResult {
	oldByID := map[int]Op{}
	for _, o := range oldOps {
		oldByID[o.ID] = o
	}
	res := PatchResult{}
	seen := map[int]bool{}
	for _, n := range newOps {
		seen[n.ID] = true
		if o, ok := oldByID[n.ID]; ok {
			if o.Hash == n.Hash && o.Bounds.Eq(n.Bounds) {
				res.Unchanged++
			} else {
				res.Changed = append(res.Changed, n)
			}
		} else {
			res.Added = append(res.Added, n)
		}
	}
	for _, o := range oldOps {
		if !seen[o.ID] {
			res.Removed = append(res.Removed, o)
		}
	}
	return res
}
