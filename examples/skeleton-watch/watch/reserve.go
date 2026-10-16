package watch

import "math"

// MaxPowerReserve is the maximum duration the mainspring barrel stores (72h).
const MaxPowerReserve = 72.0

// WindStepHours is the power reserve added by one crown winding stroke.
const WindStepHours = 6.0

// PowerReserve tracks mainspring energy, winding, and escapement drain.
type PowerReserve struct {
	hours      float64
	drainSpeed float64
}

func newPowerReserve(initialHours float64) PowerReserve {
	if initialHours < 0 {
		initialHours = 0
	}
	if initialHours > MaxPowerReserve {
		initialHours = MaxPowerReserve
	}
	return PowerReserve{
		hours:      initialHours,
		drainSpeed: 1.0,
	}
}

// Hours returns the remaining stored energy in hours.
func (pr *PowerReserve) Hours() float64 {
	return pr.hours
}

// Percent returns the reserve percentage from 0% to 100%.
func (pr *PowerReserve) Percent() int {
	pct := int(math.Round((pr.hours / MaxPowerReserve) * 100.0))
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

// Depleted reports whether the mainspring has completely unwound.
func (pr *PowerReserve) Depleted() bool {
	return pr.hours <= 0.0001
}

// Wind adds energy to the mainspring barrel up to MaxPowerReserve.
func (pr *PowerReserve) Wind(amount float64) {
	pr.hours = min(MaxPowerReserve, pr.hours+amount)
}

// Drain consumes power based on elapsed seconds.
func (pr *PowerReserve) Drain(elapsedSec float64) {
	if elapsedSec <= 0 || pr.Depleted() {
		return
	}
	consumed := (elapsedSec / 3600.0) * pr.drainSpeed
	pr.hours = max(0.0, pr.hours-consumed)
}

// Angle computes the subdial hand rotation angle (-60° empty to +60° full).
func (pr *PowerReserve) Angle() float64 {
	ratio := pr.hours / MaxPowerReserve
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	return -60.0 + ratio*120.0
}
