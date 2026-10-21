package replay

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

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

func TestTranslucentFillPixelGPU(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" {
		t.Skip("needs GPU")
	}
	src := `<body style="margin:0"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`
	d, err := render.DisplayList(context.Background(), src, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	img := ebiten.NewImage(100, 100)
	defer img.Deallocate()
	DrawVisible(img, d, 0, 0)
	px := make([]byte, 100*100*4)
	img.ReadPixels(px)
	at := (20*100 + 20) * 4
	for i, v := range [4]uint8{255, 127, 127, 255} {
		if c := int(px[at+i]) - int(v); c < -3 || c > 3 {
			t.Fatalf("pixel %v", px[at:at+4])
		}
	}
}
