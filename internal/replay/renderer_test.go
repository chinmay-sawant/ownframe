package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestRenderersEvictAndCloseIndependently(t *testing.T) {
	var first, second Renderer
	defer first.Close()
	defer second.Close()
	target := imageOp(t, tinyPNG(t, 200), 2, 2, false)
	firstImage := first.decodedImage(&target)
	secondImage := second.decodedImage(&target)
	if firstImage == nil || secondImage == nil || firstImage == secondImage {
		t.Fatal("renderers did not retain independent images")
	}
	for i := range imageCacheLimit {
		filler := imageOp(t, tinyPNG(t, uint8(i)), 2, 2, false)
		if first.decodedImage(&filler) == nil {
			t.Fatal("filler did not decode")
		}
	}
	if first.images.get(imageBytesKey(&target)) != nil {
		t.Fatal("first renderer did not evict its oldest entry")
	}
	if second.decodedImage(&target) != secondImage {
		t.Fatal("first renderer evicted the second renderer's image")
	}

	display := &layout.Display{Width: 16, Height: 16}
	first.chunkPlanFor(display)
	second.chunkPlanFor(display)
	if first.plans.entries[0].root == second.plans.entries[0].root {
		t.Fatal("renderers shared a mutable paint plan")
	}
	box := image.Rect(0, 0, 2, 2)
	key := chunkKey{id: 1}
	first.chunks.get(key, 1, box, func(*ebiten.Image) {})
	retained := second.chunks.get(key, 1, box, func(*ebiten.Image) {})
	first.Close()
	first.Close()
	if len(first.images.entries) != 0 || len(first.chunks.entries) != 0 || first.plans.entries[0].root != nil {
		t.Fatal("Close retained page images or plans")
	}
	if second.decodedImage(&target) != secondImage || second.chunks.entries[key] != retained {
		t.Fatal("closing one renderer cleared another renderer")
	}
	if first.decodedImage(&target) == nil {
		t.Fatal("closed renderer could not be reused")
	}
}
