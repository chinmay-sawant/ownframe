package ownframe

import "time"

// WantFrameAt declares the next time the page needs the window to draw.
func (page *Page) WantFrameAt(at time.Time) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameAt(at)
}

// WantFrameIn declares the next frame in d.
func (page *Page) WantFrameIn(d time.Duration) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameIn(d)
}
