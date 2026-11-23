package replay

import (
	"bytes"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
	"image"
	"math"
	"sync"
)

// faceKey identifies one shaped face: the engine font value, the pixel size,
// and the language tag handed to the Ebiten face.
type faceKey struct {
	source   any
	size     float64
	language string
	// features is the op's OpenType feature list. Ebiten keys its glyph
	// cache on the face's features, so equal sources and sizes with different
	// features need separate faces.
	features string
}

// faceSourceCap bounds parsed font sources before a reset.
// faceCap bounds shaped faces before a reset.
const (
	faceSourceCap = 64
	faceCap       = 256
)

// textCache keeps one parsed face source per engine font. Hits resolve
// under a read lock; the parse runs lock-free with a second check that
// keeps one winner under the write lock.
type textCache struct {
	mu      sync.RWMutex
	sources map[any]*text.GoTextFaceSource
	faces   map[faceKey]*text.GoTextFace
}

func (renderer *Renderer) textFace(op *layout.DisplayOp) *text.GoTextFace {
	if op.Font == nil {
		return nil
	}

	size := op.Size * pxPerPt
	key := faceKey{source: op.Font, size: size, language: op.TextLanguage(), features: op.FontFeatures()}

	renderer.fonts.mu.RLock()
	source := renderer.fonts.sources[op.Font]
	if face := renderer.fonts.faces[key]; face != nil {
		renderer.fonts.mu.RUnlock()
		return face
	}
	renderer.fonts.mu.RUnlock()

	if source == nil {
		parsed, err := text.NewGoTextFaceSource(bytes.NewReader(op.Font.Bytes()))
		if err != nil {
			return nil
		}
		source = parsed
	}

	renderer.fonts.mu.Lock()
	defer renderer.fonts.mu.Unlock()

	if face := renderer.fonts.faces[key]; face != nil {
		return face
	}
	if have := renderer.fonts.sources[op.Font]; have != nil {
		source = have
	} else {
		if renderer.fonts.sources == nil || len(renderer.fonts.sources) >= faceSourceCap {
			renderer.fonts.sources = make(map[any]*text.GoTextFaceSource)
			renderer.fonts.faces = make(map[faceKey]*text.GoTextFace)
		}
		renderer.fonts.sources[op.Font] = source
	}

	face := &text.GoTextFace{Source: source, Size: size, Language: language.Make(key.language)}
	for _, feature := range masking.ParseFontFeatures(key.features) {
		if tag, err := text.ParseTag(feature.Tag); err == nil {
			face.SetFeature(tag, feature.Value)
		}
	}
	if len(renderer.fonts.faces) >= faceCap {
		clear(renderer.fonts.faces)
	}
	renderer.fonts.faces[key] = face

	return face
}

// cachedTextRun draws one run through the text raster cache and reports
// whether it painted. The raster is a white mask shared by every color, so
// the blit tints it with the op color. It reports false when the run keeps
// the direct path: an empty run, or one whose baked image would pass the
// size caps.
func (renderer *Renderer) cachedTextRun(dst *ebiten.Image, op *layout.DisplayOp, face *text.GoTextFace, content string, x, y float64) bool {
	if content == "" {
		return false
	}

	left, top := int(math.Floor(x)), int(math.Floor(y))
	offX, offY := x-float64(left), y-float64(top)
	key := textRasterKey{
		font:     op.Font,
		content:  content,
		language: op.TextLanguage(),
		size:     op.Size,
		spacing:  op.LetterSpacing * pxPerPt,
		gap:      op.TextAutospaceGap() * pxPerPt,
		features: op.FontFeatures(),
		fakeBold: layout.DisplayFakeBold(op),
		fx:       textPhase(offX),
		fy:       textPhase(offY),
	}

	entry, ok := renderer.textRuns.lookup(key)
	if !ok {
		entry = renderer.textRuns.raster(key, func() (textRasterEntry, bool) {
			return renderer.bakeTextRun(face, key, offX, offY)
		})
	}
	if entry.img == nil {
		return false
	}

	tint := rgba(op)
	renderer.blitRaster(dst, entry.img, left+entry.dx, top+entry.dy, tint)
	if key.fakeBold {
		renderer.blitRaster(dst, entry.img, left+entry.dx+1, top+entry.dy, tint)
	}

	return true
}

// textRasterMaxSide, textRasterMaxPixels, and textRasterPad bound one baked
// run. Pad is the integer slack per side around the glyph ink: it covers the
// sub-pixel placement of the glyph images and the rounding the face does when
// it quantizes the run origin to its sub-pixel grid. Past the caps the run
// draws directly.
const (
	textRasterMaxSide   = 512
	textRasterMaxPixels = 1 << 20
	textRasterPad       = 4
)

// textPhaseGrid is the sub-pixel grid the cache keys runs on. Snapping the
// offset to it absorbs the float64 noise that an integer scroll shift leaves
// in the origin, so a run scrolled by whole pixels keeps its key. The bake
// still uses the exact first origin, so a hit is a phase match to within
// 1/32 of a pixel.
const textPhaseGrid = 16

// textPhase snaps a sub-pixel offset to the keying grid.
func textPhase(frac float64) float64 {
	return math.Round(frac*textPhaseGrid) / textPhaseGrid
}

// textRasterLayout pads the run's ink box and returns the image size and the
// image's offset from the run origin. ok is false when the run is past the
// side or pixel caps and must draw directly.
func textRasterLayout(ink image.Rectangle) (size image.Point, dx, dy int, ok bool) {
	ink = ink.Inset(-textRasterPad)
	if ink.Dx() > textRasterMaxSide || ink.Dy() > textRasterMaxSide ||
		ink.Dx()*ink.Dy() > textRasterMaxPixels {
		return image.Point{}, 0, 0, false
	}

	return image.Pt(ink.Dx(), ink.Dy()), ink.Min.X, ink.Min.Y, true
}

// textRunMaxEntries and textRunMaxPixels bound what one page caches: at most
// 256 baked runs holding 2M pixels (8 MB of GPU bytes). A page with more
// runs than this keeps the most-recently-used; the least-recently-used
// rasters are disposed.
const (
	textRunMaxEntries = 256
	textRunMaxPixels  = 2 << 20
)

// textRasterEntry is one baked run, offset from the integer origin the key
// names. A zero entry is a remembered decline: the run cannot be cached, so
// later draws take the direct path without reshaping to learn that again.
type textRasterEntry struct {
	img    *ebiten.Image
	dx, dy int
}

// pixels is the entry's pixel count, zero for a decline.
func (e textRasterEntry) pixels() int {
	if e.img == nil {
		return 0
	}

	return e.img.Bounds().Dx() * e.img.Bounds().Dy()
}

// textRasterCache keeps baked text rasters, least-recently-used first. A hit
// moves to the back, so a repaint that reuses its runs never evicts them.
type textRasterCache struct {
	mu      sync.Mutex
	entries map[textRasterKey]textRasterEntry
	order   lruOrder[textRasterKey]
	pixels  int
}

// textRasterKey names one cached rendered run. It holds everything that
// shapes the pixels: the face identity, the content after text-transform, the
// effective size, the fake-bold double strike, the letter
// spacing in pixels, and the sub-pixel phase of the run's draw origin, on a
// 1/16 pixel grid. Color is not in the key: the raster bakes as a white mask
// and the blit tints it with the op color, so equal runs in different colors
// share one raster. The integer part of the origin is not in the key, because
// the raster bakes against the run's sub-pixel phase and blits at an integer
// origin, so equal runs at different integer positions share a raster. The
// cache entry carries the baked image's offset from that integer origin, so a
// hit never reshapes the run to find out where the ink lies.
type textRasterKey struct {
	font     any
	content  string
	language string
	size     float64
	spacing  float64
	// gap is the autospace advance at each ideograph and Latin boundary, in
	// pixels; features is the op's feature list. Both change the baked run.
	gap      float64
	features string
	fakeBold bool
	fx, fy   float64
}

// lookup returns the entry for key, if one has been baked or declined. A hit
// moves to the back of the eviction order, so the cache evicts the
// least-recently-used run first.
func (c *textRasterCache) lookup(key textRasterKey) (textRasterEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if ok {
		c.touch(key)
	}

	return entry, ok
}

// touch moves key to the back of the eviction order. The caller holds mu.
func (c *textRasterCache) touch(key textRasterKey) {
	c.order.touch(key)
}

// raster returns the entry for key, running bake once while the key is new.
// bake reports whether the run can be cached at all; a decline is remembered
// as a zero entry.
func (c *textRasterCache) raster(key textRasterKey, bake func() (textRasterEntry, bool)) textRasterEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[textRasterKey]textRasterEntry)
	}
	if entry, ok := c.entries[key]; ok {
		c.touch(key)

		return entry
	}

	entry, ok := bake()
	c.entries[key] = entry
	c.order.add(key)
	if ok {
		c.pixels += entry.pixels()
	}

	c.pixels = c.order.evict(c.order.length(), textRunMaxEntries, c.pixels, textRunMaxPixels, key, func(victim textRasterKey) int {
		evicted, exists := c.entries[victim]
		if !exists {
			return 0
		}
		px := evicted.pixels()
		if evicted.img != nil {
			evicted.img.Deallocate()
		}
		delete(c.entries, victim)
		return px
	})

	return entry
}

// flush drops every raster and every remembered decline. Only the benchmark
// needs a cold cache; a page reuses one for its whole life.
func (c *textRasterCache) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, entry := range c.entries {
		if entry.img != nil {
			entry.img.Deallocate()
		}
	}

	c.entries = nil
	c.order.reset()
	c.pixels = 0
}
