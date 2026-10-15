package eyes

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onTap flips the devtools switch on its button, or cycles the mood and
// triggers a blink anywhere else.
func (a *App) onTap(_ context.Context, b ownframe.Box) error {
	if b.ID == "dev-toggle" {
		a.st.dev = !a.st.dev
		a.page.SetDevTools(a.st.dev)
		return nil
	}
	a.selectMood((a.st.mood+1)%len(moods), a.now())
	a.st.blinkOn = true
	a.st.blinkP = 0
	return nil
}

// onSwipe flicks the look target in the swipe direction.
func (a *App) onSwipe(_ context.Context, dx, dy float64) error {
	a.st.tLookX = clamp(dx*0.2, -48, 48)
	a.st.tLookY = clamp(dy*0.2, -30, 30)
	a.st.sacT = a.now().Add(2500000000)
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
