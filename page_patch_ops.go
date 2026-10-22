package ownframe

// PatchDisplay mutates the retained display and records changed regions.
func (p *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	return innerPage(p).PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (p *Page) LastPatch() DisplayPatch { return innerPage(p).LastPatch() }
