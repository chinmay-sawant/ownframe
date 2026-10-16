package watch

import (
	"math"
	"time"
)

// timeAngles computes rotation angles in degrees for standard clock hands.
func timeAngles(t time.Time) (hour, minute, second float64) {
	secVal := float64(t.Second()) + float64(t.Nanosecond())/1e9
	// High-beat escapement (8 ticks per second for 28,800 vph).
	beatSec := math.Floor(secVal*8.0) / 8.0
	second = math.Mod(beatSec*6.0, 360.0)

	minVal := float64(t.Minute()) + secVal/60.0
	minute = math.Mod(minVal*6.0, 360.0)

	hrVal := float64(t.Hour()%12) + minVal/60.0
	hour = math.Mod(hrVal*30.0, 360.0)
	return hour, minute, second
}

// tourbillonAngles computes tourbillon cage rotation and balance wheel swing.
func tourbillonAngles(t time.Time, stopped bool) (cage, balance float64) {
	if stopped {
		return 0, 0
	}
	secVal := float64(t.Second()) + float64(t.Nanosecond())/1e9
	// 1 revolution per minute (60 seconds per turn).
	cage = math.Mod(secVal*6.0, 360.0)
	// 4 Hz balance wheel oscillation (28,800 vph, +/- 35 degrees amplitude).
	balance = 35.0 * math.Sin(secVal*4.0*2.0*math.Pi)
	return cage, balance
}

// chronoAngles computes chronograph hands rotation angles.
func chronoAngles(d time.Duration) (sec, min, hr float64) {
	s := d.Seconds()
	if s < 0 {
		s = 0
	}
	sec = math.Mod(s, 60.0) * 6.0
	min = math.Mod(s/60.0, 30.0) * 12.0
	hr = math.Mod(s/3600.0, 12.0) * 30.0
	return sec, min, hr
}

// gearAngles computes rotation for the visible skeleton gear train.
func gearAngles(t time.Time, stopped bool) (center, third, fourth, barrel float64) {
	if stopped {
		return 0, 0, 0, 0
	}
	secVal := float64(t.Second()) + float64(t.Nanosecond())/1e9
	center = math.Mod(secVal/10.0, 360.0)
	third = -math.Mod(secVal*1.5, 360.0)
	fourth = math.Mod(secVal*6.0, 360.0)
	barrel = math.Mod(secVal*0.05, 360.0)
	return center, third, fourth, barrel
}
