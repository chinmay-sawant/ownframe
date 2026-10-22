package main

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"testing"
)

func TestMaskExamplePixelsAndRetainedDisplay(t *testing.T) {
	p, err := newPage()
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.Display() == nil || p.Image() != nil {
		t.Fatal("masked page did not keep its display list")
	}
	img, err := png.Decode(bytes.NewReader(p.PNG()))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name                   string
		x, y                   int
		referenceX, referenceY int
	}{
		{"alpha transparent half", 50, 110, 190, 42},
		{"alpha opaque half", 150, 110, 130, 42},
		{"luminance black half", 310, 110, 30, 42},
		{"luminance white half", 410, 110, 290, 42},
		{"composite band", 95, 280, 75, 50},
		{"composite outside band", 45, 280, 185, 50},
		{"SVG luminance black half", 310, 280, 30, 50},
		{"SVG luminance white half", 410, 280, 290, 50},
	}
	reference := imageReference(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := color.NRGBAModel.Convert(img.At(tc.x, tc.y)).(color.NRGBA)
			if !ok {
				t.Fatalf("pixel model changed at (%d,%d)", tc.x, tc.y)
			}
			want, ok := color.NRGBAModel.Convert(reference.At(tc.referenceX, tc.referenceY)).(color.NRGBA)
			if !ok {
				t.Fatalf("pixel model changed at (%d,%d)", tc.referenceX, tc.referenceY)
			}
			if got != want {
				t.Fatalf("pixel (%d,%d) = %+v, want image pixel %+v", tc.x, tc.y, got, want)
			}
		})
	}
}
