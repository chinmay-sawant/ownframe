package ownframe

// SetHotReload turns the file watch on or off.
func (t Testing) SetHotReload(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetHotReload(on)
}

// Watching reports whether a poll can reread a source file.
func (t Testing) Watching() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.Watching()
}
