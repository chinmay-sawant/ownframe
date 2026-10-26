package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawXformText draws one shaped run through its baked transform. The op's
// baseline origin maps through the matrix, and the run's own geometry stays
// in text space, so a rotation tilts the glyphs instead of stair-stepping
// them. Letter spacing advances in text space, as in the direct painter.
func (renderer *Renderer) drawXformText(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	face := renderer.textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	dst, paint := scissorOverflow(dst, op, dx, dy)
	if !paint {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	geom := xformGeom(op, dx, dy)
	if !op.XformSet {
		geom.SetElement(0, geomTransCol, op.X*pxPerPt+dx)
		geom.SetElement(1, geomTransCol, op.Y*pxPerPt+dy)
	}
	shiftLocal(&geom, 0, -face.Metrics().HAscent)

	if op.LetterSpacing == 0 {
		renderer.drawXformRun(dst, content, face, geom, op)

		return
	}

	pen := 0.0

	for _, r := range content {
		glyph := string(r)
		at := geom
		shiftLocal(&at, pen, 0)
		renderer.drawXformRun(dst, glyph, face, at, op)

		width, _ := text.Measure(glyph, face, 0)
		pen += width + op.LetterSpacing*pxPerPt
	}
}

// drawXformRun draws one string at the text origin of geom. Fake bold strikes
// the string twice one local pixel apart, as the direct painter does.
func (renderer *Renderer) drawXformRun(dst *ebiten.Image, content string, face *text.GoTextFace, geom ebiten.GeoM, op *layout.DisplayOp) {
	options := &text.DrawOptions{}
	options.GeoM = geom
	options.ColorScale.ScaleWithColor(rgba(op))

	text.Draw(dst, content, face, options)

	if layout.DisplayFakeBold(op) {
		bold := geom
		shiftLocal(&bold, 1, 0)
		options.GeoM = bold
		text.Draw(dst, content, face, options)
	}
}
