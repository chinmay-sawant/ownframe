package pet

import "time"

const (
	MinStat = 0.0
	MaxStat = 100.0

	MoodHappy    = "happy"
	MoodHungry   = "hungry"
	MoodTired    = "tired"
	MoodSleeping = "sleeping"
	MoodSad      = "sad"
	MoodPlayful  = "playful"
	MoodPurring  = "purring"
)

// State tracks pet vitals, mood, and active timers.
type State struct {
	Energy       float64
	Hunger       float64
	Happiness    float64
	Mood         string
	Sleeping     bool
	PettingUntil time.Time
	PlayUntil    time.Time
	Speech       string
	SpeechUntil  time.Time
}

func newState(now func() time.Time) State {
	t := now()
	return State{
		Energy:      80.0,
		Hunger:      80.0,
		Happiness:   80.0,
		Mood:        MoodHappy,
		Speech:      "System online! Ready to play! ⚡",
		SpeechUntil: t.Add(4 * time.Second),
	}
}

// Decay steps vitals by dt seconds according to sleep and activity.
func (s *State) Decay(dt float64, now time.Time) {
	if dt <= 0 {
		return
	}
	if s.Sleeping {
		s.Energy = clamp(s.Energy+6.0*dt, MinStat, MaxStat)
		s.Hunger = clamp(s.Hunger-1.0*dt, MinStat, MaxStat)
		s.Happiness = clamp(s.Happiness-0.5*dt, MinStat, MaxStat)
	} else {
		s.Energy = clamp(s.Energy-1.0*dt, MinStat, MaxStat)
		s.Hunger = clamp(s.Hunger-1.5*dt, MinStat, MaxStat)
		s.Happiness = clamp(s.Happiness-1.0*dt, MinStat, MaxStat)
	}
	s.updateMood(now)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
