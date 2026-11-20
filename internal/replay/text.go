package replay

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/sfnt"
	"image"
	"math"
	"unicode/utf8"
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

	renderer.drawTextStroke(dst, face, content, x, y, op)

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

// bakeTextRun renders one run into a white mask anchored at the run's integer
// origin, so blitting the mask with the op color restores the direct draw.
// The run goes through the same glyph loop the direct path uses, shifted by
// the integer origin and the ink offset, so the bake carries the exact
// sub-pixel phase of the draw that first asked for it. The mask holds one
// strike only; the blit repeats it one pixel right for fake bold, which is
// the double strike the direct path draws. Baking white keeps color out of
// the raster, so equal runs in different colors share one entry.
func (renderer *Renderer) bakeTextRun(face *text.GoTextFace, key textRasterKey, offX, offY float64) (textRasterEntry, bool) {
	ink, found := textInkBox(face, key.content, key.spacing)
	if !found {
		return textRasterEntry{}, false
	}

	size, dx, dy, ok := textRasterLayout(ink)
	if !ok {
		return textRasterEntry{}, false
	}

	img := ebiten.NewImage(size.X, size.Y)
	renderer.drawWhiteTextRun(img, face, key.content, key.spacing, offX-float64(dx), offY-float64(dy))

	return textRasterEntry{img: img, dx: dx, dy: dy}, true
}

// drawWhiteTextRun draws one single-strike run in opaque white. Letter
// spacing advances glyph by glyph exactly as the direct painter does, so the
// mask matches the direct paint.
func (renderer *Renderer) drawWhiteTextRun(dst *ebiten.Image, face *text.GoTextFace, content string, spacing float64, x, y float64) {
	if spacing == 0 {
		options := &text.DrawOptions{}
		options.GeoM.Translate(x, y)
		options.ColorScale.ScaleWithColor(opaqueWhite)
		text.Draw(dst, content, face, options)

		return
	}

	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		options := &text.DrawOptions{}
		options.GeoM.Translate(x, y)
		options.ColorScale.ScaleWithColor(opaqueWhite)
		text.Draw(dst, glyph, face, options)
		width, _ := text.Measure(glyph, face, 0)
		x += width + spacing
	}
}

// textInkBox is the union of the run's glyph image rectangles, in run-local
// coordinates with the text origin at 0, 0. It reports false when no glyph
// paints, so an empty run is never cached. A letter-spaced run advances glyph
// by glyph exactly as the painter does, so the box matches the paint.
func textInkBox(face *text.GoTextFace, content string, spacing float64) (image.Rectangle, bool) {
	var box image.Rectangle
	found := false

	add := func(glyphs []text.LazyGlyph, offset float64) {
		shift := image.Pt(int(math.Floor(offset)), 0)
		for _, glyph := range glyphs {
			if glyph.ImageBounds.Empty() {
				continue
			}
			box = box.Union(glyph.ImageBounds.Add(shift))
			found = true
		}
	}

	if spacing == 0 {
		add(text.AppendLazyGlyphs(nil, content, face, nil), 0)

		return box, found
	}

	x := 0.0
	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		add(text.AppendLazyGlyphs(nil, glyph, face, nil), x)
		width, _ := text.Measure(glyph, face, 0)
		x += width + spacing
	}

	return box, found
}

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

// strokeRingPoints is how many offsets one ring of the outline uses.
const strokeRingPoints = 12

// strokeHalfWidthDivisor turns a stroke width into the radius of its centre ring.
// wideStrokeRadius is the radius in pixels above which a smaller inner ring is
// drawn too, and innerRingDivisor sets that inner ring to half the radius.
const (
	strokeHalfWidthDivisor = 2
	wideStrokeRadius       = 1.5
	innerRingDivisor       = 2
)

// drawTextStroke paints the -webkit-text-stroke outline under the fill. The
// engine carries the stroke width and color on the op (Op.TextStroke). Ebiten
// text cannot stroke a glyph path, so the run is drawn in the stroke color at
// points around a circle of half the stroke width, and again on a smaller
// ring for wide strokes. The fill is drawn after this, so only the half of
// the stroke outside the glyph shows. A real stroke is centered on the glyph
// edge and paints over the fill, so this outline is about half as heavy.
func (renderer *Renderer) drawTextStroke(
	dst *ebiten.Image, face *text.GoTextFace, content string, x, y float64, op *layout.DisplayOp,
) {
	width, color, ok := op.TextStroke()
	if !ok {
		return
	}

	stroked := *op
	stroked.R, stroked.G, stroked.B = color[0], color[1], color[2]

	radius := width * pxPerPt / strokeHalfWidthDivisor
	rings := []float64{radius}

	if radius > wideStrokeRadius {
		rings = append(rings, radius/innerRingDivisor)
	}

	for _, ring := range rings {
		for step := range strokeRingPoints {
			angle := 2 * math.Pi * float64(step) / strokeRingPoints
			renderer.drawTextRun(dst, face, content, x+ring*math.Cos(angle), y+ring*math.Sin(angle), &stroked)
		}
	}
}

// textProjection preserves the engine's already projected baseline while
// applying its remaining linear projection to the glyphs. Unlike a retained
// 2D patch, this metadata has XformSet false and must not move the baseline
// twice.
//
// The read goes through XformValue, a pointer-receiver accessor that returns the
// raw matrix without copying the op or binding the shared empty extra.
// Transform has a value receiver, so it copies the whole 256-byte op on every
// text run, which dominated the draw-rect scan. Reading the promoted Xform
// field directly is not an option: the payload is an embedded pointer, so a
// hand-built op leaves it nil, and binding the shared extra to guard the read
// would make Transform report a zero matrix where it documents an identity.
// internal/page/dirty compares Transform to diff a painted display against the
// next one, so that bind would dirty every text run on each redraw.
func textProjection(op *layout.DisplayOp) (layout.Matrix2D, bool) {
	if op.XformSet || (op.Kind != layout.DisplayOpText && op.Kind != layout.DisplayOpBullet) {
		return layout.Matrix2D{}, false
	}
	m := op.XformValue()
	if m == (layout.Matrix2D{}) || m.IsIdentity() {
		return layout.Matrix2D{}, false
	}
	m.E = op.X - m.A*op.X - m.C*op.Y
	m.F = op.Y - m.B*op.X - m.D*op.Y
	return m, true
}

// MaxUserFontBytes caps a user-supplied font at 2 MiB.
// MaxUserFontGlyphs caps its glyph count.
const (
	MaxUserFontBytes  = 2 * 1024 * 1024
	MaxUserFontGlyphs = 32768
)

// ErrFontPolicy means a user font breaks the untrusted font policy.
var ErrFontPolicy = errors.New("ownframe: font rejected by policy")

// ValidateUserFont allows only small parseable fonts with a bounded
// glyph count. Bundled and system fonts skip this; call it on any
// font bytes that arrive from untrusted input before parsing.
func ValidateUserFont(data []byte) error {
	if len(data) == 0 || len(data) > MaxUserFontBytes {
		return ErrFontPolicy
	}
	font, err := sfnt.Parse(data)
	if err != nil {
		return ErrFontPolicy
	}
	if font.NumGlyphs() > MaxUserFontGlyphs {
		return ErrFontPolicy
	}
	return nil
}
