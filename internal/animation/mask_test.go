package animation

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestTargetMaskLimitsWhatAKeyframeWrites pins the field mask: a target only
// writes the fields its keyframes own, and a write never changes W, H, or the
// operation's alpha.
func TestTargetMaskLimitsWhatAKeyframeWrites(t *testing.T) {
	const (
		wantW      = 10
		wantH      = 10
		wantAlpha  = .4
		wantMatrix = 50
	)
	tests := []struct {
		name       string
		mask       cssmotion.Mask
		wantMatrix bool
		wantColor  bool
	}{
		{"matrix only", cssmotion.MaskMatrix, true, false},
		{"colour only", cssmotion.MaskRGB, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
				X: 1, Y: 2, W: wantW, H: wantH, R: .1, G: .2, B: .3, Alpha: wantAlpha}
			from := cssmotion.Read(op)
			from.Mask = test.mask

			effect := Effect{Frames: []Keyframe{
				{Offset: 0, Transform: TransformValue{Decomposed: Transform{Set: true}}},
				{Offset: 1, Transform: TransformValue{Decomposed: Transform{Tx: 100, Set: true}},
					R: 1, G: 1, B: 1, ColorSet: true},
			}, Timing: Timing{Duration: time.Second, Iterations: 1, Fill: FillBoth}}
			player, err := NewComposite([]Target{{Op: op, From: from}}, []Effect{effect})
			if err != nil {
				t.Fatalf("bind: %v", err)
			}
			start := time.Unix(0, 0)
			player.Start(start)
			player.Step(start.Add(500 * time.Millisecond))

			if op.XformSet != test.wantMatrix {
				t.Fatalf("matrix written = %v, want %v", op.XformSet, test.wantMatrix)
			}
			if test.wantMatrix && op.Transform().E != wantMatrix {
				t.Fatalf("translation = %v, want %v", op.Transform().E, wantMatrix)
			}
			if got := op.R != .1; got != test.wantColor {
				t.Fatalf("colour written = %v, want %v (R = %v)", got, test.wantColor, op.R)
			}
			if op.Alpha != wantAlpha {
				t.Fatalf("alpha = %v, want the untouched %v", op.Alpha, wantAlpha)
			}
			if op.W != wantW || op.H != wantH {
				t.Fatalf("size = %vx%v, want the untouched %vx%v", op.W, op.H, wantW, wantH)
			}
		})
	}
}
