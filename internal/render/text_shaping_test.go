package render_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// shapingCases are valid CSS whose text ops carry an OpenType feature list or
// an autospace gap. Each used to make the software painter reject the whole
// page, so a real page with font-variant: small-caps failed to render.
var shapingCases = map[string]string{
	"small-caps":         "font-variant:small-caps",
	"all-small-caps":     "font-variant-caps:all-small-caps",
	"tabular-nums":       "font-variant-numeric:tabular-nums slashed-zero",
	"ligatures-none":     "font-variant-ligatures:none",
	"east-asian":         "font-variant-east-asian:jis78",
	"kerning-none":       "font-kerning:none",
	"position-super":     "font-variant-position:super",
	"autospace-normal":   "text-autospace:normal",
	"autospace-ideo":     "text-autospace:ideograph-alpha",
	"features+autospace": `font-variant:small-caps;text-autospace:normal`,
}

func shapingHTML(style string) string {
	return `<body style="margin:0;background:#fff"><div style="position:absolute;left:10px;top:10px;color:#000;font-size:18px;` +
		style + `">Hello 123 日本語abc</div></body>`
}

func inkBox(img image.Image) image.Rectangle {
	var ink image.Rectangle
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r < 0xfa00 || g < 0xfa00 || bl < 0xfa00 {
				ink = ink.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}

	return ink
}

// TestShapingOptionsRenderInsteadOfFailing requires every shaping option to
// build a display, replay on it, and paint ink through the software painter.
func TestShapingOptionsRenderInsteadOfFailing(t *testing.T) {
	t.Parallel()
	for name, style := range shapingCases {
		d, err := render.DisplayList(context.Background(), shapingHTML(style), 300, 80)
		if err != nil {
			t.Fatalf("%s: DisplayList: %v", name, err)
		}
		shaped := false
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind == layout.DisplayOpText && (op.FontFeatures() != "" || op.TextAutospaceGap() != 0) {
				shaped = true
			}
		}
		if !shaped {
			t.Fatalf("%s: no text op carried a shaping option; the case tests nothing", name)
		}
		if !render.Replayable(d) {
			t.Fatalf("%s: the display should replay natively", name)
		}
		img, err := bitmap.Paint(d)
		if err != nil {
			t.Fatalf("%s: software painter: %v", name, err)
		}
		if inkBox(img).Empty() {
			t.Fatalf("%s: software painter drew no ink", name)
		}
	}
}
