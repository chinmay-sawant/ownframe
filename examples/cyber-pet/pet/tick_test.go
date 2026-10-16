package pet

import (
	"testing"
	"time"
)

func TestMoodTransitions(t *testing.T) {
	a, c := newT(t)
	now := c.now()

	a.state.Hunger = 20.0
	a.state.updateMood(now)
	if a.state.Mood != MoodHungry {
		t.Fatalf("expected hungry, got %s", a.state.Mood)
	}

	a.state.Hunger = 80.0
	a.state.Energy = 15.0
	a.state.updateMood(now)
	if a.state.Mood != MoodTired {
		t.Fatalf("expected tired, got %s", a.state.Mood)
	}

	a.state.Energy = 80.0
	a.state.Happiness = 20.0
	a.state.updateMood(now)
	if a.state.Mood != MoodSad {
		t.Fatalf("expected sad, got %s", a.state.Mood)
	}

	a.state.Happiness = 90.0
	a.state.updateMood(now)
	if a.state.Mood != MoodPlayful {
		t.Fatalf("expected playful, got %s", a.state.Mood)
	}
}

func TestTickDecayAndSleepRecovery(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)

	s0 := a.State()
	tk(t, a, c, 5*time.Second)
	s1 := a.State()

	if s1.Energy >= s0.Energy || s1.Hunger >= s0.Hunger {
		t.Fatalf("expected vitals decay: before %+v, after %+v", s0, s1)
	}

	a.Sleep()
	e0 := a.State().Energy
	tk(t, a, c, 5*time.Second)
	e1 := a.State().Energy

	if e1 <= e0 {
		t.Fatalf("expected energy recovery in sleep: before %f, after %f", e0, e1)
	}
}

func TestTickBindings(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)

	if a.pm.eyeL == nil || a.pm.eyeR == nil {
		t.Fatal("eyes not bound by tick")
	}
	if a.pm.mouth == nil {
		t.Fatal("mouth not bound by tick")
	}
}
