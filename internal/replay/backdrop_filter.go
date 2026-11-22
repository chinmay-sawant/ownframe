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

// backdropState holds the lazily compiled backdrop blur shader and pooled
// scratch GPU images owned by one Renderer.
type backdropState struct {
	mu      sync.Mutex
	inited  bool
	shader  *ebiten.Shader
	err     error
	scratch [3]*ebiten.Image
	size    image.Point
}

func (b *backdropState) getShader() (*ebiten.Shader, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.inited {
		b.shader, b.err = ebiten.NewShader([]byte(backdropBlurSource))
		b.inited = true
	}
	return b.shader, b.err
}

func (b *backdropState) acquireScratch(w, h int) (src, horiz, filt *ebiten.Image, ok bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.scratch[0] == nil || b.size.X < w || b.size.Y < h {
		newW, newH := max(b.size.X, w), max(b.size.Y, h)
		for i, img := range b.scratch {
			if img != nil {
				img.Deallocate()
				b.scratch[i] = nil
			}
		}
		opts := &ebiten.NewImageOptions{Unmanaged: true}
		for i := range b.scratch {
			b.scratch[i] = ebiten.NewImageWithOptions(image.Rect(0, 0, newW, newH), opts)
		}
		b.size = image.Pt(newW, newH)
	}

	if b.size.X == w && b.size.Y == h {
		return b.scratch[0], b.scratch[1], b.scratch[2], true
	}
	sub := image.Rect(0, 0, w, h)
	s0, ok0 := b.scratch[0].SubImage(sub).(*ebiten.Image)
	s1, ok1 := b.scratch[1].SubImage(sub).(*ebiten.Image)
	s2, ok2 := b.scratch[2].SubImage(sub).(*ebiten.Image)
	if !ok0 || !ok1 || !ok2 {
		return nil, nil, nil, false
	}
	return s0, s1, s2, true
}

func (b *backdropState) releaseScratch() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, img := range b.scratch {
		if img != nil {
			img.Deallocate()
			b.scratch[i] = nil
		}
	}
	b.size = image.Point{}
}

func (b *backdropState) close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, img := range b.scratch {
		if img != nil {
			img.Deallocate()
			b.scratch[i] = nil
		}
	}
	b.size = image.Point{}
	if b.shader != nil {
		b.shader.Deallocate()
		b.shader = nil
	}
	b.err = nil
	b.inited = false
}

func parseBackdropFilter(value string) (backdropFilter, bool) {
	return backdrop.Parse(value)
}

func (renderer *Renderer) backdropFilterShader() (*ebiten.Shader, error) {
	return renderer.backdrop.getShader()
}

func backdropFilterShader() (*ebiten.Shader, error) {
	return defaultRenderer.backdropFilterShader()
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

	shader, err := renderer.backdropFilterShader()
	if err != nil || shader == nil {
		return
	}

	filterBox := translatedBackdropBox(node.backdropBox, base)
	paintBox := filterBox.Intersect(dst.Bounds())
	if paintBox.Empty() {
		return
	}

	radius := node.backdrop.BlurPx
	refraction := node.backdrop.RefractionPx
	pad := int(math.Ceil(radius+refraction)) + 1
	sample := image.Rect(paintBox.Min.X-pad, paintBox.Min.Y-pad, paintBox.Max.X+pad, paintBox.Max.Y+pad).Intersect(dst.Bounds())
	if sample.Empty() {
		return
	}

	backdrop, horizontal, filtered, ok := renderer.backdrop.acquireScratch(sample.Dx(), sample.Dy())
	if !ok {
		return
	}

	copyBackdrop := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
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
