package replay

import (
	_ "embed"
	"image"
	"math"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/backdrop"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed backdrop_blur.kage
var backdropBlurSource string

type backdropFilter = backdrop.Filter

var backdropShader struct {
	once sync.Once
	sh   *ebiten.Shader
	err  error
}

func parseBackdropFilter(value string) (backdropFilter, bool) {
	return backdrop.Parse(value)
}

func backdropFilterShader() (*ebiten.Shader, error) {
	backdropShader.once.Do(func() {
		shader, err := ebiten.NewShader([]byte(backdropBlurSource))
		backdropShader.sh, backdropShader.err = shader, err
	})
	return backdropShader.sh, backdropShader.err
}

func (renderer *Renderer) applyBackdropFilter(
	dst *ebiten.Image,
	display *layout.Display,
	node *groupNode,
	base [2]float64,
) {
	if dst == nil || display == nil || node == nil || !node.hasBackdrop || !node.hasBackdropBox {
		return
	}

	shader, err := backdropFilterShader()
	if err != nil || shader == nil {
		return
	}

	filterBox := translatedBackdropBox(node.backdropBox, base)
	paintBox := filterBox.Intersect(dst.Bounds())
	if paintBox.Empty() {
		return
	}

	scale := display.PointsPerPixel * display.PixelPerPoint
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	radius := node.backdrop.BlurPx * scale
	refraction := node.backdrop.RefractionPx * scale
	pad := int(math.Ceil(radius+refraction)) + 1
	sample := image.Rect(paintBox.Min.X-pad, paintBox.Min.Y-pad, paintBox.Max.X+pad, paintBox.Max.Y+pad).Intersect(dst.Bounds())
	if sample.Empty() {
		return
	}

	backdrop := ebiten.NewImage(sample.Dx(), sample.Dy())
	defer backdrop.Deallocate()
	horizontal := ebiten.NewImage(sample.Dx(), sample.Dy())
	defer horizontal.Deallocate()
	filtered := ebiten.NewImage(sample.Dx(), sample.Dy())
	defer filtered.Deallocate()

	copyBackdrop := &ebiten.DrawImageOptions{}
	copyBackdrop.GeoM.Translate(
		float64(dst.Bounds().Min.X-sample.Min.X),
		float64(dst.Bounds().Min.Y-sample.Min.Y),
	)
	backdrop.DrawImage(dst, copyBackdrop)

	boxX, boxY := float32(filterBox.Min.X-sample.Min.X), float32(filterBox.Min.Y-sample.Min.Y)
	boxW, boxH := float32(filterBox.Dx()), float32(filterBox.Dy())
	cornerRadius := float32(node.backdropRadius)
	if cornerRadius < 0 {
		cornerRadius = 0
	}

	horizontalOptions := &ebiten.DrawRectShaderOptions{}
	horizontalOptions.Images[0] = backdrop
	horizontalOptions.Blend = ebiten.BlendCopy
	horizontalOptions.Uniforms = map[string]any{
		"Radius":       float32(radius),
		"Axis":         float32(0),
		"Saturation":   float32(1),
		"Refraction":   float32(0),
		"BoxX":         boxX,
		"BoxY":         boxY,
		"BoxW":         boxW,
		"BoxH":         boxH,
		"CornerRadius": cornerRadius,
	}
	horizontal.DrawRectShader(sample.Dx(), sample.Dy(), shader, horizontalOptions)

	verticalOptions := &ebiten.DrawRectShaderOptions{}
	verticalOptions.Images[0] = horizontal
	verticalOptions.Images[1] = backdrop
	verticalOptions.Blend = ebiten.BlendCopy
	verticalOptions.Uniforms = map[string]any{
		"Radius":       float32(radius),
		"Axis":         float32(1),
		"Saturation":   float32(node.backdrop.Saturation),
		"Refraction":   float32(refraction),
		"BoxX":         boxX,
		"BoxY":         boxY,
		"BoxW":         boxW,
		"BoxH":         boxH,
		"CornerRadius": cornerRadius,
	}
	filtered.DrawRectShader(sample.Dx(), sample.Dy(), shader, verticalOptions)

	filterRegion := image.Rect(
		paintBox.Min.X-sample.Min.X,
		paintBox.Min.Y-sample.Min.Y,
		paintBox.Max.X-sample.Min.X,
		paintBox.Max.Y-sample.Min.Y,
	)
	filteredSurface, ok := filtered.SubImage(filterRegion).(*ebiten.Image)
	if !ok {
		return
	}
	drawFiltered := &ebiten.DrawImageOptions{}
	drawFiltered.GeoM.Translate(float64(paintBox.Min.X), float64(paintBox.Min.Y))
	drawFiltered.Blend = ebiten.BlendCopy
	dst.DrawImage(filteredSurface, drawFiltered)
}

// HasBackdropFilter reports whether the display has any backdrop-filter
// groups. Their output depends on other painted operations, so a host that
// cannot track those dependencies must replay the full display after damage.
func (renderer *Renderer) HasBackdropFilter(display *layout.Display) bool {
	if display == nil {
		return false
	}
	for i := range display.Ops {
		if group := display.Ops[i].Group(); group != nil && group.BackdropFilter != "" {
			return true
		}
	}

	return false
}

func translatedBackdropBox(box image.Rectangle, base [2]float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(float64(box.Min.X)+base[0])),
		int(math.Floor(float64(box.Min.Y)+base[1])),
		int(math.Ceil(float64(box.Max.X)+base[0])),
		int(math.Ceil(float64(box.Max.Y)+base[1])),
	)
}
