package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkGroupClip requires a partial DrawRect over a group to match the full
// draw inside the rect, byte for byte.
func checkGroupClip() error {
	source := `<body style="margin:0;height:140px;background:#e8e8e8">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0;` +
		`mix-blend-mode:multiply"><div style="position:absolute;left:30px;top:20px;width:60px;height:40px;` +
		`background:#c05030"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("clipped group should replay")
	}

	full, part := ebiten.NewImage(200, 140), ebiten.NewImage(200, 140)
	defer full.Deallocate()
	defer part.Deallocate()

	DrawVisible(full, d, 0, 0)
	rect := image.Rect(30, 30, 120, 90)
	DrawRect(part, d, rect, 0, 0)

	a, b := make([]byte, 200*140*4), make([]byte, 200*140*4)
	full.ReadPixels(a)
	part.ReadPixels(b)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		x0, x1 := (y*200+rect.Min.X)*4, (y*200+rect.Max.X)*4
		if !bytes.Equal(a[x0:x1], b[x0:x1]) {
			return fmt.Errorf("clipped group differs on row %d in %d channels", y, countDiff(a[x0:x1], b[x0:x1]))
		}
	}

	return nil
}
