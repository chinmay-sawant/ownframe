package watch

import "time"

// ToggleLume flips Super-LumiNova luminescence night mode.
func (a *App) ToggleLume() {
	a.lume = !a.lume
	a.view.Lume = a.lume
	if a.lume {
		a.view.LumeClass = "lume-on"
		return
	}
	a.view.LumeClass = "lume-off"
}

// Lume reports whether luminescence mode is active.
func (a *App) Lume() bool {
	return a.lume
}

// Chrono returns the chronograph movement.
func (a *App) Chrono() *Chrono {
	return &a.chrono
}

// Reserve returns the mainspring power reserve state.
func (a *App) Reserve() *PowerReserve {
	return &a.reserve
}

// SetNow overrides the time function for deterministic tests.
func (a *App) SetNow(fn func() time.Time) {
	a.now = fn
}
