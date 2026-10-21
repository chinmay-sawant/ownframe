package replay

import (
	"bytes"
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
)

// checkXformParity draws a transformed page through drawOp and DrawVisible
// and requires identical pixels, then checks fills against the bitmap and
// the rotated run's placement.
func checkXformParity() error {
	source := `<body style="margin:0;height:140px;background:#f4f4f4">` +
		`<div style="position:absolute;left:20px;top:20px;width:100px;height:50px;background:#2860c8;transform:rotate(25deg)"></div>` +
		`<div style="position:absolute;left:150px;top:20px;width:60px;height:30px;background:#2ca04a;transform:skewX(20deg) scale(1.2,1)"></div>` +
		`<div style="position:absolute;left:20px;top:80px;width:90px;height:36px;border:3px solid #204020;transform:rotate(-15deg)"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 240, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("transformed page should replay")
	}

	if err := checkXformRepeat(d); err != nil {
		return err
	}

	before, after := ebiten.NewImage(240, 140), ebiten.NewImage(240, 140)
	defer before.Deallocate()
	defer after.Deallocate()

	for i := range d.Order {
		drawOp(before, &d.Ops[i], 0, 0)
	}

	DrawVisible(after, d, 0, 0)

	a, b := make([]byte, 240*140*4), make([]byte, 240*140*4)
	before.ReadPixels(a)
	after.ReadPixels(b)

	if !bytes.Equal(a, b) {
		return fmt.Errorf("transformed draw differs in %d channels", countDiff(a, b))
	}

	if err := checkXformBitmap(d, after); err != nil {
		return err
	}

	return checkXformText()
}
