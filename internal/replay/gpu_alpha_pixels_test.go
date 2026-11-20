package replay

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

// Called only from parityGame.Draw, alongside the other pixel checks.
func checkTranslucentPixels() error {
	data, err := opaquePNG()
	if err != nil {
		return err
	}
	cases := []struct {
		name, source string
		grid         bool
	}{
		{name: "translucent fill", source: `<body style="margin:0;background:white;height:140px"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`},
		{name: "translucent image", source: `<body style="margin:0;background:white;height:140px"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`},
		{name: "translucent grid", source: `<body style="margin:0;background:white;height:140px"><table style="border-collapse:collapse"><tr><td style="border:4px solid red;width:40px;height:20px;opacity:.5">x</td></tr></table></body>`, grid: true},
	}
	for _, tc := range cases {
		d, err := render.DisplayListState(context.Background(), tc.source, 200, 140,
			render.State{Images: func(string) ([]byte, error) { return data, nil }})
		if err != nil {
			return err
		}
		point := [2]int{10, 10}
		if tc.grid {
			found := false
			for i := range d.Ops {
				op := &d.Ops[i]
				if op.Kind != layout.DisplayOpGridRun || len(op.Grid.Segs) == 0 {
					continue
				}
				op.Alpha = 0.5
				segment := op.Grid.Segs[0]
				point = [2]int{int((segment.X + segment.W/2) * d.PixelPerPoint), int((segment.Y + segment.H/2) * d.PixelPerPoint)}
				found = true
				break
			}
			if !found {
				return fmt.Errorf("translucent fixture produced no grid")
			}
		}
		if err := compareGroupPixels(d, tc.name, [][2]int{point, {190, 130}}); err != nil {
			return err
		}
	}
	return nil
}

func checkGPUOpacity() error {
	img := newParityImage(32, 32)
	defer img.Deallocate()
	img.Fill(color.White)
	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, W: 24, H: 24, R: .2, G: .4, B: .6, Alpha: .5}
	drawOp(img, &op, 0, 0)
	pixels := make([]byte, 32*32*4)
	img.ReadPixels(pixels)
	at := (10*32 + 10) * 4
	want := [4]byte{153, 178, 204, 255}
	for i, v := range want {
		delta := int(pixels[at+i]) - int(v)
		if delta < -1 || delta > 1 {
			return fmt.Errorf("opacity pixel %v want %v", pixels[at:at+4], want)
		}
	}
	return nil
}

// checkSleepyEyeRenders checks that uniform and mixed-corner fills paint.
func checkSleepyEyeRenders() error {
	mk := func(r [4]float64) *layout.DisplayOp {
		return &layout.DisplayOp{Kind: layout.DisplayOpFillRect,
			X: 244.2, Y: 215.5, W: 103.4, H: 53.4,
			R: 0.95, G: 0.96, B: 1, Alpha: 1, Radius: 54,
			RadiusTopLeft: r[0], RadiusTopRight: r[1],
			RadiusBottomRight: r[2], RadiusBottomLeft: r[3]}
	}
	lit := func(op *layout.DisplayOp) int {
		img := ebiten.NewImage(300, 160)
		defer img.Deallocate()
		drawOp(img, op, -200, -180)
		px := make([]byte, 300*160*4)
		img.ReadPixels(px)
		n := 0
		for i := 0; i < len(px); i += 4 {
			if px[i] > 0x80 {
				n++
			}
		}
		return n
	}
	mixed := lit(mk([4]float64{9, 9, 26.16, 26.16}))
	plain := lit(mk([4]float64{}))
	if mixed == 0 {
		return errors.New("mixed-corner fill painted nothing")
	}
	if plain == 0 {
		return errors.New("uniform fill painted nothing")
	}
	return nil
}
