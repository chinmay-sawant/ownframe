package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"hash/fnv"
	"image"
	"sync"
	"unsafe"
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

	if op.XformSet && !op.Transform().IsIdentity() {
		clipped, paint := scissorOverflow(dst, op, dx, dy)
		if !paint {
			return
		}

		dst = clipped
	}

	options := &ebiten.DrawImageOptions{
		GeoM:   imageGeom(op, dx, dy, bounds.Dx(), bounds.Dy()),
		Filter: ebiten.FilterLinear,
	}

	if alpha := masking.ImageAlpha(op); alpha < 1 {
		options.ColorScale.ScaleAlpha(float32(alpha))
	}

	dst.DrawImage(src, options)
}

// decodedImage decodes op's payload once and caches the Ebiten image.
// Bytes to pixels, caps, and alpha live in masking; this fork only uploads.
// It returns nil for a skip or decode error, so the caller skips the draw.
func (renderer *Renderer) decodedImage(op *layout.DisplayOp) *ebiten.Image {
	if op == nil {
		return nil
	}
	key := renderer.images.keyFor(op)
	if img := renderer.images.get(key); img != nil {
		return img
	}
	src, _, skip, err := masking.DecodeImage(op)
	if skip || err != nil {
		return nil
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if masking.PixelsTooBig(w, h) {
		return nil
	}
	return renderer.images.put(key, src, w*h)
}

// imageCacheLimit bounds entries kept to 32 decoded images.
// maxImageCachePixels bounds total pixels to 16M (64 MB of GPU bytes), so
// many large images evict oldest first.
const (
	imageCacheLimit     = 32
	maxImageCachePixels = 16 << 20
	imageKeyRingSize    = 64
	imageKeyMinBytes    = 256
	imageKeySampleBytes = 16
)

type imageKeyEntry struct {
	ptr    uintptr
	length int
	isJPEG bool
	head   [imageKeySampleBytes]byte
	mid    [imageKeySampleBytes]byte
	tail   [imageKeySampleBytes]byte
	hash   uint64
}

// imageCache keeps decoded images by payload key, least recently used first.
// Hits take the read lock and stamp an atomic access tick; decode runs lock-free and only
// the GPU upload plus insert hold the write lock, with a second hit check.
type imageCache struct {
	mu      sync.RWMutex
	entries map[uint64]*ebiten.Image
	order   lruOrder[uint64]
	sizes   map[uint64]int
	total   int
	keyRing [imageKeyRingSize]imageKeyEntry
	keyNext int
}

// keyFor resolves op's payload hash, reusing a bounded ring cache for large
// buffers so unchunked image draws do not re-hash megabytes every frame.
// Head, mid, and tail byte samples guard against reused backing arrays.
func (c *imageCache) keyFor(op *layout.DisplayOp) uint64 {
	if op == nil {
		return 0
	}
	data, _, _ := op.ImageBytes()
	if len(data) < imageKeyMinBytes {
		return imageBytesKey(op)
	}

	// The address only keys the ring; the head/mid/tail samples below guard
	// against a reused backing array, and the pointer is never dereferenced.
	//nolint:gosec // reason: G103, address used as an opaque cache key and never dereferenced (reviewed, PERF-IMG-01).
	ptr := uintptr(unsafe.Pointer(unsafe.SliceData(data)))
	n := len(data)
	midStart := n/halfDiv - imageKeySampleBytes/halfDiv
	var head, mid, tail [imageKeySampleBytes]byte
	copy(head[:], data[:imageKeySampleBytes])
	copy(mid[:], data[midStart:midStart+imageKeySampleBytes])
	copy(tail[:], data[n-imageKeySampleBytes:])

	c.mu.RLock()
	for i := range c.keyRing {
		e := &c.keyRing[i]
		if e.ptr == ptr && e.length == n && e.isJPEG == op.IsJPEG &&
			e.head == head && e.mid == mid && e.tail == tail {
			h := e.hash
			c.mu.RUnlock()
			return h
		}
	}
	c.mu.RUnlock()

	h := imageBytesKey(op)
	c.mu.Lock()
	c.keyRing[c.keyNext] = imageKeyEntry{
		ptr: ptr, length: n, isJPEG: op.IsJPEG,
		head: head, mid: mid, tail: tail, hash: h,
	}
	c.keyNext = (c.keyNext + 1) % imageKeyRingSize
	c.mu.Unlock()

	return h
}

func (c *imageCache) get(key uint64) *ebiten.Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	img := c.entries[key]
	if img == nil {
		return nil
	}
	c.touch(key)
	return img
}

// touch stamps a hit key with a fresh access tick, so eviction drops the
// least recently used entry first. The caller holds at least the read lock.
func (c *imageCache) touch(key uint64) {
	c.order.touch(key)
}

func (c *imageCache) put(key uint64, src image.Image, px int) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()
	if img := c.entries[key]; img != nil {
		return img
	}
	img := ebiten.NewImageFromImage(src)
	if c.entries == nil {
		c.entries = make(map[uint64]*ebiten.Image)
		c.sizes = make(map[uint64]int)
	}
	c.entries[key] = img
	c.order.add(key)
	c.sizes[key] = px
	c.total += px
	c.total = c.order.evict(c.order.length(), imageCacheLimit, c.total, maxImageCachePixels, key, func(old uint64) int {
		if e := c.entries[old]; e != nil {
			e.Deallocate()
		}
		delete(c.entries, old)
		sz := c.sizes[old]
		delete(c.sizes, old)
		return sz
	})
	return img
}

// imageBytesKey identifies one encoded payload with FNV-1a. The IsJPEG flag is
// folded in so identical bytes tagged two ways never share an entry.
func imageBytesKey(op *layout.DisplayOp) uint64 {
	if op == nil {
		return 0
	}

	data, _, _ := op.ImageBytes()

	hash := fnv.New64a()
	if op.IsJPEG {
		_, _ = hash.Write([]byte{1})
	}

	_, _ = hash.Write(data)

	return hash.Sum64()
}

// geomTransCol is the GeoM column holding translation.
const geomTransCol = 2

// imageGeom maps the decoded image's w by h pixels onto op's point rectangle,
// moved by the page offset and then through the op's CSS transform when it has
// one. The gate allows transformed images.
func imageGeom(op *layout.DisplayOp, dx, dy float64, w, h int) ebiten.GeoM {
	var geom ebiten.GeoM

	k := pxPerPt
	sx, sy := 1/k, 1/k
	if w > 0 {
		sx = op.W / float64(w)
	}
	if h > 0 {
		sy = op.H / float64(h)
	}

	a, b, c, d, e, f := 1.0, 0.0, 0.0, 1.0, 0.0, 0.0
	if op.XformSet {
		m := op.Transform()
		a, b, c, d, e, f = m.A, m.B, m.C, m.D, m.E, m.F
	}

	geom.SetElement(0, 0, k*a*sx)
	geom.SetElement(0, 1, k*c*sy)
	geom.SetElement(0, geomTransCol, k*(a*op.X+c*op.Y+e)+dx)
	geom.SetElement(1, 0, k*b*sx)
	geom.SetElement(1, 1, k*d*sy)
	geom.SetElement(1, geomTransCol, k*(b*op.X+d*op.Y+f)+dy)

	return geom
}
