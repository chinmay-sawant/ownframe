package replay

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func opaquePNG(t *testing.T) []byte {
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

func imageKind(d *render.Display) *layout.DisplayOp {
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpImage {
			return &d.Ops[i]
		}
	}
	return nil
}

func TestTranslucentImageAgrees(t *testing.T) {
	data := opaquePNG(t)
	src := `<body style="margin:0"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return data, nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	if masking.ImageAlpha(imageKind(d)) != 0.5 {
		t.Fatal("alpha")
	}
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" {
		t.Skip("needs GPU for pixels")
	}
	want, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	img := ebiten.NewImage(100, 100)
	defer img.Dispose()
	DrawVisible(img, d, 0, 0)
	px := make([]byte, 100*100*4)
	img.ReadPixels(px)
	if err := closePixel(px, want, 100, 10, 10); err != nil {
		t.Fatal(err)
	}
}
