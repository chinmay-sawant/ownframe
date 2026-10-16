package eyes

import (
	"math"
	"time"
)

// stepSpiral eases the swirl in and out while keeping rotation frame-rate safe.
func (a *App) stepSpiral(dt float64, now time.Time) {
	s := &a.st
	target := 0.0
	if s.mood == spiralMood {
		target = 1
		if s.automatic {
			s.sacT = now.Add(500 * time.Millisecond)
			if !now.Before(s.tiltUntil) {
				s.tLookX, s.tLookY = 0, 0
			}
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
	if mood < 0 || mood >= len(poses) {
		mood = 0
	}
	s.mood = mood
	s.poseAt = now
	s.moodUntil = time.Time{}
	if s.automatic {
		s.moodUntil = now.Add(moodDuration(mood))
	}
	if mood == spiralMood {
		s.spiral = 0
	}
}
