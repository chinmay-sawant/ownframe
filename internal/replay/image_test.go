package replay

import (
	"bytes"
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"image/png"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

// tinyPNG builds a 2x2 PNG whose first pixel carries shade, enough to
// exercise decode and caching without a fixture file.
func tinyPNG(t *testing.T, shade uint8) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: shade, A: 255})

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatalf("png: %v", err)
	}

	return out.Bytes()
}

// imageOp builds an op carrying an encoded payload. A display-list consumer
// cannot attach one through the public API, so the test reaches the embedded
// payload directly and allocates a fresh extra instead of mutating the
// engine's shared empty extra.
func imageOp(t *testing.T, data []byte, w, h int, isJPEG bool) layout.DisplayOp {
	t.Helper()

	op := layout.DisplayOp{Kind: layout.DisplayOpImage, IsJPEG: isJPEG}
	root := reflect.ValueOf(&op).Elem()
	extra := reflect.New(root.FieldByName("opExtra").Type().Elem())

	fields := extra.Elem()
	set(t, fields.FieldByName("Image"), reflect.ValueOf(data))
	set(t, fields.FieldByName("ImgW"), reflect.ValueOf(w))
	set(t, fields.FieldByName("ImgH"), reflect.ValueOf(h))
	set(t, root.FieldByName("opExtra"), extra)

	return op
}

// set writes one payload field that reflect refuses to set because the payload
// type is unexported.
func set(t *testing.T, field, value reflect.Value) {
	t.Helper()

	if !field.IsValid() {
		t.Fatal("payload field missing")
	}

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(value) //nolint:gosec // reason: test-only write to an unexported field.
}

func TestImageBadPayloadSkips(t *testing.T) {
	src := `<body style="margin:0"><img src="bad.png" style="width:20px;height:20px"></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return []byte("not an image"), nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	op := imageKind(d)
	if op == nil {
		t.Fatal("no image op")
	}
	if _, _, skip, err := masking.DecodeImage(op); skip || err == nil {
		t.Fatalf("skip %v err %v", skip, err)
	}
	if decodedImage(op) != nil {
		t.Fatal("bad payload decoded")
	}
}

// testImage builds a small in-memory image for cache-level tests without an
// encoded payload, so the tests do not depend on shared renderer state.
func testImage(w, h int) image.Image {
	return image.NewNRGBA(image.Rect(0, 0, w, h))
}

func TestDecodedImageCacheHitAndMiss(t *testing.T) {
	data := tinyPNG(t, 255)
	op := imageOp(t, data, 2, 2, false)

	first := decodedImage(&op)
	if first == nil {
		t.Fatal("valid PNG did not decode")
	}

	same := imageOp(t, append([]byte(nil), data...), 2, 2, false)
	if decodedImage(&same) != first {
		t.Fatal("identical payload was not cached")
	}

	other := imageOp(t, tinyPNG(t, 7), 2, 2, false)
	if decodedImage(&other) == first {
		t.Fatal("different payloads share a decoded image")
	}
}

func TestDecodedImageSkipsBadPayloads(t *testing.T) {
	if decodedImage(nil) != nil {
		t.Fatal("nil op decoded")
	}

	empty := layout.DisplayOp{Kind: layout.DisplayOpImage}
	if decodedImage(&empty) != nil {
		t.Fatal("empty payload decoded")
	}

	bad := imageOp(t, []byte("not an image"), 1, 1, false)
	if decodedImage(&bad) != nil {
		t.Fatal("bad payload decoded")
	}
}

func TestDecodedImageEvictsOldest(t *testing.T) {
	target := imageOp(t, tinyPNG(t, 200), 2, 2, false)

	first := decodedImage(&target)
	if first == nil {
		t.Fatal("target did not decode")
	}

	for i := range imageCacheLimit {
		filler := imageOp(t, tinyPNG(t, uint8(100+i)), 2, 2, false)
		if decodedImage(&filler) == nil {
			t.Fatalf("filler %d did not decode", i)
		}
	}

	if decodedImage(&target) == first {
		t.Fatal("oldest entry was not evicted")
	}
}

func TestImageCacheHitRefreshesRecency(t *testing.T) {
	var c imageCache
	for i := range imageCacheLimit {
		c.put(uint64(i+1), testImage(4, 4), 16)
	}
	if c.get(1) == nil {
		t.Fatal("first entry missing before refresh")
	}
	c.put(uint64(imageCacheLimit+1), testImage(4, 4), 16)

	if c.get(1) == nil {
		t.Fatal("recently used entry was evicted before an untouched one")
	}
	if c.get(2) != nil {
		t.Fatal("least recently used entry survived eviction")
	}
}

func TestImageCacheCapsTotalPixels(t *testing.T) {
	var c imageCache
	third := maxImageCachePixels/3 + 1
	c.put(1, testImage(4, 4), third)
	c.put(2, testImage(4, 4), third)
	c.put(3, testImage(4, 4), third)

	if c.get(1) != nil {
		t.Fatal("pixel cap did not evict the oldest entry")
	}
	if c.total > maxImageCachePixels {
		t.Fatalf("cache holds %d pixels over the %d cap", c.total, maxImageCachePixels)
	}
	if c.order.length() != 2 {
		t.Fatalf("cache holds %d entries, want 2 after pixel eviction", c.order.length())
	}
}

func TestImageGeomScalesAndTranslates(t *testing.T) {
	op := layout.DisplayOp{X: 10, Y: 20, W: 30, H: 40}
	geom := imageGeom(&op, 5, -3, 15, 20)

	x, y := geom.Apply(0, 0)
	assertNear(t, x, op.X*pxPerPt+5)
	assertNear(t, y, op.Y*pxPerPt-3)

	x, y = geom.Apply(15, 20)
	assertNear(t, x, (op.X+op.W)*pxPerPt+5)
	assertNear(t, y, (op.Y+op.H)*pxPerPt-3)
}

func TestImageGeomKeepsTranslateWhenBoundsMissing(t *testing.T) {
	op := layout.DisplayOp{X: 4, Y: 6, W: 30, H: 40}
	geom := imageGeom(&op, 1, 2, 0, 0)

	x, y := geom.Apply(15, 20)
	assertNear(t, x, 15+op.X*pxPerPt+1)
	assertNear(t, y, 20+op.Y*pxPerPt+2)
}

func TestImageGeomAppliesTransform(t *testing.T) {
	source := `<html><body><div style="width:20px;height:20px;` +
		`background:linear-gradient(red,blue);transform:rotate(20deg)"></div></body></html>`

	display, err := render.DisplayList(context.Background(), source, 320, 200)
	if err != nil {
		t.Fatal(err)
	}

	for _, op := range display.Ops {
		if op.Kind != layout.DisplayOpImage || !op.XformSet {
			continue
		}

		geom := imageGeom(&op, 0, 0, 10, 10)
		_, y0 := geom.Apply(0, 0)
		_, y1 := geom.Apply(10, 0)
		if math.Abs(y1-y0) < 1e-6 {
			t.Fatal("rotation did not tilt the top edge")
		}

		return
	}

	t.Fatal("no transformed image op")
}

func assertNear(t *testing.T, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestImageBytesKeyStableAndDistinct(t *testing.T) {
	data := tinyPNG(t, 255)

	first := imageOp(t, data, 2, 2, false)
	same := imageOp(t, append([]byte(nil), data...), 2, 2, false)
	if imageBytesKey(&first) != imageBytesKey(&same) {
		t.Fatal("same payload produced different keys")
	}

	other := imageOp(t, tinyPNG(t, 7), 2, 2, false)
	if imageBytesKey(&first) == imageBytesKey(&other) {
		t.Fatal("different payloads share a key")
	}

	jpeg := imageOp(t, data, 2, 2, true)
	if imageBytesKey(&first) == imageBytesKey(&jpeg) {
		t.Fatal("IsJPEG is not folded into the key")
	}

	if imageBytesKey(nil) != 0 {
		t.Fatal("nil op key is not zero")
	}
}

func TestImageCachePutDoesNotSelfEvict(t *testing.T) {
	t.Parallel()

	var c imageCache
	oversizedPx := maxImageCachePixels + 1024
	img := c.put(42, testImage(4, 4), oversizedPx)
	if img == nil {
		t.Fatal("put returned nil")
	}
	if c.get(42) != img {
		t.Fatal("newly inserted image exceeding pixel cap self-evicted")
	}
	if c.order.length() != 1 {
		t.Fatalf("len(order) = %d, want 1", c.order.length())
	}
}

func TestImageCacheKeyForRingMatchesFullHash(t *testing.T) {
	t.Parallel()

	var c imageCache
	buf := bytes.Repeat([]byte("ownframe-image-payload-"), 32) // 736 bytes >= 256
	op := imageOp(t, buf, 4, 4, false)

	want := imageBytesKey(&op)
	if got := c.keyFor(&op); got != want {
		t.Fatalf("first keyFor = %d, want %d", got, want)
	}
	if got := c.keyFor(&op); got != want {
		t.Fatalf("cached keyFor = %d, want %d", got, want)
	}

	jpegOp := imageOp(t, buf, 4, 4, true)
	if got := c.keyFor(&jpegOp); got == want || got != imageBytesKey(&jpegOp) {
		t.Fatalf("jpeg keyFor = %d, want %d (distinct from %d)", got, imageBytesKey(&jpegOp), want)
	}

	// Mutating the tail in place must miss the ring cache and recompute.
	buf[len(buf)-1] ^= 0xff
	if got := c.keyFor(&op); got == want || got != imageBytesKey(&op) {
		t.Fatalf("mutated keyFor = %d, want %d", got, imageBytesKey(&op))
	}
}

// TestImageCacheKeyForRingHitAllocsZero pins PERF-IMG-01: a ring hit no
// longer boxes the slice header through reflect.
func TestImageCacheKeyForRingHitAllocsZero(t *testing.T) {
	var c imageCache
	buf := bytes.Repeat([]byte("ownframe-image-payload-"), 32)
	op := imageOp(t, buf, 4, 4, false)
	c.keyFor(&op)
	if allocs := testing.AllocsPerRun(100, func() { c.keyFor(&op) }); allocs != 0 {
		t.Fatalf("keyFor ring hit allocs/op = %v, want 0", allocs)
	}
}

func TestImageGeomBounds(t *testing.T) {
	t.Parallel()

	// 1x60 linear gradient stretched to 140x60 at (10, 20).
	var stretch ebiten.GeoM
	stretch.Scale(140, 1)
	stretch.Translate(10, 20)
	if got, want := imageGeomBounds(stretch, 1, 60), image.Rect(10, 20, 150, 80); got != want {
		t.Fatalf("stretch bounds = %v, want %v", got, want)
	}

	// 90-degree rotation around origin moves (0..10, 0..20) to (-20..0, 0..10).
	var rot ebiten.GeoM
	rot.Rotate(math.Pi / 2)
	rot.Translate(30, 5)
	if got, want := imageGeomBounds(rot, 10, 20), image.Rect(10, 5, 30, 15); got != want {
		t.Fatalf("rotated bounds = %v, want %v", got, want)
	}

	if got := imageGeomBounds(stretch, 0, 10); !got.Empty() {
		t.Fatalf("empty source bounds = %v, want empty", got)
	}
}

func TestBackgroundBlendModeNotChunked(t *testing.T) {
	t.Parallel()

	src := `<body style="margin:0"><div style="width:140px;height:60px;` +
		`background-color:#f1c40f;background-image:linear-gradient(#c0392b,#2980b9);` +
		`background-blend-mode:multiply"></div></body>`
	d, err := render.DisplayList(context.Background(), src, 200, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !render.Replayable(d) {
		t.Fatal("background-blend-mode:multiply display should be replayable")
	}

	p := buildChunks(d, buildPlan(d))
	foundDirect := false
	for _, it := range p.items {
		if it.run != nil {
			for _, op := range it.run.ops {
				if op.BlendModeName() == "multiply" {
					t.Fatal("blended background image was folded into a chunk run")
				}
			}
		}
		if it.op != nil && it.op.Kind == layout.DisplayOpImage && it.op.BlendModeName() == "multiply" {
			foundDirect = true
		}
	}
	if !foundDirect {
		t.Fatal("expected direct DisplayOpImage with blend mode multiply in chunk plan")
	}
}
