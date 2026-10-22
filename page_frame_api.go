package ownframe

import "time"

// WantFrameAt declares the next time the page needs the window to draw.
func (p *Page) WantFrameAt(at time.Time) { innerPage(p).WantFrameAt(at) }

// WantFrameIn declares the next frame in d.
func (p *Page) WantFrameIn(d time.Duration) { innerPage(p).WantFrameIn(d) }
