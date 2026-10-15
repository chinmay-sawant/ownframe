package eyes

import (
	"math"
	"time"
)

// stepSpiral eases the swirl in and out while keeping rotation frame-rate safe.
func (a *App) stepSpiral(dt float64, now time.Time) {
	s := &a.st
	target := 0.0
	if s.mood == 3 {
		target = 1
		s.sacT = now.Add(500 * time.Millisecond)
		if !now.Before(s.tiltUntil) {
			s.tLookX, s.tLookY = 0, 0
		}
	}
	s.spiralMix += (target - s.spiralMix) * (1 - math.Exp(-dt*12))
	if math.Abs(target-s.spiralMix) < 0.001 {
		s.spiralMix = target
	}
	if s.spiralMix > 0.001 {
		s.spiral = math.Mod(s.spiral+dt*600, 360)
	}
}

func (a *App) selectMood(mood int, now time.Time) {
	s := &a.st
	s.mood, s.moodT, s.spiralUntil = mood, now, time.Time{}
	if mood == 3 {
		s.spiralUntil = now.Add(2 * time.Second)
	}
}

func (a *App) expireSpiral(now time.Time) {
	s := &a.st
	if s.mood != 3 || s.spiralUntil.IsZero() || now.Before(s.spiralUntil) {
		return
	}
	a.selectMood(0, now)
}
