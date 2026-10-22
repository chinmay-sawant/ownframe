package bitmap_test

import (
	"context"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestTranslucentGridPixel(t *testing.T) {
	src := `<body style="margin:0"><table style="border-collapse:collapse"><tr><td style="border:4px solid red;width:40px;height:20px;opacity:.5">x</td></tr></table></body>`
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
	op.Alpha = 0.5
	line := masking.GridLine(op, 0)
	if line.Grid != nil || line.Kind != layout.DisplayOpLine {
		t.Fatal("expansion")
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	seg := op.Grid.Segs[0]
	x, y := int((seg.X+seg.W/2)*d.PixelPerPoint), int((seg.Y+seg.H/2)*d.PixelPerPoint)
	got, ok := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got.A != 255 || got.R < 200 {
		t.Fatalf("pixel %+v at %d,%d", got, x, y)
	}
}
