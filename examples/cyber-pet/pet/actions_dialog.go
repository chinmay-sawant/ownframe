package pet

import "time"

var cyberDialogs = []string{
	"Beep boop! All circuits operational! ⚡",
	"Hello human! You are my favorite user! <3",
	"01001000 01101001! (That means Hi!) ✨",
	"System check: Cuteness levels at 100%! 🐾",
	"Need more RAM... Just kidding, I need hugs! 💕",
	"Optimizing happiness algorithms... Done! 🤖",
	"My neural net dreams of electric yarn! 🧶",
}

// Sleep toggles sleeping state for resting and energy recovery.
func (a *App) Sleep() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.state.Sleeping = !a.state.Sleeping
	if a.state.Sleeping {
		a.state.Speech = "Powering down... Zzz... 💤"
	} else {
		a.state.Speech = "Online and recharged! ⚡"
	}
	a.state.SpeechUntil = now.Add(3 * time.Second)
	a.state.updateMood(now)
	a.refreshView(now)
}

// Talk makes the pet speak a cyber dialogue.
func (a *App) Talk() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	idx := a.rng.IntN(len(cyberDialogs))
	a.state.Happiness = clamp(a.state.Happiness+5.0, MinStat, MaxStat)
	a.state.Speech = cyberDialogs[idx]
	a.state.SpeechUntil = now.Add(4 * time.Second)
	a.state.updateMood(now)
	a.refreshView(now)
}
