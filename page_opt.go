package ownframe

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or off.
func (p *Page) SetPerf(on bool) { innerPage(p).SetPerf(on) }

// Perf reports whether pipeline timing and dirty counters are recorded.
func (p *Page) Perf() bool { return innerPage(p).Perf() }

// SetDevTools turns the window devtools overlay on or off.
func (p *Page) SetDevTools(on bool) { innerPage(p).SetDevTools(on) }

// DevTools reports whether the window devtools overlay starts on.
func (p *Page) DevTools() bool { return innerPage(p).DevTools() }

// Stats returns one snapshot of the page counters an inspector shows.
func (p *Page) Stats() Stats { return innerPage(p).Stats() }
