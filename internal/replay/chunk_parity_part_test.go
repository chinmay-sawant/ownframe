package replay

import (
	"bytes"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

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

// runHolds reports a run holding an op with element opacity or alpha under
// one, or none when want is false.
func runHolds(p *chunkPlan, want bool) bool {
	for i := range p.items {
		if p.items[i].run == nil {
			continue
		}

		for _, op := range p.items[i].run.ops {
			if (op.Opacity() < 1) == want {
				return true
			}
		}
	}

	return false
}

// diffPixels describes the first differing pixels of two buffers, for the
// parity failure message.
func diffPixels(a, b []byte, width int) string {
	out := ""

	for i := 0; i+4 <= len(a) && len(out) < 400; i += 4 {
		if bytes.Equal(a[i:i+4], b[i:i+4]) {
			continue
		}

		p := i / 4
		out += fmt.Sprintf("(%d,%d) %v vs %v ", p%width, p/width, a[i:i+4], b[i:i+4])
	}

	return out
}
