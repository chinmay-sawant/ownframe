package page

// PatchDisplay mutates the retained display and records the regions that
// changed. It returns the patch report. A nil display reports nothing.
func (p *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	if p.display == nil || fn == nil {
		return p.lastPatch
	}

	buf := p.rentOps(len(p.display.Ops))
	buf = append(buf, p.display.Ops...)
	before := *p.display
	before.Ops = buf

	fn(p.display)

	d, full, changed := diffDisplay(&before, p.display)
	if full {
		p.markFull()
	} else {
		for i := 0; i < d.n; i++ {
			p.markRect(d.rects[i])
		}
	}

	patch := DisplayPatch{
		Changed: changed,
		Added:   d.added,
		Removed: d.removed,
		Regions: d.n,
		Full:    full,
	}
	if patch.Regions == 0 && !full {
		patch.Regions = 0
	}
	if full {
		patch.Regions = 1
	}

	patch = p.withChangedProps(patch)
	p.lastPatch = patch
	p.freeOps(buf)
	return patch
}

// PatchOp mutates one retained operation and dirties its old and new bounds.
func (p *Page) PatchOp(i int, fn func(o *DisplayOp)) bool {
	if p.display == nil || i < 0 || i >= len(p.display.Ops) || fn == nil {
		return false
	}

	old := opBounds(&p.display.Ops[i], p.display)
	fn(&p.display.Ops[i])
	nw := opBounds(&p.display.Ops[i], p.display)
	p.markRect(old.Union(nw))
	p.lastPatch = p.withChangedProps(DisplayPatch{Changed: 1, Regions: 1, Full: false})
	return true
}
