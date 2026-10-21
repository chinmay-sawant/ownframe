package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

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
