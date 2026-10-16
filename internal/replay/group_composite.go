package replay

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
)

// compositeLayer draws one buffered group onto dst at at, once, with the
// group's mix-blend-mode. A normal-mode layer is plain source-over, and so is
// a blended layer when the shader cannot compile; a blended layer with the
// shader reads the backdrop already on dst and rewrites the region with the
// CSS compositing formula.
func compositeLayer(dst *ebiten.Image, layer *ebiten.Image, at image.Point, mode string) {
	shift := &ebiten.DrawImageOptions{}
	shift.GeoM.Translate(float64(at.X), float64(at.Y))

	shader := (*ebiten.Shader)(nil)
	if !normalMode(mode) {
		shader = groupShader()
	}

	if shader == nil {
		dst.DrawImage(layer, shift)

		return
	}

	backdrop := ebiten.NewImage(layer.Bounds().Dx(), layer.Bounds().Dy())
	defer backdrop.Dispose()

	back := &ebiten.DrawImageOptions{}
	back.GeoM.Translate(
		float64(dst.Bounds().Min.X-at.X),
		float64(dst.Bounds().Min.Y-at.Y),
	)
	backdrop.DrawImage(dst, back)

	blend := &ebiten.DrawRectShaderOptions{}
	blend.GeoM = shift.GeoM
	blend.Images[0] = layer
	blend.Images[1] = backdrop
	blend.Blend = ebiten.BlendCopy
	blend.Uniforms = map[string]any{"Mode": float32(groupModeID(mode))}
	dst.DrawRectShader(layer.Bounds().Dx(), layer.Bounds().Dy(), shader, blend)
}

// normalMode reports a group that needs no blend step.
func normalMode(mode string) bool {
	return masking.BlendNormal(mode)
}

// groupModeID maps a CSS mix-blend-mode to the shader's mode number. Only
// the separable modes reach a compositor that asks.
func groupModeID(mode string) int {
	return masking.BlendModeID(mode)
}
