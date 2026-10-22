package ownframe

// PatchDisplay mutates the retained display and records changed regions.
func (t Testing) PatchDisplay(fn func(d *Display)) DisplayPatch {
	p := innerPage(t.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (t Testing) LastPatch() DisplayPatch {
	p := innerPage(t.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.LastPatch()
}
