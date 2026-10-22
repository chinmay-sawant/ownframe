package ownframe

// PatchOp mutates one retained operation.
// Spotify-player and depth-alert patch ops from the tick.
func (p *Page) PatchOp(i int, fn func(o *DisplayOp)) bool {
	return innerPage(p).PatchOp(i, fn)
}

// PatchDisplay mutates the retained display and records changed regions.
func (p *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	return innerPage(p).PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (p *Page) LastPatch() DisplayPatch { return innerPage(p).LastPatch() }
