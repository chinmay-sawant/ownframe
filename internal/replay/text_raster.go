package replay

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// cachedTextRun draws one run through the text raster cache and reports
// whether it painted. It reports false when the run keeps the direct path: an
// empty run, or one whose baked image would pass the size caps.
func (renderer *Renderer) cachedTextRun(dst *ebiten.Image, op *layout.DisplayOp, face *text.GoTextFace, content string, x, y float64) bool {
	if content == "" {
		return false
	}

	left, top := int(math.Floor(x)), int(math.Floor(y))
	offX, offY := x-float64(left), y-float64(top)
	key := textRasterKey{
		font:     op.Font,
		content:  content,
		language: op.TextLanguage(),
		size:     op.Size,
		spacing:  op.LetterSpacing * pxPerPt,
		col:      rgba(op),
		fakeBold: layout.DisplayFakeBold(op),
		fx:       textPhase(offX),
		fy:       textPhase(offY),
	}

	entry, ok := renderer.textRuns.lookup(key)
	if !ok {
		entry = renderer.textRuns.raster(key, func() (textRasterEntry, bool) {
			return renderer.bakeTextRun(op, face, key, offX, offY)
		})
	}
	if entry.img == nil {
		return false
	}

	blitRaster(dst, entry.img, left+entry.dx, top+entry.dy)
	if key.fakeBold {
		blitRaster(dst, entry.img, left+entry.dx+1, top+entry.dy)
	}

	return true
}
