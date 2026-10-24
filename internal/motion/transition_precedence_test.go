package motion

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// fadeAnimation is one 100ms linear opacity animation from 0 to 1 with both
// fills, so it owns the field before the first sample and after it finishes.
func fadeAnimation() (layout.Animation, layout.KeyframesRule) {
	return layout.Animation{
			Name: "fade", DurationMS: transitionFrameMS, IterationCount: 1,
			TimingFunction: "linear", FillMode: "both", PlayState: "running",
		}, layout.KeyframesRule{
			Name: "fade",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "opacity", Value: "0"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "opacity", Value: "1"}}},
			},
		}
}

// TestAnimationPrecedenceSkipsTheTransitionCapture checks the cascade rule: an
// animation on opacity owns the field, so the transition capture is counted
// and never becomes a renderer. The animation's value is what every sampled
// frame shows, including after it finishes.
func TestAnimationPrecedenceSkipsTheTransitionCapture(t *testing.T) {
	record := linearTransition("opacity")
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, .5, 0, record)
	animation, keyframes := fadeAnimation()
	next.Boxes[0].Animations = []layout.Animation{animation}
	next.Keyframes = []layout.KeyframesRule{keyframes}

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	if err := r.Bind(next); err != nil {
		t.Fatalf("bind animation: %v", err)
	}
	captures := r.CaptureTransitions(old, next)
	if len(captures) != 1 {
		t.Fatalf("captures = %d, want the opacity change", len(captures))
	}
	r.BindTransitions(next, captures, start)
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("started = %d, want the animation to win", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want 0", got)
	}
	if got := r.Stats().TransitionsSkipped; got != 1 {
		t.Fatalf("skipped = %d, want the precedence skip counted", got)
	}

	half := start.Add(transitionFrameMS / 2 * time.Millisecond)
	r.Step(next, half)
	if got := next.Ops[0].Opacity(); !near(got, .5) {
		t.Fatalf("opacity at the midpoint = %v, want the animation's .5", got)
	}
	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if got := next.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity at the end = %v, want the animation's 1", got)
	}

	// A later commit cannot revive the skipped transition.
	r.BindTransitions(next, nil, start.Add(transitionFrameMS*time.Millisecond))
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("a later commit started %d transitions, want 0", got)
	}
	r.Step(next, start.Add(transitionFrameMS*2*time.Millisecond))
	if got := next.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity after the animation = %v, want the fill's 1", got)
	}
}

// TestAnimationPrecedenceCancelsARunningTransition checks the other direction:
// when an animation appears while a transition is in flight, the transition is
// dropped and does not resume after the animation ends.
func TestAnimationPrecedenceCancelsARunningTransition(t *testing.T) {
	record := layout.Transition{Property: "opacity", DurationMS: 1000,
		TimingFunction: "linear"}
	old := transitionChip(1, 1, 0, record)
	next := transitionChip(1, .5, 0, record)

	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	captures := r.CaptureTransitions(old, next)
	r.BindTransitions(next, captures, start)
	interrupt := start.Add(250 * time.Millisecond)
	r.Step(next, interrupt)
	if got := next.Ops[0].Opacity(); !near(got, .875) {
		t.Fatalf("transition sample = %v, want .875", got)
	}

	animated := transitionChip(1, .5, 0, record)
	animation, keyframes := fadeAnimation()
	animated.Boxes[0].Animations = []layout.Animation{animation}
	animated.Keyframes = []layout.KeyframesRule{keyframes}
	r.now = func() time.Time { return interrupt }
	if err := r.Bind(animated); err != nil {
		t.Fatalf("bind animation: %v", err)
	}
	r.BindTransitions(animated, nil, interrupt)
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("cancelled = %d, want the transition dropped", got)
	}
	if got := r.Stats().TransitionsBound; got != 0 {
		t.Fatalf("bound = %d, want 0", got)
	}

	r.Step(animated, interrupt.Add(transitionFrameMS/2*time.Millisecond))
	if got := animated.Ops[0].Opacity(); !near(got, .5) {
		t.Fatalf("opacity at the animation midpoint = %v, want .5", got)
	}
	r.Step(animated, interrupt.Add(transitionFrameMS*time.Millisecond))
	if got := animated.Ops[0].Opacity(); !near(got, 1) {
		t.Fatalf("opacity after the animation = %v, want the fill's 1", got)
	}
	if got := r.Stats().TransitionsStarted; got != 0 {
		t.Fatalf("the cancelled transition revived: started = %d", got)
	}
}
