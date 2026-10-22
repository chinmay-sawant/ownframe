package cssanimation

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const routeFrames = `@keyframes ten{from{transform:translateX(0px)}to{transform:translateX(10px)}}
@keyframes twenty{from{transform:translateX(0px)}to{transform:translateX(20px)}}`

func runtimePlayer(t *testing.T, declarations string) (*ownframe.Display, ownframe.Box, []int, []cssmotion.Paint, *animation.Player, time.Time) {
	t.Helper()
	page, err := ownframe.New(ownframe.Config{HTML: `<div id="target"></div>`, Width: 200, Height: 100,
		Theme: `#target{width:40px;height:20px;background-color:rgb(20,40,60);` + declarations + `}` + routeFrames})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := page.Testing().Display()
	var box ownframe.Box
	for _, candidate := range page.Boxes() {
		if candidate.ID == "target" {
			box = candidate
		}
	}
	indices, base := cssmotion.OperationIndices(d, box), cssmotion.Capture(d, box)
	if len(indices) == 0 || len(indices) != len(base) {
		t.Fatal("fixture has no target paint")
	}
	start := time.Unix(100, 0)
	p, err := Bind(d, box, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	return d, box, indices, base, p, start
}

func movedX(d *ownframe.Display, indices []int, base []cssmotion.Paint) float64 {
	return d.Ops[indices[0]].X - base[0].X
}
