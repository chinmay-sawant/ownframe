package main

import (
	"context"
	"time"
)

// Typing simulation cadence: a teammate types for typingVisibleFor, the
// indicator clears for typingHiddenFor, then someone else types. The first
// indicator waits typingFirstDelay after startup. The tick re-arms itself
// every typingTickEvery via WantFrameAt; rotateTyping gates on typingUntil
// so early wakeups (input) are no-ops.
const (
	typingFirstDelay = 2 * time.Second
	typingVisibleFor = 5 * time.Second
	typingHiddenFor  = 3 * time.Second
	typingTickEvery  = time.Second
	typingPairEvery  = 3
)

// startTyping builds the rotation from the current roster (non-bot members
// with Theo first, Priya second, then the rest in roster order) and installs
// the once-per-second tick that drives the indicator.
func (a *app) startTyping() {
	var theo, priya string
	var rest []string
	for _, m := range a.currentMembers() {
		if m.Bot {
			continue
		}
		switch m.Name {
		case "Theo":
			theo = m.Name
		case "Priya":
			priya = m.Name
		default:
			rest = append(rest, m.Name)
		}
	}
	names := make([]string, 0, len(rest)+2)
	if theo != "" {
		names = append(names, theo)
	}
	if priya != "" {
		names = append(names, priya)
	}
	names = append(names, rest...)
	a.typingNames = names
	a.typing = nil
	a.typingUntil = time.Now().Add(typingFirstDelay)
	if a.page == nil {
		return
	}
	a.page.SetTick(func(ctx context.Context) error { return a.tickTyping(ctx) })
	a.page.WantFrameAt(time.Now().Add(typingTickEvery))
}

// rotateTyping advances the typing indicator state machine. It touches only
// typing, typingIdx, and typingUntil, so it stays unit-testable without a
// window. It reports whether the screen changed.
func (a *app) rotateTyping(now time.Time) bool {
	if now.Before(a.typingUntil) {
		return false
	}
	if len(a.typing) > 0 {
		a.typing = nil
		a.typingUntil = now.Add(typingHiddenFor)
		return true
	}
	if len(a.typingNames) == 0 {
		return false
	}
	a.typingIdx++
	idx := (a.typingIdx - 1) % len(a.typingNames)
	if a.typingIdx%typingPairEvery == 0 && len(a.typingNames) > 1 {
		next := (idx + 1) % len(a.typingNames)
		a.typing = []string{a.typingNames[idx], a.typingNames[next]}
	} else {
		a.typing = []string{a.typingNames[idx]}
	}
	a.typingUntil = now.Add(typingVisibleFor)
	return true
}

// tickTyping re-arms the next wakeup and refreshes only when the state
// machine changed something. It never touches drafts or forms.
func (a *app) tickTyping(ctx context.Context) error {
	if a.page == nil {
		return nil
	}
	a.page.WantFrameAt(time.Now().Add(typingTickEvery))
	if a.rotateTyping(time.Now()) {
		return a.refresh(ctx)
	}
	return nil
}
