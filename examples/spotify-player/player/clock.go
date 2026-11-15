package player

import "time"

// virtualFreezeTicks is how many consecutive frames a live engine position
// may stand still while playing before the tick treats the engine as
// frozen. A real voice advances every frame, so this only trips for
// engines with no advancing clock (headless -web Serve has no audio
// device) or a stalled decode. At the 60 Hz tick it is half a second.
const virtualFreezeTicks = 30

// clock reports the playback position and duration. While the engine
// advances on its own it returns the engine values verbatim. When the
// engine is frozen or unusable but playback is wanted, it advances a
// virtual position on the wall clock so the seek bar, the time labels,
// and the tier-1 tick patches keep moving. Call once per tick; it owns
// the freeze bookkeeping. live is false while the virtual clock drives.
func (a *App) clock() (pos, dur time.Duration, live bool) {
	now := time.Now()
	dt := now.Sub(a.vlast)
	if dt < 0 || dt > time.Second {
		dt = 0
	}
	a.vlast = now

	epos, edur := a.audio.Position(), a.audio.Duration()
	if a.audio.Loading() || a.audio.Err() != nil || edur <= 0 || a.seconds <= 0 {
		if a.vengaged && a.view.Playing {
			a.vpos = advancePos(a.vpos, dt, a.vdur())
		}

		return a.vpos, a.vdur(), false
	}

	if epos != a.vlastPos {
		a.vlastPos = epos
		a.vfreeze = 0
		a.vengaged = false
		a.vpos = epos

		return epos, edur, true
	}

	if a.view.Playing {
		if a.vengaged {
			a.vpos = advancePos(a.vpos, dt, a.vdur())

			return a.vpos, a.vdur(), false
		}

		a.vfreeze++
		if a.vfreeze >= virtualFreezeTicks {
			a.vengaged = true
			a.vpos = epos

			return epos, edur, false
		}
	} else {
		a.vfreeze = 0
	}

	return epos, edur, true
}

// vdur is the virtual track length from the view metadata.
func (a *App) vdur() time.Duration {
	return time.Duration(a.seconds) * time.Second
}

// advancePos moves p forward by dt, clamped to [0, dur].
func advancePos(p, dt, dur time.Duration) time.Duration {
	p += dt
	if p < 0 {
		return 0
	}
	if p > dur {
		return dur
	}

	return p
}

// noteSeek arms the virtual clock at fraction immediately, so the next tick
// paints from the clicked spot instead of waiting for a frozen engine to
// move. The engine seek runs alongside; when it reports motion the virtual
// clock snaps to it and disengages.
func (a *App) noteSeek(fraction float64) {
	if fraction < 0 {
		fraction = 0
	}

	if fraction > 1 {
		fraction = 1
	}

	a.vpos = time.Duration(fraction * float64(a.vdur()))
	a.vengaged = true
	a.vfreeze = 0
}

// advanceVirtual moves the virtual position by an explicit delta. Tests use
// this instead of waiting on the wall clock.
func (a *App) advanceVirtual(dt time.Duration) {
	if a.view.Playing && a.vdur() > 0 {
		a.vpos = advancePos(a.vpos, dt, a.vdur())
	}
}

// virtualEnded reports end-of-track for the virtual clock.
func (a *App) virtualEnded() bool {
	return a.vengaged && a.vdur() > 0 && a.vpos >= a.vdur()
}
