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
			s.sacT = now.Add(spiralSettle)
			if !now.Before(s.tiltUntil) {
				s.tLookX, s.tLookY = 0, 0
			}
		}
	}
	s.spiralMix += (target - s.spiralMix) * (1 - math.Exp(-dt*spiralRate))
	if math.Abs(target-s.spiralMix) < spiralMinMix {
		s.spiralMix = target
	}
	if s.spiralMix > spiralMinMix {
		s.spiral = math.Mod(s.spiral+dt*spiralSpin, spiralTurn)
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
