package ownframe

import "context"

// PollReload rereads a watched source file.
// The reload example tests the file watch from app code.
func (p *Page) PollReload(ctx context.Context) (bool, error) {
	return innerPage(p).PollReload(ctx)
}

// SetHotReload turns the file watch on or off.
func (p *Page) SetHotReload(on bool) { innerPage(p).SetHotReload(on) }

// Watching reports whether a poll can reread a source file.
func (p *Page) Watching() bool { return innerPage(p).Watching() }
