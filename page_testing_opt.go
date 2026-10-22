package ownframe

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or off.
func (t Testing) SetPerf(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPerf(on)
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (t Testing) Perf() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.Perf()
}

// SetDevTools turns the window devtools overlay on or off.
func (t Testing) SetDevTools(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetDevTools(on)
}

// DevTools reports whether the window devtools overlay starts on.
func (t Testing) DevTools() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.DevTools()
}

// Stats returns one snapshot of the page counters an inspector shows.
func (t Testing) Stats() Stats {
	p := innerPage(t.page)
	if p == nil {
		return Stats{}
	}

	return p.Stats()
}
