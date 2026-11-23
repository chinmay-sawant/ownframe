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
	sizing  scratchSizing

	// horizontalUniforms and verticalUniforms are reused across draws so a
	// frame with a backdrop filter allocates no uniform maps. Only the render
	// loop draws through one renderer, so no lock guards them.
	horizontalUniforms map[string]any
	verticalUniforms   map[string]any
}

const (
	// blurUniformCount is the number of shader uniforms a blur pass sets.
	blurUniformCount = 9
	// scratchShrinkFrames is how many acquires one shrink decision watches.
	scratchShrinkFrames = 120
	// scratchShrinkRatio is how much larger than the window's peak demand the
	// scratch area may be before it is reallocated smaller.
	scratchShrinkRatio = 4
)

// scratchSizing decides when the three backdrop scratch textures grow, shrink,
// or are refused. It is pure state so aspect-ratio swings and the
// MaxImageSize clamp are testable without a GPU.
type scratchSizing struct {
	frames       int
	peakW, peakH int
	peakArea     int
}

// plan answers for one acquire of a w by h sample while the textures hold
// current. It returns the size the textures must have, and ok false when the
// request cannot be served (empty, or beyond maxEdge).
func (s *scratchSizing) plan(current image.Point, w, h, maxEdge int) (image.Point, bool) {
	if w <= 0 || h <= 0 || w > maxEdge || h > maxEdge {
		return current, false
	}

	s.frames++
	s.peakW, s.peakH = max(s.peakW, w), max(s.peakH, h)
	s.peakArea = max(s.peakArea, w*h)

	target := current
	if current.X < w || current.Y < h {
		target = image.Pt(min(max(current.X, w), maxEdge), min(max(current.Y, h), maxEdge))
	}

	if s.frames >= scratchShrinkFrames {
		if target.X*target.Y > scratchShrinkRatio*s.peakArea {
			target = image.Pt(s.peakW, s.peakH)
		}
		s.frames, s.peakW, s.peakH, s.peakArea = 0, 0, 0, 0
	}

	return target, true
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

	edge := ebiten.MaxImageSize()
	if edge <= 0 {
		// No running game reports no limit; leave the request unclamped.
		edge = math.MaxInt32
	}
	target, ok := b.sizing.plan(b.size, w, h, edge)
	if !ok {
		return nil, nil, nil, false
	}

	if b.scratch[0] == nil || target != b.size {
		for i, img := range b.scratch {
			if img != nil {
				img.Deallocate()
				b.scratch[i] = nil
			}
		}
		opts := &ebiten.NewImageOptions{Unmanaged: true}
		for i := range b.scratch {
			b.scratch[i] = ebiten.NewImageWithOptions(image.Rect(0, 0, target.X, target.Y), opts)
		}
		b.size = target
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

// blurUniforms fills the reused uniform map held in slot and returns it.
func (b *backdropState) blurUniforms(slot *map[string]any, radius, axis, saturation, refraction, boxX, boxY, boxW, boxH, cornerRadius float32) map[string]any {
	m := *slot
	if m == nil {
		m = make(map[string]any, blurUniformCount)
		*slot = m
	}
	m["Radius"], m["Axis"], m["Saturation"], m["Refraction"] = radius, axis, saturation, refraction
	m["BoxX"], m["BoxY"], m["BoxW"], m["BoxH"], m["CornerRadius"] = boxX, boxY, boxW, boxH, cornerRadius

	return m
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
	b.sizing = scratchSizing{}
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
	b.sizing = scratchSizing{}
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
	horizontalOptions.Uniforms = renderer.backdrop.blurUniforms(&renderer.backdrop.horizontalUniforms,
		float32(radius), 0, 1, 0, boxX, boxY, boxW, boxH, cornerRadius)
	horizontal.DrawRectShader(sample.Dx(), sample.Dy(), shader, horizontalOptions)

	verticalOptions := &ebiten.DrawRectShaderOptions{}
	verticalOptions.Images[0] = horizontal
	verticalOptions.Images[1] = backdrop
	verticalOptions.Blend = ebiten.BlendCopy
	verticalOptions.Uniforms = renderer.backdrop.blurUniforms(&renderer.backdrop.verticalUniforms,
		float32(radius), 1, float32(node.backdrop.Saturation), float32(refraction), boxX, boxY, boxW, boxH, cornerRadius)
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
