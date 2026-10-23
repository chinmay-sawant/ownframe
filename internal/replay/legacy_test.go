package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	rasters  = &defaultRenderer.rasters
	fonts    = &defaultRenderer.fonts
	chunks   = &defaultRenderer.chunks
	textRuns = &defaultRenderer.textRuns
)

func chunkPlanFor(d *layout.Display) *chunkPlan {
	return defaultRenderer.chunkPlanFor(d)
}

func chunkRaster(d *layout.Display, p *chunkPlan, c *chunkRun) *chunkEntry {
	return defaultRenderer.chunkRaster(d, p, c)
}

func decodedImage(op *layout.DisplayOp) *ebiten.Image {
	return defaultRenderer.decodedImage(op)
}

func drawOp(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	defaultRenderer.drawOp(dst, op, dx, dy)
}

func drawTextRun(dst *ebiten.Image, face *text.GoTextFace, content string, x, y float64, op *layout.DisplayOp) {
	defaultRenderer.drawTextRun(dst, face, content, x, y, op)
}

func textFace(op *layout.DisplayOp) *text.GoTextFace {
	return defaultRenderer.textFace(op)
}
