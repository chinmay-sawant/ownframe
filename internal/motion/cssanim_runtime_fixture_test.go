package motion_test

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

const routeFrames = `@keyframes ten{from{transform:translateX(0px)}to{transform:translateX(10px)}}
@keyframes twenty{from{transform:translateX(0px)}to{transform:translateX(20px)}}`

func runtimePlayer(t *testing.T, declarations string) (*ownframe.Display, ownframe.Box, []int, []motion.Paint, *motion.Player, time.Time) {
	t.Helper()
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);` + declarations + `}` + routeFrames})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Probe().Display()
	var box ownframe.Box
	for _, candidate := range page.Boxes() {
		if candidate.ID == "target" {
			box = candidate
		}
	}
	indices, base := motion.OperationIndices(d, box), motion.Capture(d, box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatal("fixture has no target paint")
	}
	start := time.Unix(100, 0)
	p, err := motion.BindAnimation(d, box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	return d, box, indices, base, p, start
}

// movedX reports how far a sampled transform moved the operation along x. A
// transform-list keyframe writes the operation matrix, not the paint X.
func movedX(d *ownframe.Display, indices []int, base []motion.Paint) float64 {
	return motion.Read(&d.Ops[indices[0]]).Matrix.E - base[0].Matrix.E
}
