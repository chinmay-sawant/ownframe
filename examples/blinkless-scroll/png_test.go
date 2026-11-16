package main

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestPNGChangesWithScroll(t *testing.T) {
	html, err := os.ReadFile("testdata/page.html")
	if err != nil {
		t.Fatal(err)
	}

	render := func(scroll int) []byte {
		b, err := renderPNG(context.Background(), string(html), 400, 300, scroll)
		if err != nil {
			t.Fatal(err)
		}

		return b
	}

	a, b, a2 := render(0), render(900), render(0)
	if bytes.Equal(a, b) {
		t.Fatal("scroll 0 and 900 must give different pictures")
	}

	if !bytes.Equal(a, a2) {
		t.Fatal("same offset must give the same picture")
	}
}

// TestRealFixtureTriggerPaints checks the trigger box really turns green in
// the picture when its view timeline is in range, and stays beige when not.
func TestRealFixtureTriggerPaints(t *testing.T) {
	html := fixture(t)
	ctx := context.Background()

	s, err := newSession(ctx, html, 880, 700, nil)
	if err != nil {
		t.Fatal(err)
	}

	i, _ := s.host.IndexOf("trigger-box")
	r := s.host.Nodes[i].Rect

	pixel := func(scroll int) color.Color {
		b, err := renderPNG(ctx, html, 880, 700, scroll)
		if err != nil {
			t.Fatal(err)
		}

		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}

		return img.At(int(r.X)+20, int(r.Y)+60)
	}

	on, off := pixel(int(r.Y)-300), pixel(0)
	if on == off {
		t.Fatalf("trigger-box paints the same on and off: %v", on)
	}

	if cr, cg, cb, _ := on.RGBA(); cr>>8 != 0xbb || cg>>8 != 0xf7 || cb>>8 != 0xd0 {
		t.Fatalf("on colour %v, want #bbf7d0", on)
	}
}
