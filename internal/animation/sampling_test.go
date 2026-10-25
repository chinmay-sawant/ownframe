package animation

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestStepSamplesTransformOpacityAndColor(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 10, Y: 4,
		W: 20, H: 10, R: 0, B: 1, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 1, X: 180, Move: true, Opacity: 1, OpacitySet: true, R: 0, B: 1, ColorSet: true},
		{Offset: .5, X: 90, Move: true, Opacity: .72, OpacitySet: true, R: 226.0 / 255, G: 122.0 / 255, ColorSet: true},
		{Offset: 0, Move: true, Opacity: 1, OpacitySet: true, R: 0, B: 1, ColorSet: true},
	}, Timing{Duration: 2 * time.Second, Iterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(10, 0)
	p.Start(start)
	if !p.Step(start.Add(time.Second)) {
		t.Fatal("midpoint should remain active")
	}
	if op.X != 100 || math.Abs(op.Opacity()-.72) > .001 || math.Abs(op.R-226.0/255) > .001 {
		t.Fatalf("sample x=%v opacity=%v rgb=%v,%v,%v", op.X, op.Opacity(), op.R, op.G, op.B)
	}
}

func TestStepUsesEasingInsideKeyframeInterval(t *testing.T) {
	op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect, Alpha: 1}
	p, err := New([]Target{{Op: op, From: cssmotion.Read(op)}}, []Keyframe{
		{Offset: 0, X: 0, Move: true}, {Offset: .5, X: 90, Move: true},
	}, Timing{Duration: time.Second, Iterations: 1, Easing: func(v float64) float64 { return v * v }})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(20, 0)
	p.Start(start)
	p.Step(start.Add(250 * time.Millisecond))
	if math.Abs(op.X-22.5) > .001 {
		t.Fatalf("x = %v, want 22.5", op.X)
	}
}
