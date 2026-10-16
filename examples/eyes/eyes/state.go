package eyes

import "time"

const spiralMood = 3

func moodName(i int) string {
	return poses[i%len(poses)].name
}

// state is the eye simulation the tick advances. Eye styles ease toward each
// pose while both modes keep blinking and breathing.
type state struct {
	mood       int
	moodUntil  time.Time
	poseAt     time.Time
	hudUntil   time.Time
	automatic  bool
	lastTapDir int
	pressDir   int
	lastTapAt  time.Time
	lookX      float64
	lookY      float64
	tLookX     float64
	tLookY     float64
	sacT       time.Time
	blinkT     time.Time
	blinkOn    bool
	blinkP     float64
	moodL      float64
	moodR      float64
	styleL     eyeStyle
	styleR     eyeStyle
	dispL      float64
	dispR      float64
	// bm is the blink multiplier: lids close fully on blinks and partly
	// on moods, so squints read as lids instead of a shrinking fish.
	bm    float64
	birth time.Time
	drag  bool
	dev   bool
	// tiltX/tiltY hold the latest screen-aligned phone tilt.
	tiltX     float64
	tiltY     float64
	tiltUntil time.Time
	spiral    float64
	spiralMix float64
}

func newState(now func() time.Time) state {
	t := now()
	return state{moodUntil: t.Add(moodDuration(0)), poseAt: t,
		hudUntil: t.Add(hudIdle), automatic: true,
		sacT: t, blinkT: t.Add(2 * time.Second),
		moodL: 1, moodR: 1, styleL: eyeStyle{open: 1, height: 1},
		styleR: eyeStyle{open: 1, height: 1}, dispL: 1, dispR: 1, birth: t}
}
