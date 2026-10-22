package eyes

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// onTap flips the devtools switch on its button, or cycles the mood and
// triggers a blink anywhere else.
func (a *App) onTap(_ context.Context, b ownframe.Box) error {
	now := a.now()
	direction := a.st.pressDir
	a.st.pressDir = 0
	a.wakeHUD(now)
	if b.ID == "dev-toggle" {
		a.st.dev = !a.st.dev
		a.page.SetDevTools(a.st.dev)
		return nil
	}
	if b.ID == "mode-toggle" {
		a.toggleMode(a.now())
		return nil
	}
	if direction != 0 {
		a.onSideTap(direction, now)
		return nil
	}
	a.selectMood((a.st.mood+1)%len(poses), now)
	if a.st.automatic {
		a.st.blinkOn = true
		a.st.blinkP = 0
	}
	return nil
}

func (a *App) toggleMode(now time.Time) {
	s := &a.st
	s.automatic = !s.automatic
	if s.automatic {
		s.moodUntil = now.Add(moodDuration(s.mood))
		s.sacT = now.Add(time.Second)
		return
	}
	s.moodUntil = time.Time{}
}

// onSwipe flicks the look target in the swipe direction.
func (a *App) onSwipe(_ context.Context, dx, dy float64) error {
	a.wakeHUD(a.now())
	a.st.tLookX = clamp(dx*swipeGain, -lookXRange, lookXRange)
	a.st.tLookY = clamp(dy*swipeGain, -lookYRange, lookYRange)
	a.st.sacT = a.now().Add(swipeCooldown)
	return nil
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
