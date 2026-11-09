package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// typingTextID names the text span inside the typing bar that patchTyping
// writes. Sibling-owned markup (messages.go) exposes ID("dc-typing") on the
// bar and ID("dc-typing-text") on the span; until that markup lands, the
// patch misses and patchTyping falls back to a full refresh.
const typingTextID = "dc-typing-text"

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

// typingPinnedNames counts the names startTyping places first (Theo and
// Priya) regardless of roster order.
const typingPinnedNames = 2

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
	names := make([]string, 0, len(rest)+typingPinnedNames)
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

// typingText flattens the indicator exactly the way mainCol renders it:
// the names joined with ", " plus the " is typing"/" are typing" verb.
// Empty when nobody is typing; the caller patches the span to "" so the
// bar keeps its 22px box with no layout shift. It builds the string with
// strings.Join plus one concatenation: no fmt, no new maps or slices, and
// no appends on the ticking path.
func (a *app) typingText() string {
	if len(a.typing) == 0 {
		return ""
	}
	verb := " is typing"
	if len(a.typing) > 1 {
		verb = " are typing"
	}
	return strings.Join(a.typing, ", ") + verb
}

// patchTyping writes the new indicator text into the retained typing bar
// instead of rebuilding the whole view. A refused patch drops the cached
// tree, so any failure falls back to a full refresh and the bar never
// goes stale.
func (a *app) patchTyping(ctx context.Context) error {
	text := a.typingText()
	err := a.page.PatchElement(ctx, typingTextID, ownframe.ElementEdit{Text: &text})
	if err == nil {
		return nil
	}
	if errors.Is(err, ownframe.ErrPatchRefused) {
		return a.refresh(ctx)
	}
	return a.refresh(ctx)
}

// tickTyping re-arms the next wakeup and patches only when the state
// machine changed something. It never touches drafts or forms. An expired
// toast clears here, since this tick already wakes every second.
func (a *app) tickTyping(ctx context.Context) error {
	if a.page == nil {
		return nil
	}
	now := time.Now()
	a.page.WantFrameAt(now.Add(typingTickEvery))
	if a.toastOpen && !a.toastUntil.IsZero() && !now.Before(a.toastUntil) {
		a.toastOpen = false
		a.toastTitle = ""
		a.toastBody = ""
		a.toastUntil = time.Time{}
		return a.refresh(ctx)
	}
	if a.rotateTyping(now) {
		return a.patchTyping(ctx)
	}
	return nil
}
