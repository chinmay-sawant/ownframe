package snip

import (
	"image/color"
	"testing"
)

func rgbaAt(img interface{ At(int, int) color.Color }, x, y int) color.RGBA {
	return color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
}

func TestAddClearAnnotations(t *testing.T) {
	var anns []Annotation
	anns = AddAnnotation(anns, Annotation{Kind: "rect", X: 1, Y: 1})
	if len(anns) != 1 {
		t.Fatalf("Add len = %d, want 1", len(anns))
	}
	anns = AddAnnotation(anns, Annotation{Kind: "arrow"})
	if len(anns) != 2 {
		t.Fatalf("Add len = %d, want 2", len(anns))
	}
	if anns = ClearAnnotations(anns); len(anns) != 0 {
		t.Fatalf("Clear len = %d, want 0", len(anns))
	}
}

func TestDrawAnnotations(t *testing.T) {
	anns := []Annotation{
		{Kind: "rect", X: 1, Y: 1, W: 4, H: 4, Color: "#ff0000"},
		{Kind: "arrow", X: 0, Y: 0, W: 7, H: 5, Color: "#00ff00"},
		{Kind: "text", X: 5, Y: 0, W: 3, H: 3, Label: "hi", Color: "blue"},
	}
	out := DrawAnnotations(testImage(), anns)
	if c := rgbaAt(out, 1, 3); c != (color.RGBA{255, 0, 0, 255}) {
		t.Fatalf("rect edge = %v, want red", c)
	}
	if c := rgbaAt(out, 7, 5); c != (color.RGBA{0, 255, 0, 255}) {
		t.Fatalf("arrow tip = %v, want green", c)
	}
	if c := rgbaAt(out, 5, 0); c != (color.RGBA{0, 0, 255, 255}) {
		t.Fatalf("text corner = %v, want blue", c)
	}
	plain := DrawAnnotations(testImage(), nil)
	if c := rgbaAt(plain, 1, 3); c != rgbaAt(testImage(), 1, 3) {
		t.Fatal("no annotations must leave pixels untouched")
	}
}
