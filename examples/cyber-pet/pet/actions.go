package pet

import (
	"time"
)

// Feed satisfies hunger and adds a little happiness.
func (a *App) Feed() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.state.Sleeping = false
	a.state.Hunger = clamp(a.state.Hunger+25.0, MinStat, MaxStat)
	a.state.Happiness = clamp(a.state.Happiness+5.0, MinStat, MaxStat)
	a.state.Speech = "Yum! Delicious cyber snacks! 🍗"
	a.state.SpeechUntil = now.Add(3 * time.Second)
	a.state.updateMood(now)
	a.refreshView(now)
}

// Play costs energy and hunger to boost happiness.
func (a *App) Play() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.state.Sleeping = false
	if a.state.Energy < 15.0 {
		a.state.Speech = "Too tired to play... *yawn* 💤"
		a.state.SpeechUntil = now.Add(3 * time.Second)
		a.state.updateMood(now)
		a.refreshView(now)
		return
	}

	a.state.Energy = clamp(a.state.Energy-15.0, MinStat, MaxStat)
	a.state.Happiness = clamp(a.state.Happiness+20.0, MinStat, MaxStat)
	a.state.Hunger = clamp(a.state.Hunger-10.0, MinStat, MaxStat)
	a.state.PlayUntil = now.Add(2500 * time.Millisecond)
	a.state.Speech = "Wheee! High voltage fun! ⚡"
	a.state.SpeechUntil = now.Add(3 * time.Second)
	a.state.updateMood(now)
	a.refreshView(now)
}

// Pet triggers purring, hearts, and boosts happiness.
func (a *App) Pet() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.state.Happiness = clamp(a.state.Happiness+15.0, MinStat, MaxStat)
	a.state.PettingUntil = now.Add(3 * time.Second)
	a.state.Speech = "Purrrrrr... <3"
	a.state.SpeechUntil = now.Add(3 * time.Second)
	a.triggerHeart(now)
	a.state.updateMood(now)
	a.refreshView(now)
}
