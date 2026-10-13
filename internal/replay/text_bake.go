package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// bakeTextRun renders one run into a new image anchored at the run's integer
// origin, so blitting the image at that origin restores the direct draw. The
// run goes through the same glyph loop the direct path uses, shifted by the
// integer origin and the ink offset, so the bake carries the exact sub-pixel
// phase of the draw that first asked for it.
func bakeTextRun(op *layout.DisplayOp, face *text.GoTextFace, key textRasterKey, offX, offY float64) (textRasterEntry, bool) {
	ink, found := textInkBox(face, key.content, key.spacing)
	if !found {
		return textRasterEntry{}, false
	}

	size, dx, dy, ok := textRasterLayout(ink)
	if !ok {
		return textRasterEntry{}, false
	}

	img := ebiten.NewImage(size.X, size.Y)
	strike := op
	if key.fakeBold {
		// The bake holds one strike; the pad above covers the extra pixel the
		// second strike reaches. The second strike is the same raster blitted
		// one pixel right, which is the double strike the direct path draws.
		single := *op
		single.Bold = false
		strike = &single
	}
	drawTextRun(img, face, key.content, offX-float64(dx), offY-float64(dy), strike)

	return textRasterEntry{img: img, dx: dx, dy: dy}, true
}
