package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

func checkGroupMask() error {
	data, err := os.ReadFile("../../examples/css-masking/assets/alpha-half.png")
	if err != nil {
		return err
	}
	source := `<body style="margin:0;height:140px;background:#159447">` +
		`<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;` +
		`background:#df3348;mask-image:url(mask.png);mask-mode:alpha;` +
		`mask-size:100% 100%;mask-repeat:no-repeat"></div></body>`
	d, err := render.DisplayListState(context.Background(), source, 200, 140,
		render.State{Images: func(string) ([]byte, error) { return data, nil }})
	if err != nil {
		return err
	}
	if !render.Replayable(d) {
		return fmt.Errorf("masked group should replay")
	}
	if err := compareGroupPixels(d, "alpha mask", [][2]int{{50, 50}, {150, 50}, {5, 5}}); err != nil {
		return err
	}
	full, part := ebiten.NewImage(200, 140), ebiten.NewImage(200, 140)
	defer full.Dispose()
	defer part.Dispose()
	DrawVisible(full, d, 0, 0)
	rect := image.Rect(35, 30, 125, 90)
	DrawRect(part, d, rect, 0, 0)
	a, b := make([]byte, 200*140*4), make([]byte, 200*140*4)
	full.ReadPixels(a)
	part.ReadPixels(b)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		lo, hi := (y*200+rect.Min.X)*4, (y*200+rect.Max.X)*4
		if !bytes.Equal(a[lo:hi], b[lo:hi]) {
			return fmt.Errorf("masked partial replay differs on row %d", y)
		}
	}
	return nil
}
