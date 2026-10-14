package displaypatch

import (
	"image"
	"testing"
)

func TestTileDamageTouchesAndCoalesces(t *testing.T) {
	td := TileDamage{}
	td.Merge(image.Rect(0, 0, 10, 10))
	td.Merge(image.Rect(10, 0, 20, 10))
	td.Merge(image.Rect(0, 10, 20, 20))
	if len(td.Rects) != 1 || td.Rects[0] != image.Rect(0, 0, 20, 20) {
		t.Fatalf("rects %v", td.Rects)
	}
}

func TestTileDamageIgnoresEmptyAndKeepsDisjoint(t *testing.T) {
	td := TileDamage{}
	td.Merge(image.Rectangle{})
	td.Merge(image.Rect(0, 0, 5, 5))
	td.Merge(image.Rect(100, 100, 105, 105))
	if len(td.Rects) != 2 {
		t.Fatalf("rects %v", td.Rects)
	}
}
