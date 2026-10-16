package watch

import "time"

// Chrono models a dual-pusher mechanical chronograph with column-wheel action.
type Chrono struct {
	running   bool
	start     time.Time
	accum     time.Duration
	lap       time.Duration
	lapActive bool
}

// Running reports whether the chronograph movement is actively timing.
func (c *Chrono) Running() bool {
	return c.running
}

// Elapsed returns the total recorded duration up to the given time.
func (c *Chrono) Elapsed(now time.Time) time.Duration {
	if !c.running {
		return c.accum
	}
	return c.accum + now.Sub(c.start)
}

// DisplayDuration returns the visible time (lap time when split is active).
func (c *Chrono) DisplayDuration(now time.Time) time.Duration {
	if c.lapActive {
		return c.lap
	}
	return c.Elapsed(now)
}

// Start engages the chronograph clutch.
func (c *Chrono) Start(now time.Time) {
	if !c.running {
		c.start = now
		c.running = true
		c.lapActive = false
	}
}

// Stop disengages the chronograph clutch and halts accumulation.
func (c *Chrono) Stop(now time.Time) {
	if c.running {
		c.accum += now.Sub(c.start)
		c.running = false
	}
}

// Toggle flips chronograph operation between start and stop.
func (c *Chrono) Toggle(now time.Time) {
	if c.running {
		c.Stop(now)
	} else {
		c.Start(now)
	}
}
