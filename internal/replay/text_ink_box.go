package replay

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// textInkBox is the union of the run's glyph image rectangles, in run-local
// coordinates with the text origin at 0, 0. It reports false when no glyph
// paints, so an empty run is never cached. A letter-spaced run advances glyph
// by glyph exactly as the painter does, so the box matches the paint.
func textInkBox(face *text.GoTextFace, content string, spacing float64) (image.Rectangle, bool) {
	var box image.Rectangle
	found := false

	add := func(glyphs []text.LazyGlyph, offset float64) {
		shift := image.Pt(int(math.Floor(offset)), 0)
		for _, glyph := range glyphs {
			if glyph.ImageBounds.Empty() {
				continue
			}
			box = box.Union(glyph.ImageBounds.Add(shift))
			found = true
		}
	}

	if spacing == 0 {
		add(text.AppendLazyGlyphs(nil, content, face, nil), 0)

		return box, found
	}

	x := 0.0
	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		add(text.AppendLazyGlyphs(nil, glyph, face, nil), x)
		width, _ := text.Measure(glyph, face, 0)
		x += width + spacing
	}

	return box, found
}
