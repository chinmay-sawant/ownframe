package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
	"math"
	"sync"
)

// rasterMaxEntries and rasterMaxTotalPixels bound what one page caches: at
// most 128 baked shape rasters holding 2M pixels (8 MB of GPU bytes). A
// page with more shapes than this keeps the newest; the oldest rasters are
// disposed.
const (
	rasterMaxEntries     = 128
	rasterMaxTotalPixels = 2 << 20
)

// rasterCache keeps baked vector rasters, oldest insertion first.
type rasterCache struct {
	mu      sync.Mutex
	entries map[rasterKey]*ebiten.Image
	order   []rasterKey
	pixels  int
}

var defaultRenderer Renderer

// raster returns the raster for key, baking it into an image of size once.
// The bake draws the shape so the origin's integer part is the image's
// top-left, which blitRaster then restores.
func (c *rasterCache) raster(key rasterKey, size image.Point, bake func(*ebiten.Image)) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()

	if img := c.entries[key]; img != nil {
		c.touch(key)
		return img
	}

	if c.entries == nil {
		c.entries = make(map[rasterKey]*ebiten.Image)
	}

	img := ebiten.NewImage(size.X, size.Y)
	bake(img)
	c.entries[key] = img
	c.order = append(c.order, key)
	c.pixels += size.X * size.Y

	for len(c.order) > rasterMaxEntries || c.pixels > rasterMaxTotalPixels {
		victim := rasterKey{}
		idx := -1
		for i, candidate := range c.order {
			if candidate == key {
				continue
			}
			victim = candidate
			idx = i
			break
		}
		if idx < 0 {
			// Only the entry being returned remains; keep it.
			break
		}

		c.order = append(c.order[:idx], c.order[idx+1:]...)
		evicted := c.entries[victim]
		delete(c.entries, victim)
		c.pixels -= evicted.Bounds().Dx() * evicted.Bounds().Dy()
		evicted.Deallocate()
	}

	return img
}

// touch moves a hit key to the back of the order, so eviction drops the
// least-recently-used entry first. The caller holds mu.
func (c *rasterCache) touch(key rasterKey) {
	for i, candidate := range c.order {
		if candidate == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, key)
			return
		}
	}
}

// rasterKey names one cached vector rasterization. It holds everything that
// shapes the pixels: the variant, the size and corner radii, the stroke width
// and mask, the color, and the sub-pixel offset of the op's origin. The
// integer part of the position is not in the key, because the raster bakes at
// the fractional offset and blits at the integer one, so equal shapes at
// different integer positions share one raster.
type rasterKey struct {
	variant uint8
	mask    uint8
	w, h    float32
	width   float32
	r, ry   [4]float32
	fx, fy  float32
}

// Raster variants.
const (
	rasterFill uint8 = iota + 1
	rasterStroke
)

// rasterPad is the integer slack per side that covers antialiasing spill.
const rasterPad = 2

// rasterMaxSide and rasterMaxPixels bound one cached raster. Past them the
// caller draws the shape directly.
const (
	rasterMaxSide   = 512
	rasterMaxPixels = 1 << 20
)

// rasterBox returns the integer blit origin and the image size for a shape
// whose filled box is x,y,w,h, grown by grow on every side. ok is false when
// the shape cannot be cached.
func rasterBox(x, y, w, h, grow float32) (left, top int, size image.Point, ok bool) {
	// Float32 cannot hold a small pixel offset next to a huge origin, so a
	// shape far from the page origin keeps the direct path.
	if x < -1e6 || y < -1e6 || x > 1e6 || y > 1e6 {
		return 0, 0, image.Point{}, false
	}

	left = int(math.Floor(float64(x-grow))) - rasterPad
	top = int(math.Floor(float64(y-grow))) - rasterPad
	right := int(math.Ceil(float64(x+w+grow))) + rasterPad
	bottom := int(math.Ceil(float64(y+h+grow))) + rasterPad
	size = image.Pt(right-left, bottom-top)
	if size.X <= 0 || size.Y <= 0 ||
		size.X > rasterMaxSide || size.Y > rasterMaxSide ||
		size.X*size.Y > rasterMaxPixels {
		return 0, 0, image.Point{}, false
	}

	return left, top, size, true
}

// opaqueWhite is the tint that leaves a baked white raster unchanged.
var opaqueWhite = color.RGBA{
	R: math.MaxUint8, G: math.MaxUint8, B: math.MaxUint8, A: math.MaxUint8,
}

// blitRaster copies a baked white raster to dst at an integer origin and
// applies the op's colour. An integer translation leaves every texel on its
// center, so the copy is exact. Baking the shape as a white mask lets one
// cached raster serve every colour, so a colour animation reuses it instead
// of re-rasterizing the path each frame.
func (renderer *Renderer) blitRaster(dst, img *ebiten.Image, left, top int, tint color.RGBA) {
	if renderer == nil {
		renderer = &defaultRenderer
	}
	opts := &renderer.drawOpts
	opts.GeoM.Reset()
	opts.ColorScale.Reset()
	opts.GeoM.Translate(float64(left), float64(top))
	opts.ColorScale.ScaleWithColor(tint)
	dst.DrawImage(img, opts)
}

func blitRaster(dst, img *ebiten.Image, left, top int, tint color.RGBA) {
	defaultRenderer.blitRaster(dst, img, left, top, tint)
}

// cachedRoundedFill draws a rounded fill through the raster cache. ok is
// false when the shape is too large to cache; the caller draws it directly.
func (renderer *Renderer) cachedRoundedFill(dst *ebiten.Image, op *layout.DisplayOp, radii [4]float64, dx, dy float64) bool {
	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	left, top, size, ok := rasterBox(x, y, w, h, 0)
	if !ok {
		return false
	}

	r := radius4(radii)
	key := rasterKey{
		variant: rasterFill,
		w:       w, h: h,
		r:  r,
		fx: x - float32(left),
		fy: y - float32(top),
	}
	img := renderer.rasters.raster(key, size, func(img *ebiten.Image) {
		path := acquireVectorPath()
		populateRoundedFillPath(path, key.fx, key.fy, w, h, r)

		var options vector.DrawPathOptions
		options.AntiAlias = true

		vector.FillPath(img, path, nil, &options)
		releaseVectorPath(path)
	})
	renderer.blitRaster(dst, img, left, top, rgba(op))

	return true
}

// cachedStroke draws a rounded or masked border through the raster cache. ok
// is false for a shape too large to cache or an unknown mask bit, which the
// caller handles as before.
func (renderer *Renderer) cachedStroke(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) bool {
	if op.StrokeMask&^0x0F != 0 {
		return false
	}

	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}

	left, top, size, ok := rasterBox(x, y, w, h, width/2+1)
	if !ok {
		return false
	}

	rx, ry := masking.RadiiXY(op)
	key := rasterKey{
		variant: rasterStroke,
		mask:    op.StrokeMask,
		w:       w, h: h, width: width,
		r:  radius4(rx),
		ry: radius4(ry),
		fx: x - float32(left),
		fy: y - float32(top),
	}
	img := renderer.rasters.raster(key, size, func(img *ebiten.Image) {
		path := acquireVectorPath()
		populateStrokePath(path, key)

		var stroke vector.StrokeOptions
		stroke.Width = width
		stroke.LineJoin = vector.LineJoinRound

		var options vector.DrawPathOptions
		options.AntiAlias = true

		vector.StrokePath(img, path, &stroke, &options)
		releaseVectorPath(path)
	})
	renderer.blitRaster(dst, img, left, top, rgba(op))

	return true
}
