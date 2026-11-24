package masking

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func solidMaskPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode solid png: %v", err)
	}
	return buf.Bytes()
}

func TestValidateAcceptsPaddingAndContentBox(t *testing.T) {
	t.Parallel()

	pngBytes := solidMaskPNG(t, 4, 4)
	for _, box := range []string{"", "border-box", "padding-box", "content-box"} {
		desc := &layout.MaskDescriptor{
			Layers: []layout.MaskLayer{{
				Source:     "url(m.png)",
				ImageBytes: pngBytes,
				Size:       "auto",
				Position:   "0 0",
				Repeat:     "no-repeat",
				Origin:     box,
				Clip:       box,
			}},
		}
		if err := Validate(desc); err != nil {
			t.Fatalf("Validate(origin/clip=%q) unexpected error: %v", box, err)
		}
	}

	bad := &layout.MaskDescriptor{
		Layers: []layout.MaskLayer{{
			Source:     "url(m.png)",
			ImageBytes: pngBytes,
			Size:       "auto",
			Position:   "0 0",
			Origin:     "fill-box",
			Clip:       "border-box",
		}},
	}
	if err := Validate(bad); err == nil {
		t.Fatal("Validate(origin=fill-box) expected error")
	}
}

func TestMaskOriginAndClipGeometry(t *testing.T) {
	t.Parallel()

	// 40x40 border-box, 4px border (3pt), 6px padding (4.5pt).
	// padding-box: [4, 36) x [4, 36) (32x32)
	// content-box: [10, 30) x [10, 30) (20x20)
	bounds := image.Rect(0, 0, 40, 40)
	borderPt := [4]float64{3, 3, 3, 3}
	paddingPt := [4]float64{4.5, 4.5, 4.5, 4.5}
	pngBytes := solidMaskPNG(t, 8, 8)

	t.Run("clip padding-box", func(t *testing.T) {
		t.Parallel()
		desc := &layout.MaskDescriptor{
			BorderInsets:  borderPt,
			PaddingInsets: paddingPt,
			Layers: []layout.MaskLayer{{
				Source:     "url(m.png)",
				ImageBytes: pngBytes,
				Size:       "100% 100%",
				Position:   "0 0",
				Repeat:     "no-repeat",
				Origin:     "border-box",
				Clip:       "padding-box",
			}},
		}
		img, err := Image(bounds, desc)
		if err != nil {
			t.Fatalf("Image: %v", err)
		}
		if got := img.NRGBAAt(2, 2).A; got != 0 {
			t.Fatalf("border pixel alpha = %d, want 0", got)
		}
		if got := img.NRGBAAt(5, 5).A; got != 255 {
			t.Fatalf("padding pixel alpha = %d, want 255", got)
		}
		if got := img.NRGBAAt(20, 20).A; got != 255 {
			t.Fatalf("content pixel alpha = %d, want 255", got)
		}
	})

	t.Run("clip content-box", func(t *testing.T) {
		t.Parallel()
		desc := &layout.MaskDescriptor{
			BorderInsets:  borderPt,
			PaddingInsets: paddingPt,
			Layers: []layout.MaskLayer{{
				Source:     "url(m.png)",
				ImageBytes: pngBytes,
				Size:       "100% 100%",
				Position:   "0 0",
				Repeat:     "no-repeat",
				Origin:     "border-box",
				Clip:       "content-box",
			}},
		}
		img, err := Image(bounds, desc)
		if err != nil {
			t.Fatalf("Image: %v", err)
		}
		if got := img.NRGBAAt(6, 6).A; got != 0 {
			t.Fatalf("padding pixel alpha = %d, want 0", got)
		}
		if got := img.NRGBAAt(12, 12).A; got != 255 {
			t.Fatalf("content pixel alpha = %d, want 255", got)
		}
	})

	t.Run("origin padding-box positions no-repeat mask inside border", func(t *testing.T) {
		t.Parallel()
		desc := &layout.MaskDescriptor{
			BorderInsets:  borderPt,
			PaddingInsets: paddingPt,
			Layers: []layout.MaskLayer{{
				Source:     "url(m.png)",
				ImageBytes: pngBytes,
				Size:       "8px 8px",
				Position:   "0 0",
				Repeat:     "no-repeat",
				Origin:     "padding-box",
				Clip:       "border-box",
			}},
		}
		img, err := Image(bounds, desc)
		if err != nil {
			t.Fatalf("Image: %v", err)
		}
		if got := img.NRGBAAt(2, 2).A; got != 0 {
			t.Fatalf("before padding-box origin alpha = %d, want 0", got)
		}
		if got := img.NRGBAAt(4, 4).A; got != 255 {
			t.Fatalf("at padding-box origin alpha = %d, want 255", got)
		}
		if got := img.NRGBAAt(11, 11).A; got != 255 {
			t.Fatalf("inside 8x8 mask from padding-box origin alpha = %d, want 255", got)
		}
		if got := img.NRGBAAt(13, 13).A; got != 0 {
			t.Fatalf("outside 8x8 mask alpha = %d, want 0", got)
		}
	})
}
