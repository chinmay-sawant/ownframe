package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cardParityPage is the chunk parity fixture: translucent, rounded, plain cards.
func cardParityPage() (*layout.Display, error) {
	source := `<body style="margin:0;width:640px;height:200px;background:#ddeeff">` +
		`<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;` +
		`background:rgba(51,102,204,0.5)"></div>` +
		`<div style="position:absolute;left:220px;top:20px;width:160px;height:100px;` +
		`background:#ffddaa;border:3px solid #aa6622;border-radius:8px;padding:6px;font-size:13px">Card B detail</div>` +
		`<div style="position:absolute;left:420px;top:20px;width:160px;height:100px;` +
		`background:#ccffcc;border:1px solid #338833"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 640, 200)
	if err != nil {
		return nil, err
	}

	if !render.Replayable(d) {
		return nil, fmt.Errorf("the card fixture should replay")
	}

	return d, nil
}

// chunkPartialParity repaints one card's rect into a fresh image and requires
// it to match the full draw inside that rect, byte for byte.
func chunkPartialParity(d *layout.Display, full []byte) error {
	part := ebiten.NewImage(640, 200)
	defer part.Dispose()

	rect := image.Rect(30, 30, 130, 90)
	DrawRect(part, d, rect, 0, 0)
	pixels := make([]byte, 640*200*4)
	part.ReadPixels(pixels)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		x0, x1 := (y*640+rect.Min.X)*4, (y*640+rect.Max.X)*4
		if !bytes.Equal(full[x0:x1], pixels[x0:x1]) {
			return fmt.Errorf("chunked partial draw differs on row %d in %d channels",
				y, countDiff(full[x0:x1], pixels[x0:x1]))
		}
	}

	return nil
}

// chunkEntries reports the number of cached chunk images.
func chunkEntries() int {
	chunks.mu.Lock()
	defer chunks.mu.Unlock()

	return len(chunks.entries)
}
