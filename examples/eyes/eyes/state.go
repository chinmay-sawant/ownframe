package eyes

import "time"

// moods cycle automatically; tap advances to the next one.
var moods = []string{"idle", "happy", "sleepy", "surprised", "curious"}

func moodName(i int) string {
	return moods[i%len(moods)]
}

// state is the eye simulation the tick advances. moodL/R ease toward the
// mood target slowly so squints glide; dispL/R add the fast blink and the
// breathing wave on top each frame.
type state struct {
	mood    int
	moodT   time.Time
	lookX   float64
	lookY   float64
	tLookX  float64
	tLookY  float64
	sacT    time.Time
	blinkT  time.Time
	blinkOn bool
	blinkP  float64
	moodL   float64
	moodR   float64
	dispL   float64
	dispR   float64
	// bm is the blink multiplier: lids close fully on blinks and partly
	// on moods, so squints read as lids instead of a shrinking fish.
	bm    float64
	birth time.Time
	drag  bool
	dev   bool
}

func newState(now func() time.Time) state {
	t := now()
	return state{moodT: t, sacT: t, blinkT: t.Add(2 * time.Second),
		moodL: 1, moodR: 1, dispL: 1, dispR: 1, birth: t}
}

// targetOpen returns the per-eye open ratio for a mood.
func targetOpen(mood int) (float64, float64) {
	switch moodName(mood) {
	case "happy":
		return 0.32, 0.32
	case "sleepy":
		return 0.42, 0.42
	case "surprised":
		return 1.18, 1.18
	case "curious":
		return 1.0, 0.62
	default:
		return 1.0, 1.0
	}
}
