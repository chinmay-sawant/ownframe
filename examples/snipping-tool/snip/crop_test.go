package snip

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func testImage() image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			m.Set(x, y, color.RGBA{uint8(x * 30), uint8(y * 40), 0, 255})
		}
	}
	return m
}

func TestCrop(t *testing.T) {
	got := Crop(testImage(), Rect{2, 1, 3, 2})
	if got.Bounds().Dx() != 3 || got.Bounds().Dy() != 2 {
		t.Fatalf("Crop size = %v, want 3x2", got.Bounds())
	}
	want := color.RGBAModel.Convert(testImage().At(2, 1))
	if c := color.RGBAModel.Convert(got.At(0, 0)); c != want {
		t.Fatalf("Crop pixel = %v, want %v", c, want)
	}
	full := Crop(testImage(), Rect{-5, -5, 99, 99})
	if full.Bounds().Dx() != 8 || full.Bounds().Dy() != 6 {
		t.Fatalf("Crop clamped size = %v, want 8x6", full.Bounds())
	}
	empty := Crop(testImage(), Rect{2, 2, 0, 0})
	if empty.Bounds().Dx() != 8 || empty.Bounds().Dy() != 6 {
		t.Fatalf("Crop empty size = %v, want copy of src", empty.Bounds())
	}
}

func TestCropPNG(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImage()); err != nil {
		t.Fatal(err)
	}
	out, err := CropPNG(buf.Bytes(), Rect{1, 1, 4, 3})
	if err != nil {
		t.Fatal(err)
	}
	dec, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if dec.Bounds().Dx() != 4 || dec.Bounds().Dy() != 3 {
		t.Fatalf("CropPNG size = %v, want 4x3", dec.Bounds())
	}
	if _, err := CropPNG([]byte("not a png"), Rect{}); err == nil {
		t.Fatal("CropPNG must fail on bad input")
	}
}
