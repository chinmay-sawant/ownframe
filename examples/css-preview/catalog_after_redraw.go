package main

// afterRedraw points an active manual sample at the display the page just
// published. CSS animation players rebind themselves; the manual paths keep
// their own operation pointers, so a scroll or a window resize has to bind
// them again.
func (d *demo) afterRedraw() {
	if !d.manualLive() {
		return
	}
	// One bad bind drops that sample's paint rather than the page; the error
	// is not part of a redraw's contract.
	_ = d.bindManual(d.v.Rows)
}
