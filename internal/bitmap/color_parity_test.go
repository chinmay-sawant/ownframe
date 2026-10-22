package bitmap_test

import (
	"context"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestTranslucentFillPixel(t *testing.T) {
	src := `<body style="margin:0"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`
	d, err := render.DisplayList(context.Background(), src, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := color.NRGBAModel.Convert(img.At(20, 20)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got != (color.NRGBA{R: 255, G: 127, B: 127, A: 255}) {
		t.Fatalf("pixel %+v", got)
	}
	op := &d.Ops[d.Order[0]]
	for i := range d.Ops {
		if d.Ops[i].R == 1 && d.Ops[i].G == 0 && d.Ops[i].B == 0 {
			op = &d.Ops[i]
		}
	}
	if masking.PaintNRGBA(op) != (color.NRGBA{R: 255, A: 128}) {
		t.Fatalf("policy %+v", masking.PaintNRGBA(op))
	}
}
