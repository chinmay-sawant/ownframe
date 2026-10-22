package page_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

const backgroundHTML = `<body style="margin:0">` +
	`<div style="width:40px;height:40px;background-image:url('bg');` +
	`background-size:100% 100%"></div></body>`

// redPNG builds a 2x2 opaque red PNG.
func redPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
		img.Pix[i+3] = 255
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// centerPixel decodes a PNG and returns its center pixel.
func centerPixel(t *testing.T, data []byte) color.NRGBA {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	bounds := img.Bounds()

	c, ok := color.NRGBAModel.Convert(img.At(bounds.Dx()/2, bounds.Dy()/2)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}

	return c
}

func TestSetImagePaintsAndRemovesBackground(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	screen, err := page.New(page.Config{HTML: backgroundHTML, Width: 40, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	paint := func() color.NRGBA {
		if err := screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		return centerPixel(t, screen.PNGWithContext(ctx))
	}

	screen.SetImage("bg", redPNG(t))
	if got := paint(); got.R < 200 || got.G > 60 || got.B > 60 {
		t.Fatalf("center pixel %+v, want red", got)
	}

	screen.SetImage("bg", nil)
	if got := paint(); got.R < 200 || got.G < 200 || got.B < 200 {
		t.Fatalf("center pixel %+v after removal, want white", got)
	}
}
