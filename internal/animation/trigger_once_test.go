package animation

import (
	"testing"
	"time"
)

func TestPlayOnceUsesOneIteration(t *testing.T) {
	p, op, start := triggeredPlayer(t, Timing{Duration: time.Second, Iterations: 3, Fill: FillBoth})
	if err := p.TriggerEffect(0, "play-once", start); err != nil {
		t.Fatal(err)
	}
	if p.Step(start.Add(1500*time.Millisecond)) || op.X != 110 {
		t.Fatalf("play-once x=%v, active=%v; want final x=110 and finished", op.X, p.Active())
	}
}
