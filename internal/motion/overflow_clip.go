package motion

import "github.com/chinmay-sawant/blinkless/layout"

// stampTransformClips records the ancestor overflow box on every operation a
// transform player writes. Geometry-baked clipping moves with the sampled
// matrix, so the replay intersects transformed paint with this rect instead.
// Ops are display-local, so a rebind re-stamps the new list; entries that
// write no matrix are left alone.
func stampTransformClips(d *layout.Display, entries []entry) {
	if d == nil || len(entries) == 0 {
		return
	}

	ppp := d.PointsPerPixel
	if ppp <= 0 {
		ppp = 1
	}

	for i := range entries {
		e := &entries[i]
		if e.mask&fieldMatrix == 0 || len(e.indices) == 0 {
			continue
		}

		for _, index := range e.indices {
			if index < 0 || index >= len(d.Ops) || index >= len(d.OpOwners) {
				continue
			}

			owner := d.OpOwners[index]
			if owner < 0 || owner >= len(d.Boxes) {
				continue
			}

			box := &d.Boxes[owner]
			if !box.HasOverflowClip {
				continue
			}

			c := box.OverflowClip
			d.Ops[index].SetOverflowClip(c[0]*ppp, c[1]*ppp, c[2]*ppp, c[3]*ppp)
		}
	}
}
