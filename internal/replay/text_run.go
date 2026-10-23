package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawRun draws one string at the text origin. The color comes from the op,
// and fake bold strikes the string twice one pixel apart. Both the direct
// path and the bake call this, so a baked run is the direct run shifted by an
// integer origin.
func (renderer *Renderer) drawRun(dst *ebiten.Image, content string, face *text.GoTextFace, x, y float64, op *layout.DisplayOp) {
	options := &text.DrawOptions{}
	options.GeoM.Translate(x, y)
	options.ColorScale.ScaleWithColor(rgba(op))

	text.Draw(dst, content, face, options)

	if layout.DisplayFakeBold(op) {
		options.GeoM.Translate(1, 0)
		text.Draw(dst, content, face, options)
	}
}
