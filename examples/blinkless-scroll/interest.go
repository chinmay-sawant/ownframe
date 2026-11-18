package main

import "time"

// Clock is the time source for interest delays. Tests pass a fake.
type Clock interface{ Now() time.Time }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Delays the host uses when the page says "normal". The CSS spec leaves the
// normal value to the user agent; these are this host's choice.
const (
	NormalStartDelay = 500 * time.Millisecond
	NormalEndDelay   = 250 * time.Millisecond
)

// InterestDelay mirrors the engine's interest-delay values.
type InterestDelay struct {
	StartNormal, EndNormal bool
	StartMs, EndMs         float64
}

func (d InterestDelay) start() time.Duration {
	if d.StartNormal {
		return NormalStartDelay
	}

	return time.Duration(d.StartMs * float64(time.Millisecond))
}

func (d InterestDelay) end() time.Duration {
	if d.EndNormal {
		return NormalEndDelay
	}

	return time.Duration(d.EndMs * float64(time.Millisecond))
}

// Interest tracks one element's interest state. Hover starts a start timer;
// the state turns on when the timer runs out while the pointer is still on
// the element. Leaving starts the end timer; the state turns off when it runs
// out. Coming back before a timer runs out cancels it.
type Interest struct {
	delay   InterestDelay
	clock   Clock
	active  bool
	hovered bool
	since   time.Time // when the pending hover or leave began
	pending bool
}

// NewInterest returns an off tracker. A nil clock uses the wall clock.
func NewInterest(d InterestDelay, c Clock) *Interest {
	if c == nil {
		c = realClock{}
	}

	return &Interest{delay: d, clock: c}
}

// Hover reports whether the pointer is on the element now.
func (i *Interest) Hover(on bool) {
	if on == i.hovered {
		return
	}

	i.hovered = on
	// A hover while off, or a leave while on, starts a timer. The opposite
	// move cancels the timer the other one started.
	i.pending = on != i.active
	if i.pending {
		i.since = i.clock.Now()
	}
}

// Tick advances the timers and returns the current state. Call it once a
// frame.
func (i *Interest) Tick() bool {
	if !i.pending {
		return i.active
	}

	wait := i.delay.end()
	if i.hovered {
		wait = i.delay.start()
	}

	if i.clock.Now().Sub(i.since) >= wait {
		i.active = i.hovered
		i.pending = false
	}

	return i.active
}

// Active is the last state Tick returned.
func (i *Interest) Active() bool { return i.active }

// NextDeadline returns when the pending timer fires, or false when none runs.
func (i *Interest) NextDeadline() (time.Time, bool) {
	if !i.pending {
		return time.Time{}, false
	}

	wait := i.delay.end()
	if i.hovered {
		wait = i.delay.start()
	}

	return i.since.Add(wait), true
}
