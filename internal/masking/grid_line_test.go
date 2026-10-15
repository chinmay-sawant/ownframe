package masking_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestGridLinePropagatesOwner(t *testing.T) {
	t.Parallel()
	src := `<body style="margin:0"><table style="border-collapse:collapse"><tr><td style="border:2px solid red;width:40px;height:20px">x</td></tr></table></body>`
	d, err := render.DisplayList(context.Background(), src, 200, 140)
	if err != nil {
		t.Fatal(err)
	}
	var op *layout.DisplayOp
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpGridRun {
			op = &d.Ops[i]
		}
	}
	if op == nil || len(op.Grid.Segs) == 0 {
		t.Fatal("no gridrun")
	}
	op.ID, op.Alpha, op.Fixed, op.XformSet = 7, 0.5, true, true
	seg := op.Grid.Segs[0]
	line := masking.GridLine(op, 0)
	if line.Kind != layout.DisplayOpLine || line.Grid != nil {
		t.Fatalf("kind %+v", line.Kind)
	}
	if line.ID != 7 || line.Alpha != 0.5 || !line.Fixed || !line.XformSet {
		t.Fatalf("owner %+v", line)
	}
	if line.X != seg.X || line.W != seg.W || line.R != seg.R || line.Width != seg.Width {
		t.Fatalf("seg %+v want %+v", line, seg)
	}
}
