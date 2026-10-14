package runtimebehavior

import (
	"image"
	"testing"
)

func TestVisualCompareImageEdge(t *testing.T) {
	v := NewVisualComparer()
	var nilImg image.Image
	if !v.CompareImage(nilImg, nilImg) {
		t.Fatal("nil images should match")
	}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if v.CompareImage(img, nilImg) {
		t.Fatal("one nil image should mismatch")
	}
	if !v.Mismatch {
		t.Fatal("mismatch flag wrong")
	}
	same := image.NewRGBA(image.Rect(0, 0, 2, 2))
	other := image.NewRGBA(image.Rect(0, 0, 3, 3))
	if v.CompareImage(same, other) {
		t.Fatal("bounds mismatch should fail")
	}
	if !v.CompareImage(same, same) {
		t.Fatal("identical images should match")
	}
}

func TestVisualCompareAliases(t *testing.T) {
	v := NewVisualComparer()
	if !v.CompareImages([]byte(" same "), []byte(" same ")) {
		t.Fatal("CompareImages broken")
	}
	if !v.Run([]byte("x"), []byte("x")) || !v.RunCompare([]byte("x"), []byte("x")) {
		t.Fatal("run aliases broken")
	}
	v2 := NewVisualComparer()
	if v2.CompareImages(nil, nil) != true || !v2.IsEqual() {
		t.Fatal("nil compare alias broken")
	}
}
