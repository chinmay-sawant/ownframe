package replay

import (
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

var asciiGlyphs [128]string //nolint:gochecknoglobals // reason: immutable ASCII lookup keeping glyphAt alloc-free, covered by TestGlyphAt_ASCIIZeroAllocs.

func init() {
	for i := 0; i < 128; i++ {
		asciiGlyphs[i] = string(rune(i))
	}
}

// glyphAt returns the single-rune substring at byte offset i, and its byte
// length, avoiding per-rune heap allocation.
func glyphAt(content string, i int) (string, int) {
	if b := content[i]; b < utf8.RuneSelf {
		return asciiGlyphs[b], 1
	}
	_, size := utf8.DecodeRuneInString(content[i:])
	return content[i : i+size], size
}

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

	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		renderer.drawRun(dst, glyph, face, x, y, op)
		width, _ := text.Measure(glyph, face, 0)
		x += width + op.LetterSpacing*pxPerPt
	}
}
