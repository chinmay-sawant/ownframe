package replay

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func gridRunOp(t *testing.T) (*render.Display, *layout.DisplayOp) {
	t.Helper()
	src := `<body style="margin:0"><table style="border-collapse:collapse"><tr><td style="border:4px solid red;width:40px;height:20px;opacity:.5">x</td></tr></table></body>`
	d, err := render.DisplayList(context.Background(), src, 200, 140)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpGridRun && len(d.Ops[i].Grid.Segs) > 0 {
			return d, &d.Ops[i]
		}
	}
	t.Fatal("no gridrun")
	return nil, nil
}

func TestTranslucentGridAgrees(t *testing.T) {
	d, op := gridRunOp(t)
	op.Alpha, op.Fixed = 0.5, true
	line := masking.GridLine(op, 0)
	if line.Grid != nil || line.Alpha != op.Alpha {
		t.Fatalf("line %+v owner %+v", line, op)
	}
	if rgba(&line).A != 128 {
		t.Fatalf("alpha %+v", rgba(&line))
	}
	_ = d
}

func TestTranslucentGridAgreesGPU(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" {
		t.Skip("needs GPU")
	}
	d, op := gridRunOp(t)
	op.Alpha = 0.5
	s := op.Grid.Segs[0]
	x := int((s.X + s.W/2) * d.PixelPerPoint)
	y := int((s.Y + s.H/2) * d.PixelPerPoint)
	want, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	img := ebiten.NewImage(200, 140)
	defer img.Dispose()
	DrawVisible(img, d, 0, 0)
	px := make([]byte, 200*140*4)
	img.ReadPixels(px)
	if err := closePixel(px, want, 200, x, y); err != nil {
		t.Fatal(err)
	}
}
