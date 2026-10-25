package cssanimation_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
)

// TestStaticPerspectiveFeedsAPartialKeyframe checks the whole propagation path:
// Blinkless resolves `perspective: 800px` on the element, BindBox hands that
// base to the keyframe builder, and a list with only a 50% stop starts at the
// static camera instead of none.
func TestStaticPerspectiveFeedsAPartialKeyframe(t *testing.T) {
	page, err := ownframe.New(ownframe.Config{
		HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);
			perspective:800px;perspective-origin:10px 20px;
			animation:camera 1s linear both}
			@keyframes camera{50%{perspective:500px}}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	d := page.Probe().Display()
	boxes := page.Boxes()
	index := -1
	var box ownframe.Box
	for i, candidate := range boxes {
		if candidate.ID == "target" {
			index, box = i, candidate
		}
	}
	if index < 0 {
		t.Fatal("the fixture has no target box")
	}
	if !box.Transform.HasPerspective {
		t.Fatalf("the engine did not report the static perspective: %+v", box.Transform)
	}

	start := time.Unix(300, 0)
	bounds, _, err := cssanimation.BindBox(d, box, index, start)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one", len(bounds))
	}

	bounds[0].Player.Step(start)
	values := bounds[0].Player.RestampValues()
	if len(values) != 1 {
		t.Fatalf("restamp values = %d, want one", len(values))
	}
	if !values[0].HasPerspective || !nearValue(values[0].Perspective, box.Transform.Perspective) {
		t.Fatalf("0%% perspective = %+v, want the static %v points",
			values[0], box.Transform.Perspective)
	}
	if !nearValue(values[0].PerspectiveOriginX, box.Transform.PerspectiveOriginX) ||
		!nearValue(values[0].PerspectiveOriginY, box.Transform.PerspectiveOriginY) {
		t.Fatalf("0%% origin = %v,%v, want the static %v,%v",
			values[0].PerspectiveOriginX, values[0].PerspectiveOriginY,
			box.Transform.PerspectiveOriginX, box.Transform.PerspectiveOriginY)
	}

	bounds[0].Player.Step(start.Add(500 * time.Millisecond))
	values = bounds[0].Player.RestampValues()
	if len(values) != 1 || !nearValue(values[0].Perspective, 375) {
		t.Fatalf("50%% perspective = %+v, want the declared 500px as 375 points", values)
	}
}

func nearValue(got, want float64) bool { return math.Abs(got-want) <= 1e-9 }
