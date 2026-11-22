package dirty_test

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// verticalTextHTML is one vertical run on a white page, so every non-white
// pixel the software painter lays down is that run's ink.
func verticalTextHTML(mode string) string {
	return `<body style="margin:0;background:#fff"><div style="position:absolute;left:60px;top:20px;` +
		`width:60px;height:150px;color:#000;font-size:18px;writing-mode:` + mode + `">Vertical text</div></body>`
}

// inkBounds returns the box of every pixel that differs from white.
func inkBounds(img image.Image) image.Rectangle {
	var ink image.Rectangle
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c, ok := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if !ok || c.R < 250 || c.G < 250 || c.B < 250 {
				ink = ink.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}

	return ink
}

// TestRotatedTextDirtyBoundsCoverSoftwareInk holds the dirty bounds of a
// quarter-turned run to the pixels the software painter draws for it. A
// partial redraw clips to these bounds, so ink outside them would go stale
// when the run changes.
func TestRotatedTextDirtyBoundsCoverSoftwareInk(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"vertical-rl", "vertical-lr"} {
		d, err := render.DisplayList(context.Background(), verticalTextHTML(mode), 240, 200)
		if err != nil {
			t.Fatal(err)
		}
		var run *layout.DisplayOp
		for i := range d.Ops {
			if d.Ops[i].Kind == layout.DisplayOpText && d.Ops[i].RotateDeg != 0 {
				run = &d.Ops[i]
			}
		}
		if run == nil {
			t.Fatalf("%s: the fixture emitted no rotated text run", mode)
		}
		img, err := bitmap.Paint(d)
		if err != nil {
			t.Fatal(err)
		}
		ink := inkBounds(img)
		if ink.Empty() {
			t.Fatalf("%s: the software painter drew no ink", mode)
		}
		box := dirty.OpBounds(run, d)
		if !ink.In(box) {
			t.Fatalf("%s: software ink %v falls outside the dirty bounds %v", mode, ink, box)
		}
		// The bounds follow the rotation: a tall run, not the upright strip.
		if box.Dy() <= box.Dx() {
			t.Fatalf("%s: dirty bounds %v are not the rotated run's tall box", mode, box)
		}
	}
}
