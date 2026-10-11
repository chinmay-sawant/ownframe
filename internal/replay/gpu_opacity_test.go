package replay

import (
	"fmt"
	"image/color"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func checkGPUOpacity() error {
	img := ebiten.NewImage(32, 32)
	defer img.Dispose()
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
