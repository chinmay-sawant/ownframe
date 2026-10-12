package dump

import "time"

// Pace, when set, sleeps between ticks so a run sees the real frame cadence
// the window produces. Zero ticks as fast as possible for pure page timing.
var Pace time.Duration

// pace sleeps between ticks when Pace is set.
func pace() {
	if Pace > 0 {
		time.Sleep(Pace)
	}
}

// plan describes one scenario's draws and tick count. midAt advances the
// feed again after that many ticks, so the flap cascade runs against flaps
// that were already bound, the shape a real 12 second rotation produces.
type plan struct {
	draw    bool
	advance bool
	midAt   int
	ticks   int
}

// scenarioPlan returns the shape of one scenario.
func scenarioPlan(name string) (plan, bool) {
	switch name {
	case "warmup":
		return plan{draw: true, ticks: 30}, true
	case "flap":
		return plan{draw: true, advance: true, ticks: 120}, true
	case "idle":
		return plan{draw: true, ticks: 240}, true
	case "loop":
		return plan{draw: true, advance: true, ticks: 240}, true
	case "cascade":
		// Bind the flaps, then rotate, so the row bands appear.
		return plan{draw: true, ticks: 150, midAt: 3}, true
	default:
		return plan{}, false
	}
}
