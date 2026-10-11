package bitmap_test

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestBitmapScrollKeepsFixedLayerAndPaintOrder(t *testing.T) {
	source := `<body style="margin:0"><div style="height:300px;background:red;border-radius:8px / 4px"></div>` +
		`<div style="position:fixed;top:0;left:0;width:20px;height:20px;background:blue"></div></body>`
	d, err := render.DisplayList(context.Background(), source, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, scroll := range []int{0, 100, 200} {
		img, err := bitmap.Viewport(d, image.Rect(0, scroll, 100, scroll+100), 1)
		if err != nil {
			t.Fatal(err)
		}
		got := color.NRGBAModel.Convert(img.At(10, 10)).(color.NRGBA)
		if got.B != 255 || got.R != 0 {
			t.Fatalf("scroll %d fixed pixel %+v", scroll, got)
		}
	}
	if d.Width != 100 || d.Ops[0].XformSet {
		t.Fatal("viewport changed retained display")
	}
}
