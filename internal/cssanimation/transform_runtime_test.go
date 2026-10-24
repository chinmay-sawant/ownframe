package cssanimation_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestAnimationResolvesPercentageTranslateAtBind(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);` +
			`animation-name:pct;animation-duration:1s;animation-fill-mode:both;animation-timing-function:linear;}` +
			`@keyframes pct{from{transform:translate(0%,0%)}to{transform:translate(50%,25%)}}`})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Probe().Display()
	var box layout.Box
	for _, candidate := range page.Boxes() {
		if candidate.ID == "target" {
			box = candidate
		}
	}
	indices, base := cssmotion.OperationIndices(d, box), cssmotion.Capture(d, box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatal("fixture has no target paint")
	}
	start := time.Unix(200, 0)
	p, err := cssanimation.Bind(d, box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	p.Step(start.Add(time.Second))
	got := cssmotion.Read(&d.Ops[indices[0]]).Matrix
	if moved := (got.E - base[0].Matrix.E) / d.PointsPerPixel; math.Abs(moved-20) > .02 {
		t.Errorf("50%% translate moved %v px, want 20", moved)
	}
	if moved := (got.F - base[0].Matrix.F) / d.PointsPerPixel; math.Abs(moved-5) > .02 {
		t.Errorf("25%% translate moved %v px, want 5", moved)
	}
}
