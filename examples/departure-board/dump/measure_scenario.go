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

// plan describes one scenario's draws, its mid action, and its tick count.
// midAt runs the mid action after that many ticks, so the flap cascade runs
// against flaps that were already bound, the shape a real 12 second rotation
// produces. mid is "advance" for the SetData and Redraw resync or "rotate"
// for the tick's retarget path.
type plan struct {
	draw    bool
	advance bool
	mid     string
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
		// Bind the flaps, then resync with Advance, so the row bands run
		// after the Redraw shape.
		return plan{draw: true, ticks: 150, midAt: 3, mid: "advance"}, true
	case "rotate":
		// Bind the flaps, then take the tick's retarget path, which must
		// leave the page generation alone.
		return plan{draw: true, ticks: 150, midAt: 3, mid: "rotate"}, true
	default:
		return plan{}, false
	}
}
