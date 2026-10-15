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
		dt = 1.0 / 60
	}
	a.expireSpiral(now)
	a.stepSpiral(dt, now)
	if !s.drag && s.mood != 3 && now.Sub(s.moodT) > 8*time.Second {
		s.mood = (s.mood + 1) % 3
		s.moodT = now
	}
	if !s.drag && now.Before(s.tiltUntil) {
		s.tLookX = clamp(s.tiltX*96, -48, 48)
		s.tLookY = clamp(-s.tiltY*96, -30, 30)
		s.sacT = now.Add(1500 * time.Millisecond)
	} else if !s.drag && s.mood != 3 && now.After(s.sacT) {
		s.tLookX = (a.rng.Float64()*2 - 1) * 44
		s.tLookY = (a.rng.Float64()*2 - 1) * 26
		switch moodName(s.mood) {
		case "squint":
			s.tLookY = 6 + a.rng.Float64()*6
		case "angry":
			s.tLookY = -6 - a.rng.Float64()*6
		}
		s.sacT = now.Add(time.Duration(1400+
			a.rng.IntN(1600)) * time.Millisecond)
	}
	if !s.blinkOn && now.After(s.blinkT) {
		s.blinkOn = true
		s.blinkP = 0
	}
	if s.blinkOn {
		s.blinkP += dt / 0.22
		if s.blinkP >= 1 {
			s.blinkOn = false
			s.blinkP = 0
			s.blinkT = now.Add(time.Duration(2200+
				a.rng.IntN(2600)) * time.Millisecond)
		}
	}
	tL, tR := targetOpen(s.mood)
	kM := 1 - math.Exp(-dt*4.5)
	s.moodL += (tL - s.moodL) * kM
	s.moodR += (tR - s.moodR) * kM
	bm := 1.0
	if s.blinkOn {
		bm = 1 - 0.93*math.Sin(math.Pi*min(max(s.blinkP, 0), 1))
	}
	s.bm = bm
	br := 1 + 0.014*math.Sin(2*math.Pi*0.35*now.Sub(s.birth).Seconds())
	s.dispL = s.moodL * bm * br
	s.dispR = s.moodR * bm * br
	kL := 1 - math.Exp(-dt*18)
	s.lookX += (s.tLookX - s.lookX) * kL
	s.lookY += (s.tLookY - s.lookY) * kL
}
