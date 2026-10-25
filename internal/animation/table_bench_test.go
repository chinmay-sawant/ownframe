package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const stepTargets = 4

// liveStepPlayer builds a player whose effects stay live for every frame of a
// benchmark or allocation test, with stops on every sampled field.
func liveStepPlayer(tb testing.TB, effectCount int) (*Player, time.Time) {
	tb.Helper()
	targets := make([]Target, stepTargets)
	for i := range targets {
		op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
			X: float64(i * 10), Y: float64(i * 5), W: 40, H: 20, Alpha: 1}
		targets[i] = Target{Op: op, From: cssmotion.Read(op)}
	}
	effects := make([]Effect, effectCount)
	for i := range effects {
		effects[i] = Effect{
			Frames: liveFrames(i),
			Timing: Timing{Duration: time.Hour, Infinite: true},
		}
	}
	player, err := NewComposite(targets, effects)
	if err != nil {
		tb.Fatal(err)
	}
	start := time.Unix(0, 0)
	player.Start(start)
	player.Step(start)
	return player, start
}

// liveFrames returns keyframes with every sampled field present, so the
// measurement covers all six stops of the table.
func liveFrames(seed int) []Keyframe {
	base := float64(seed * stepTargets)
	return []Keyframe{
		{Offset: 0, X: base, Y: base, Opacity: 1, OpacitySet: true,
			R: 0.1, G: 0.2, B: 0.3, ColorSet: true, Move: true},
		{Offset: 0.33, X: base + 3, Y: base + 2, Opacity: 0.7, OpacitySet: true,
			R: 0.4, G: 0.5, B: 0.6, ColorSet: true, Move: true},
		{Offset: 0.66, X: base + 6, Y: base + 4, Opacity: 0.4, OpacitySet: true,
			R: 0.7, G: 0.8, B: 0.9, ColorSet: true, Move: true},
		{Offset: 1, X: base + 9, Y: base + 6, Opacity: 1, OpacitySet: true,
			R: 0.1, G: 0.2, B: 0.3, ColorSet: true, Move: true},
	}
}

func benchmarkStep(b *testing.B, effectCount int) {
	b.Helper()
	player, start := liveStepPlayer(b, effectCount)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.Step(start.Add(time.Duration(i) * 16 * time.Millisecond))
	}
}

func BenchmarkSampleOneEffect(b *testing.B) { benchmarkStep(b, 1) }

func BenchmarkSampleTenEffects(b *testing.B) { benchmarkStep(b, 10) }
