package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawText draws one shaped run at its baseline through the run raster cache,
// so a repaint that covers static text blits one baked image instead of
// reshaping and redrawing every glyph. A run the cache cannot hold takes the
// direct path below.
func (renderer *Renderer) drawText(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	face := renderer.textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	x := op.X*pxPerPt + dx
	y := op.Y*pxPerPt + dy - face.Metrics().HAscent

	if renderer.cachedTextRun(dst, op, face, content, x, y) {
		return
	}

	renderer.drawTextRun(dst, face, content, x, y, op)
}

// drawTextRun is the direct path. The engine stores X and Y as the baseline
// origin, while Ebiten's text origin is the top of the line, so the caller
// has already moved up by the face ascent. A letter-spaced run is drawn glyph
// by glyph with the extra advance.
func (renderer *Renderer) drawTextRun(dst *ebiten.Image, face *text.GoTextFace, content string, x, y float64, op *layout.DisplayOp) {
	if op.LetterSpacing == 0 {
		renderer.drawRun(dst, content, face, x, y, op)

		return
	}

	for _, r := range content {
		glyph := string(r)
		renderer.drawRun(dst, glyph, face, x, y, op)
		width, _ := text.Measure(glyph, face, 0)
		x += width + op.LetterSpacing*pxPerPt
	}
}
