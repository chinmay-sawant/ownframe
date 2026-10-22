package eyes

import (
	"math"
	"time"
)

// step advances look targets, blinks, mood cycling, and smoothing. Rates
// use 1-exp(-dt*r) so 60 Hz and 90 Hz phones feel the same: mood glides at
// ~4.5/s, eye tracking settles at ~18/s, and the blink sine stays immediate.
func (a *App) step(dt float64, now time.Time) {
	s := &a.st
	if dt <= 0 {
		dt = 1.0 / baseFPS
	}
	if s.automatic && !s.drag && !now.Before(s.moodUntil) {
		a.selectMood((s.mood+1)%len(poses), now)
	}
	a.stepSpiral(dt, now)
	if !s.drag && now.Before(s.tiltUntil) {
		s.tLookX = clamp(s.tiltX*tiltGain, -lookXRange, lookXRange)
		s.tLookY = clamp(-s.tiltY*tiltGain, -lookYRange, lookYRange)
		s.sacT = now.Add(tiltHold)
	} else if !s.drag && s.mood != spiralMood && now.After(s.sacT) {
		s.tLookX = (a.rng.Float64()*unitSpan - 1) * saccadeX
		s.tLookY = (a.rng.Float64()*unitSpan - 1) * saccadeY
		switch moodName(s.mood) {
		case "squint":
			s.tLookY = squintShift + a.rng.Float64()*squintShift
		case "angry":
			s.tLookY = -squintShift - a.rng.Float64()*squintShift
		}
		s.sacT = now.Add(time.Duration(saccadeDelayMin+
			a.rng.IntN(saccadeDelayVar)) * time.Millisecond)
	}
	if !s.blinkOn && now.After(s.blinkT) {
		s.blinkOn = true
		s.blinkP = 0
	}
	if s.blinkOn {
		s.blinkP += dt / blinkPeriod
		if s.blinkP >= 1 {
			s.blinkOn = false
			s.blinkP = 0
			s.blinkT = now.Add(time.Duration(blinkDelayMin+
				a.rng.IntN(blinkDelayVar)) * time.Millisecond)
		}
	}
	a.stepPose(dt)
	bm := 1.0
	if s.blinkOn {
		bm = 1 - lidClose*math.Sin(math.Pi*min(max(s.blinkP, 0), 1))
	}
	s.bm = bm
	br := 1 + breathAmp*math.Sin(twoPi*breathHz*now.Sub(s.birth).Seconds())
	s.dispL = s.styleL.open * bm * br
	s.dispR = s.styleR.open * bm * br
	kL := 1 - math.Exp(-dt*lookRate)
	s.lookX += (s.tLookX - s.lookX) * kL
	s.lookY += (s.tLookY - s.lookY) * kL
}
