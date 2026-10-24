package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
)

// compositeLayer draws one buffered group onto dst at at, once, with the
// group's mix-blend-mode. A normal-mode layer is plain source-over, and so is
// a blended layer when the shader cannot compile; a blended layer with the
// shader reads the backdrop already on dst and rewrites the region with the
// CSS compositing formula.
func compositeLayer(dst *ebiten.Image, layer *ebiten.Image, at image.Point, mode string) {
	var geo ebiten.GeoM
	geo.Translate(float64(at.X), float64(at.Y))
	compositeLayerImage(dst, layer, geo, 1, ebiten.FilterNearest, mode)
}

// compositeLayerPose draws a retained compositor layer through its animation
// matrix and group alpha. The texture already holds the base pose; geo maps it
// to the current page position. The blend shader path only accepts a
// translation, which resolveLayer enforces for non-normal modes.
func compositeLayerPose(dst *ebiten.Image, layer *ebiten.Image, geo ebiten.GeoM, alpha float64, mode string) {
	compositeLayerImage(dst, layer, geo, alpha, ebiten.FilterLinear, mode)
}

// compositeLayerImage is the shared group and layer composite. A non-normal
// mode reads dst as the backdrop and runs the blend shader over the layer's
// destination rectangle; any other mode is a source-over blit.
func compositeLayerImage(dst *ebiten.Image, layer *ebiten.Image, geo ebiten.GeoM,
	alpha float64, filter ebiten.Filter, mode string) {
	if dst == nil || layer == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{GeoM: geo, Filter: filter}
	if alpha < 1 {
		opts.ColorScale.ScaleAlpha(float32(alpha))
	}

	shader := (*ebiten.Shader)(nil)
	if !normalMode(mode) {
		shader = groupShader()
	}
	if shader == nil {
		dst.DrawImage(layer, opts)

		return
	}

	at, ok := geomTranslation(geo)
	if !ok {
		// The shader path cannot rotate or scale its destination rectangle.
		// resolveLayer keeps a blended layer a translation, so this is only a
		// defensive fallback for a caller that bypassed it.
		dst.DrawImage(layer, opts)

		return
	}

	backdrop := ebiten.NewImage(layer.Bounds().Dx(), layer.Bounds().Dy())
	defer backdrop.Deallocate()

	back := &ebiten.DrawImageOptions{}
	back.GeoM.Translate(
		float64(dst.Bounds().Min.X)-at[0],
		float64(dst.Bounds().Min.Y)-at[1],
	)
	backdrop.DrawImage(dst, back)

	blend := &ebiten.DrawRectShaderOptions{}
	blend.GeoM = geo
	blend.Images[0] = layer
	blend.Images[1] = backdrop
	blend.Blend = ebiten.BlendCopy
	if alpha < 1 {
		blend.ColorScale.ScaleAlpha(float32(alpha))
	}
	blend.Uniforms = map[string]any{"Mode": float32(groupModeID(mode))}
	dst.DrawRectShader(layer.Bounds().Dx(), layer.Bounds().Dy(), shader, blend)
}

// geomTranslation returns a GeoM's translation when its linear part is the
// identity.
func geomTranslation(geo ebiten.GeoM) ([2]float64, bool) {
	const eps = 1e-9
	close := func(a, b float64) bool {
		d := a - b
		if d < 0 {
			d = -d
		}

		return d < eps
	}
	if !close(geo.Element(0, 0), 1) || !close(geo.Element(0, 1), 0) ||
		!close(geo.Element(1, 0), 0) || !close(geo.Element(1, 1), 1) {
		return [2]float64{}, false
	}

	return [2]float64{geo.Element(0, geomTransCol), geo.Element(1, geomTransCol)}, true
}

// layerGeoM maps a layer texture's pixels into dst. Box.Min is the texture's
// page origin and Matrix is the animation delta in display point space, so its
// translation is converted from points to pixels before the page offset.
func layerGeoM(layer *Layer, d *layout.Display, dx, dy float64) ebiten.GeoM {
	var geo ebiten.GeoM
	if layer == nil {
		return geo
	}

	m := layer.Matrix
	ppp := layerPixelPerPoint(d)
	box := layer.Box

	geo.SetElement(0, 0, m.A)
	geo.SetElement(0, 1, m.C)
	geo.SetElement(0, geomTransCol, m.A*float64(box.Min.X)+m.C*float64(box.Min.Y)+m.E*ppp+dx)
	geo.SetElement(1, 0, m.B)
	geo.SetElement(1, 1, m.D)
	geo.SetElement(1, geomTransCol, m.B*float64(box.Min.X)+m.D*float64(box.Min.Y)+m.F*ppp+dy)

	return geo
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
