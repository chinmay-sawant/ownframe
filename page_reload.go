package ownframe

// SetHotReload turns the file watch on or off.
func (p *Page) SetHotReload(on bool) { innerPage(p).SetHotReload(on) }

// Watching reports whether a poll can reread a source file.
func (p *Page) Watching() bool { return innerPage(p).Watching() }
