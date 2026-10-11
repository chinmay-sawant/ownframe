package bitmap_test

import (
	"context"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestBitmapBlendGroupCompositesOnce(t *testing.T) {
	source := `<body style="margin:0;background:blue"><div style="width:40px;height:40px;` +
		`background:red;mix-blend-mode:multiply"><div style="width:20px;height:20px;background:red"></div></div></body>`
	d, err := render.DisplayList(context.Background(), source, 80, 80)
	if err != nil {
		t.Fatal(err)
	}
	if render.Replayable(d) {
		t.Fatal("blend should use bitmap")
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	c := color.NRGBAModel.Convert(img.At(10, 10)).(color.NRGBA)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Fatal(c)
	}
}
