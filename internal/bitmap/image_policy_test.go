package bitmap_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func redPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestImageBadPayloadErrors(t *testing.T) {
	src := `<body style="margin:0"><img src="bad.png" style="width:20px;height:20px"></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return []byte("not an image"), nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	var found *layout.DisplayOp
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpImage {
			found = &d.Ops[i]
		}
	}
	if found == nil {
		t.Fatal("no image op")
	}
	if _, _, skip, err := masking.DecodeImage(found); skip || err == nil {
		t.Fatalf("skip %v err %v", skip, err)
	}
	if _, err := bitmap.Paint(d); err == nil {
		t.Fatal("bad image painted")
	}
}

func TestTranslucentImagePixel(t *testing.T) {
	png := redPNG(t)
	src := `<body style="margin:0"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return png, nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := color.NRGBAModel.Convert(img.At(10, 10)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}
	if got.A != 255 || got.R != 255 || got.G < 120 || got.G > 135 {
		t.Fatalf("pixel %+v", got)
	}
}
