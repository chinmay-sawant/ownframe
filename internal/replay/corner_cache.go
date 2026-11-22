package replay

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// cornerMaskMaxRadius and cornerMaskMaxWidth cap CPU-baked analytical corner
// masks. Larger shapes fall back to vector paths.
const (
	cornerMaskMaxRadius = 128
	cornerMaskMaxWidth  = 64
	cornerCacheMaxItems = 256
	cornerCacheMaxPix   = 512 << 10
	cornerSubpixelScale = 32
	cornerAASteps       = 4
	cornerAASamples     = cornerAASteps * cornerAASteps
	cornerAARound       = cornerAASamples / 2
	cornerMaxAlpha      = 255
	cornerBytesPerPix   = 4
	cornerPadPx         = 1
	cornerPixelCenter   = 0.5
)

type cornerMaskKind uint8

const (
	cornerMaskFill cornerMaskKind = iota + 1
	cornerMaskStroke
)

// cornerMaskKey uniquely identifies one analytical white corner alpha mask
// quantized to 1/32 pixel so direct and raster-cached paths share identical
// masks without float32 translation drift.
type cornerMaskKey struct {
	kind   cornerMaskKind
	corner roundedCorner
	w, h   int16
	ox, oy int16
	pw, ph int16
	rx, ry int16
	width  int16
}

// cornerCache keeps small CPU-rasterized corner masks so rounded fills and
// borders draw as simple quads instead of pushing vector.FillPath/StrokePath
// vertex buffers into Ebiten's permanent graphics command queue.
type cornerCache struct {
	mu      sync.Mutex
	entries map[cornerMaskKey]*ebiten.Image
	order   []cornerMaskKey
	pixels  int
}

func quantizeCorner(v float32) int16 {
	q := math.Round(float64(v) * cornerSubpixelScale)
	if q < math.MinInt16 {
		return math.MinInt16
	}
	if q > math.MaxInt16 {
		return math.MaxInt16
	}
	return int16(q)
}

func (c *cornerCache) get(key cornerMaskKey, bake func() []byte) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()

	if img := c.entries[key]; img != nil {
		c.touch(key)
		return img
	}
	if c.entries == nil {
		c.entries = make(map[cornerMaskKey]*ebiten.Image)
	}

	w, h := int(key.w), int(key.h)
	img := ebiten.NewImage(w, h)
	img.WritePixels(bake())

	c.entries[key] = img
	c.order = append(c.order, key)
	c.pixels += w * h

	c.order, c.pixels = lruEvict(c.order, len(c.order), cornerCacheMaxItems, c.pixels, cornerCacheMaxPix, key, func(victim cornerMaskKey) int {
		evicted := c.entries[victim]
		delete(c.entries, victim)
		px := evicted.Bounds().Dx() * evicted.Bounds().Dy()
		evicted.Deallocate()
		return px
	})

	return img
}

func (c *cornerCache) touch(key cornerMaskKey) {
	c.order = lruTouch(c.order, key)
}

func (c *cornerCache) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.entries {
		img.Deallocate()
	}
	clear(c.entries)
	c.order = nil
	c.pixels = 0
}

func (c *cornerCache) releaseIdle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.entries {
		img.Deallocate()
	}
	c.entries = nil
	c.order = nil
	c.pixels = 0
}

func (c *cornerCache) warmIdle(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.entries == nil {
		c.entries = make(map[cornerMaskKey]*ebiten.Image, n)
	}
	if c.order == nil {
		c.order = make([]cornerMaskKey, 0, n)
	}
}

// fillCornerAnalytical blits a cached analytical quarter-circle fill mask for
// [px, px+pw] x [py, py+ph]. Returns false when the corner exceeds the mask
// cache size cap.
func (renderer *Renderer) fillCornerAnalytical(
	dst *ebiten.Image,
	px, py, pw, ph, rad float32,
	corner roundedCorner,
	clr color.RGBA,
) bool {
	if rad > cornerMaskMaxRadius || pw > cornerMaskMaxRadius+cornerPadPx*halfDiv || ph > cornerMaskMaxRadius+cornerPadPx*halfDiv {
		return false
	}

	ix0 := int(math.Floor(float64(px)))
	iy0 := int(math.Floor(float64(py)))
	ix1 := int(math.Ceil(float64(px + pw)))
	iy1 := int(math.Ceil(float64(py + ph)))
	iw, ih := ix1-ix0, iy1-iy0
	if iw <= 0 || ih <= 0 || iw > math.MaxInt16 || ih > math.MaxInt16 {
		return true
	}

	ox := px - float32(ix0)
	oy := py - float32(iy0)
	key := cornerMaskKey{
		kind:   cornerMaskFill,
		corner: corner,
		w:      int16(iw),
		h:      int16(ih),
		ox:     quantizeCorner(ox),
		oy:     quantizeCorner(oy),
		pw:     quantizeCorner(pw),
		ph:     quantizeCorner(ph),
		rx:     quantizeCorner(rad),
	}

	if renderer == nil {
		renderer = &defaultRenderer
	}
	img := renderer.corners.get(key, func() []byte {
		return bakeFillCornerMask(key)
	})
	renderer.blitRaster(dst, img, ix0, iy0, clr)
	return true
}

func bakeFillCornerMask(key cornerMaskKey) []byte {
	iw, ih := int(key.w), int(key.h)
	ox := float32(key.ox) / cornerSubpixelScale
	oy := float32(key.oy) / cornerSubpixelScale
	pw := float32(key.pw) / cornerSubpixelScale
	ph := float32(key.ph) / cornerSubpixelScale
	rad := float32(key.rx) / cornerSubpixelScale
	r2 := rad * rad

	var cx, cy float32
	switch key.corner {
	case cornerTL:
		cx, cy = ox+rad, oy+rad
	case cornerTR:
		cx, cy = ox+pw-rad, oy+rad
	case cornerBR:
		cx, cy = ox+pw-rad, oy+ph-rad
	case cornerBL:
		cx, cy = ox+rad, oy+ph-rad
	}

	pix := make([]byte, iw*ih*cornerBytesPerPix)
	const step = 1.0 / float32(cornerAASteps)
	const halfStep = step / 2

	for iy := range ih {
		pyCenter := float32(iy) + cornerPixelCenter
		if pyCenter < oy || pyCenter >= oy+ph {
			continue
		}
		for ix := range iw {
			pxCenter := float32(ix) + cornerPixelCenter
			if pxCenter < ox || pxCenter >= ox+pw {
				continue
			}

			hits := 0
			for syIdx := range cornerAASteps {
				sy := float32(iy) + float32(syIdx)*step + halfStep
				dy := sy - cy
				for sxIdx := range cornerAASteps {
					sx := float32(ix) + float32(sxIdx)*step + halfStep
					dx := sx - cx
					inCurveQuad := false
					switch key.corner {
					case cornerTL:
						inCurveQuad = dx < 0 && dy < 0
					case cornerTR:
						inCurveQuad = dx > 0 && dy < 0
					case cornerBR:
						inCurveQuad = dx > 0 && dy > 0
					case cornerBL:
						inCurveQuad = dx < 0 && dy > 0
					}
					if !inCurveQuad || dx*dx+dy*dy <= r2 {
						hits++
					}
				}
			}
			if hits == 0 {
				continue
			}
			a := uint8((hits*cornerMaxAlpha + cornerAARound) / cornerAASamples)
			off := (iy*iw + ix) * cornerBytesPerPix
			pix[off] = a
			pix[off+1] = a
			pix[off+2] = a
			pix[off+3] = a
		}
	}
	return pix
}

// strokeCornerAnalytical blits a cached analytical quarter-ellipse border arc
// mask for one corner quadrant. Returns false when the corner exceeds the mask
// cache size cap.
func (renderer *Renderer) strokeCornerAnalytical(
	dst *ebiten.Image,
	dstBounds image.Rectangle,
	cx, cy, rx, ry, width float32,
	corner roundedCorner,
	clr color.RGBA,
) bool {
	if rx > cornerMaskMaxRadius || ry > cornerMaskMaxRadius || width > cornerMaskMaxWidth {
		return false
	}
	half := width / halfDiv
	var minX, minY, maxX, maxY float32
	switch corner {
	case cornerTL:
		minX, maxX = cx-rx-half-cornerPadPx, cx+cornerPadPx
		minY, maxY = cy-ry-half-cornerPadPx, cy+cornerPadPx
	case cornerTR:
		minX, maxX = cx-cornerPadPx, cx+rx+half+cornerPadPx
		minY, maxY = cy-ry-half-cornerPadPx, cy+cornerPadPx
	case cornerBR:
		minX, maxX = cx-cornerPadPx, cx+rx+half+cornerPadPx
		minY, maxY = cy-cornerPadPx, cy+ry+half+cornerPadPx
	case cornerBL:
		minX, maxX = cx-rx-half-cornerPadPx, cx+cornerPadPx
		minY, maxY = cy-cornerPadPx, cy+ry+half+cornerPadPx
	}

	ix0 := int(math.Floor(float64(minX)))
	iy0 := int(math.Floor(float64(minY)))
	ix1 := int(math.Ceil(float64(maxX)))
	iy1 := int(math.Ceil(float64(maxY)))
	box := image.Rect(ix0, iy0, ix1, iy1)
	if !box.Overlaps(dstBounds) {
		return true
	}

	iw, ih := ix1-ix0, iy1-iy0
	if iw <= 0 || ih <= 0 || iw > math.MaxInt16 || ih > math.MaxInt16 {
		return true
	}

	// Store center relative to (ix0, iy0) in ox/oy.
	lcx := cx - float32(ix0)
	lcy := cy - float32(iy0)
	key := cornerMaskKey{
		kind:   cornerMaskStroke,
		corner: corner,
		w:      int16(iw),
		h:      int16(ih),
		ox:     quantizeCorner(lcx),
		oy:     quantizeCorner(lcy),
		rx:     quantizeCorner(rx),
		ry:     quantizeCorner(ry),
		width:  quantizeCorner(width),
	}

	if renderer == nil {
		renderer = &defaultRenderer
	}
	img := renderer.corners.get(key, func() []byte {
		return bakeStrokeCornerMask(key)
	})
	renderer.blitRaster(dst, img, ix0, iy0, clr)
	return true
}

func bakeStrokeCornerMask(key cornerMaskKey) []byte {
	iw, ih := int(key.w), int(key.h)
	cx := float32(key.ox) / cornerSubpixelScale
	cy := float32(key.oy) / cornerSubpixelScale
	rx := float32(key.rx) / cornerSubpixelScale
	ry := float32(key.ry) / cornerSubpixelScale
	half := float32(key.width) / cornerSubpixelScale / halfDiv

	outRX, outRY := rx+half, ry+half
	invOutX, invOutY := 1/outRX, 1/outRY
	hasInner := rx > half && ry > half
	var invInX, invInY float32
	if hasInner {
		invInX, invInY = 1/(rx-half), 1/(ry-half)
	}

	pix := make([]byte, iw*ih*cornerBytesPerPix)
	const step = 1.0 / float32(cornerAASteps)
	const halfStep = step / 2

	for iy := range ih {
		pyCenter := float32(iy) + cornerPixelCenter
		if (key.corner == cornerTL || key.corner == cornerTR) && pyCenter >= cy {
			continue
		}
		if (key.corner == cornerBL || key.corner == cornerBR) && pyCenter < cy {
			continue
		}
		for ix := range iw {
			pxCenter := float32(ix) + cornerPixelCenter
			if (key.corner == cornerTL || key.corner == cornerBL) && pxCenter >= cx {
				continue
			}
			if (key.corner == cornerTR || key.corner == cornerBR) && pxCenter < cx {
				continue
			}

			hits := 0
			for syIdx := range cornerAASteps {
				sy := float32(iy) + float32(syIdx)*step + halfStep
				dy := sy - cy
				for sxIdx := range cornerAASteps {
					sx := float32(ix) + float32(sxIdx)*step + halfStep
					dx := sx - cx

					oxn := dx * invOutX
					oyn := dy * invOutY
					if oxn*oxn+oyn*oyn > 1 {
						continue
					}
					if hasInner {
						ixn := dx * invInX
						iyn := dy * invInY
						if ixn*ixn+iyn*iyn < 1 {
							continue
						}
					}
					hits++
				}
			}
			if hits == 0 {
				continue
			}
			a := uint8((hits*cornerMaxAlpha + cornerAARound) / cornerAASamples)
			off := (iy*iw + ix) * cornerBytesPerPix
			pix[off] = a
			pix[off+1] = a
			pix[off+2] = a
			pix[off+3] = a
		}
	}
	return pix
}
