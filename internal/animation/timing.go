package animation

import "time"

// Direction controls the order in which iterations play.
type Direction uint8

const (
	Normal Direction = iota
	Reverse
	Alternate
	AlternateReverse
)

// FillMode controls paint before and after the active interval.
type FillMode uint8

const (
	FillNone FillMode = iota
	FillForwards
	FillBackwards
	FillBoth
)

// PlayState controls whether Step applies the keyframes.
type PlayState uint8

const (
	Running PlayState = iota
	Paused
)

// Composition controls how an effect combines with the sampled paint.
type Composition uint8

const (
	Replace Composition = iota
	Add
	Accumulate
)

// Range maps a named timeline range to normalized timeline progress.
type Range struct{ Start, End float64 }

// Timing describes the animation clock, timeline, and iteration behavior.
type Timing struct {
	Duration, Delay, EndDelay time.Duration
	DurationAuto              bool
	Iterations                float64
	Infinite                  bool
	Direction                 Direction
	Fill                      FillMode
	Play                      PlayState
	TimingFunction            string
	Easing                    Easing
	Composition               Composition
	Timeline                  string
	RangeStart, RangeEnd      string
	RangeCenter               string
}

// FrameInput carries host clock and timeline progress for one frame.
type FrameInput struct {
	Now                            time.Time
	ScrollProgress, ViewProgress   float64
	SourceProgress, TargetProgress float64
	Timelines                      map[string]float64
	Ranges                         map[string]Range
	RangeEdges                     map[string]float64
}

// Effect is one animation item sharing its target paint with sibling effects.
type Effect struct {
	Frames  []Keyframe
	Timing  Timing
	Trigger string
}
