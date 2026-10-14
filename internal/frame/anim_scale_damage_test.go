package frame

import (
	"image"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestScaleDamageCoversTheLargerBox(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := ScaleTo(d, []*ownframe.DisplayOp{op}, 2, 100*time.Millisecond, Linear)
	a.Step(t0)
	dmg, _ := a.Step(t0.Add(50 * time.Millisecond))

	// At 1.5x the card spans 5..35 x 7.5..22.5 points, 5..47 x 9..31 CSS
	// px, wider than the 12..41 x 12..27 px it started in.
	if !image.Pt(45, 25).In(dmg.New) {
		t.Fatalf("new box %v misses the scaled corner", dmg.New)
	}

	if image.Pt(45, 25).In(dmg.Old) {
		t.Fatalf("old box %v already covers the scaled corner", dmg.Old)
	}
}
