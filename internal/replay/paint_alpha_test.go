package replay

import (
	"bytes"
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func opaquePNG() ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func imageKind(d *render.Display) *layout.DisplayOp {
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpImage {
			return &d.Ops[i]
		}
	}
	return nil
}

func TestTranslucentImageAgrees(t *testing.T) {
	data, err := opaquePNG()
	if err != nil {
		t.Fatal(err)
	}
	src := `<body style="margin:0"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return data, nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	if masking.ImageAlpha(imageKind(d)) != 0.5 {
		t.Fatal("alpha")
	}
}
func TestTranslucentFillPixel(t *testing.T) {
	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, R: 1, Alpha: 0.5}
	if got := rgba(&op); got != masking.PaintRGBA(&op) || got.A != 128 || got.R != 128 {
		t.Fatalf("rgba %+v", got)
	}
	src := `<body style="margin:0"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`
	d, err := render.DisplayList(context.Background(), src, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d.Ops {
		if d.Ops[i].R == 1 && d.Ops[i].G == 0 && d.Ops[i].Kind == layout.DisplayOpFillRect {
			if rgba(&d.Ops[i]).A != 128 {
				t.Fatalf("alpha %+v", rgba(&d.Ops[i]))
			}
			return
		}
	}
	t.Fatal("no red op")
}

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
