package replay

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// drawImage draws one decoded image op into its layout rectangle. The payload
// is decoded once and kept in the image cache, so later frames reuse the same
// Ebiten image. A nil decode (empty payload or bad bytes) skips the op. The
// gate rejects a non-identity transform only on non-image ops, so the
// placement is the op rectangle scaled from the decoded pixel bounds and then
// the op's own transform.
func (renderer *Renderer) drawImage(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	src := renderer.decodedImage(op)
	if src == nil || masking.SkipImageSize(op) {
		return
	}

	bounds := src.Bounds()

	options := &ebiten.DrawImageOptions{
		GeoM:   imageGeom(op, dx, dy, bounds.Dx(), bounds.Dy()),
		Filter: ebiten.FilterLinear,
	}

	if alpha := masking.ImageAlpha(op); alpha < 1 {
		options.ColorScale.ScaleAlpha(float32(alpha))
	}

	dst.DrawImage(src, options)
}
