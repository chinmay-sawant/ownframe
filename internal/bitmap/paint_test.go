package bitmap_test

import (
	"context"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestFallbackPaintsGeometryAndCompositing(t *testing.T) {
	cases := []struct {
		name, style string
		x, y        int
		want        color.NRGBA
	}{
		{name: "ellipse corner", style: "background:red;border-radius:50% / 25%", x: 0, y: 0, want: color.NRGBA{R: 255, G: 255, B: 255, A: 255}},
		{name: "ellipse center", style: "background:red;border-radius:50% / 25%", x: 20, y: 20, want: color.NRGBA{R: 255, A: 255}},
		{name: "border", style: "border:4px solid blue", x: 1, y: 20, want: color.NRGBA{B: 255, A: 255}},
		{name: "opacity", style: "background:red;opacity:.5", x: 20, y: 20, want: color.NRGBA{R: 255, G: 127, B: 127, A: 255}},
		{name: "translation", style: "background:red;transform:translate(20px,0)", x: 30, y: 20, want: color.NRGBA{R: 255, A: 255}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := `<body style="margin:0"><div style="width:40px;height:40px;` + tc.style + `"></div></body>`
			d, err := render.DisplayList(context.Background(), source, 100, 100)
			if err != nil {
				t.Fatal(err)
			}
			img, err := bitmap.Paint(d)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := color.NRGBAModel.Convert(img.At(tc.x, tc.y)).(color.NRGBA)
			if !ok {
				t.Fatal("pixel model changed")
			}
			if got != tc.want {
				t.Fatalf("pixel %+v want %+v", got, tc.want)
			}
		})
	}
}

func TestBitmapWritingModeAndMaskedGroupPaintPixels(t *testing.T) {
	const (
		canvasWidth   = 160
		canvasHeight  = 220
		darkThreshold = 80
		tintDelta     = 10
	)
	source := `<body style="margin:0;background:#ffffff">` +
		`<div style="margin-top:80px;width:120px;height:120px;writing-mode:vertical-rl;color:#000000;font-size:18px;` +
		`background:#dbeafe;border:2px solid #1d4ed8">Vertical</div>` +
		`<div style="width:120px;height:60px;background:#1d4ed8;mask-image:linear-gradient(black, transparent)"></div>` +
		`</body>`
	d, err := render.DisplayList(context.Background(), source, canvasWidth, canvasHeight)
	if err != nil {
		t.Fatal(err)
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}

	darkPixels := 0
	tintedPixels := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c, ok := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if !ok {
				t.Fatal("pixel model changed")
			}
			if c.R < darkThreshold && c.G < darkThreshold && c.B < darkThreshold {
				darkPixels++
			}
			if c.B > c.R+tintDelta {
				tintedPixels++
			}
		}
	}
	if darkPixels == 0 {
		t.Fatal("vertical-rl text painted 0 dark glyph pixels")
	}
	if tintedPixels == 0 {
		t.Fatal("masked vertical-rl box painted 0 tinted background/border pixels")
	}
}
