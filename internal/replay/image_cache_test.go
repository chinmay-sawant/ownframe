package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
	if len(c.order) != 2 {
		t.Fatalf("cache holds %d entries, want 2 after pixel eviction", len(c.order))
	}
}
