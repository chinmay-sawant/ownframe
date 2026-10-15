package page

// dirtyFromDisplay adds the difference between prev and display to the
// dirty region. The first draw and a size change dirty the frame. With Perf
// on it also stores the changed, dirty, and region counters Stats reports;
// TakeDirty clears only the region, never these counters.
func (p *Page) dirtyFromDisplay(prev, display *Display) {
	if p.dirtyFull {
		return
	}

	if prev == nil {
		p.markFull()
		if p.perf {
			p.recordFull(len(display.Ops))
		}
		p.lastPatch = DisplayPatch{
			Changed: len(display.Ops),
			Added:   len(display.Ops),
			Regions: 1,
			Full:    true,
		}

		return
	}

	d, full, changed := diffDisplay(prev, display)
	if full {
		p.markFull()
		if p.perf {
			p.recordFull(len(display.Ops))
		}
		p.lastPatch = DisplayPatch{
			Changed: changed,
			Added:   d.added,
			Removed: d.removed,
			Regions: 1,
			Full:    true,
		}

		return
	}

	for i := 0; i < d.n; i++ {
		p.markRect(d.rects[i])
	}

	if p.perf {
		p.recordPartial(changed, d.bounds(), display)
	}

	regionCount := d.n
	if regionCount == 0 {
		regionCount = 0
	}
	p.lastPatch = DisplayPatch{
		Changed: changed,
		Added:   d.added,
		Removed: d.removed,
		Regions: regionCount,
		Full:    false,
	}
}

// applyPending marks the new box of every pending id, so an element that
// grew or shrank repaints the space it now uses too, then clears the map
// for reuse instead of dropping it.
func (p *Page) applyPending() {
	if p.pending == nil {
		p.pending = map[string]bool{}
	}
	for id := range p.pending {
		p.markElement(id)
	}

	clear(p.pending)
}
