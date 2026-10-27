package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// bakeTextRun renders one run into a white mask anchored at the run's integer
// origin, so blitting the mask with the op color restores the direct draw.
// The run goes through the same glyph loop the direct path uses, shifted by
// the integer origin and the ink offset, so the bake carries the exact
// sub-pixel phase of the draw that first asked for it. The mask holds one
// strike only; the blit repeats it one pixel right for fake bold, which is
// the double strike the direct path draws. Baking white keeps color out of
// the raster, so equal runs in different colors share one entry.
func (renderer *Renderer) bakeTextRun(face *text.GoTextFace, key textRasterKey, offX, offY float64) (textRasterEntry, bool) {
	ink, found := textInkBox(face, key.content, key.spacing)
	if !found {
		return textRasterEntry{}, false
	}

	size, dx, dy, ok := textRasterLayout(ink)
	if !ok {
		return textRasterEntry{}, false
	}

	img := ebiten.NewImage(size.X, size.Y)
	renderer.drawWhiteTextRun(img, face, key.content, key.spacing, offX-float64(dx), offY-float64(dy))

	return textRasterEntry{img: img, dx: dx, dy: dy}, true
}

// drawWhiteTextRun draws one single-strike run in opaque white. Letter
// spacing advances glyph by glyph exactly as the direct painter does, so the
// mask matches the direct paint.
func (renderer *Renderer) drawWhiteTextRun(dst *ebiten.Image, face *text.GoTextFace, content string, spacing float64, x, y float64) {
	if spacing == 0 {
		options := &text.DrawOptions{}
		options.GeoM.Translate(x, y)
		options.ColorScale.ScaleWithColor(opaqueWhite)
		text.Draw(dst, content, face, options)

		return
	}

	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		options := &text.DrawOptions{}
		options.GeoM.Translate(x, y)
		options.ColorScale.ScaleWithColor(opaqueWhite)
		text.Draw(dst, glyph, face, options)
		width, _ := text.Measure(glyph, face, 0)
		x += width + spacing
	}
}
