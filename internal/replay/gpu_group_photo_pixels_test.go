package replay

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func gpuDisplayPixels(d *render.Display) []byte {
	img := ebiten.NewImage(200, 140)
	defer img.Dispose()
	DrawVisible(img, d, 0, 0)
	pixels := make([]byte, 200*140*4)
	img.ReadPixels(pixels)
	return pixels
}

func compareGPUPixel(got, want []byte, x, y int, name string) error {
	at := (y*200 + x) * 4
	for i := 0; i < 4; i++ {
		delta := int(got[at+i]) - int(want[at+i])
		if delta < -3 || delta > 3 {
			return fmt.Errorf("%s pixel (%d,%d): got %v, want %v", name, x, y, got[at:at+4], want[at:at+4])
		}
	}
	return nil
}

func photoMaskHTML(overlay, masked bool) string {
	html := `<body style="margin:0;height:140px;background-image:url(assets/coast.jpg);background-size:cover;background-position:center">`
	if overlay {
		html += `<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;background-image:url(assets/flowers.jpg);background-size:cover;background-position:center;`
		if masked {
			html += `mask-image:url(mask.png);mask-mode:alpha;mask-size:100% 100%;mask-repeat:no-repeat;`
		}
		html += `"></div>`
	}
	return html + `</body>`
}
