package watch

import (
	"fmt"
	"time"
)

// Reset clears the counter to zero when stopped, or performs flyback if running.
func (c *Chrono) Reset(now time.Time) {
	if c.running {
		c.start = now
		c.accum = 0
		c.lap = 0
		c.lapActive = false
		return
	}
	c.accum = 0
	c.lap = 0
	c.lapActive = false
}

// Lap freezes or unfreezes split timing while running.
func (c *Chrono) Lap(now time.Time) {
	if !c.running {
		return
	}
	if c.lapActive {
		c.lapActive = false
		return
	}
	c.lap = c.Elapsed(now)
	c.lapActive = true
}

// Format returns mm:ss.cc display format for chronograph readout.
func (c *Chrono) Format(now time.Time) string {
	d := c.DisplayDuration(now)
	totalSec := d.Seconds()
	if totalSec < 0 {
		totalSec = 0
	}
	mins := int(totalSec) / 60
	secs := int(totalSec) % 60
	centis := int(d.Milliseconds()/10) % 100
	return fmt.Sprintf("%02d:%02d.%02d", mins, secs, centis)
}
