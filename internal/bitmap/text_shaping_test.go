package bitmap_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func shapedInk(t *testing.T, style, content string) image.Rectangle {
	t.Helper()
	d, err := render.DisplayList(context.Background(),
		`<body style="margin:0;background:#fff"><div style="position:absolute;left:10px;top:10px;color:#000;font-size:18px;`+
			style+`">`+content+`</div></body>`, 300, 60)
	if err != nil {
		t.Fatal(err)
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatalf("%q: software painter: %v", style, err)
	}
	var ink image.Rectangle
	for y := 0; y < 60; y++ {
		for x := 0; x < 300; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 0xfa00 || g < 0xfa00 || b < 0xfa00 {
				ink = ink.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}

	return ink
}

// TestSoftwarePainterAcceptsFontFeatures: the painter has no shaper, so it
// draws the run without its features instead of rejecting the page.
func TestSoftwarePainterAcceptsFontFeatures(t *testing.T) {
	t.Parallel()
	for _, style := range []string{
		"font-variant-numeric:tabular-nums",
		"font-variant-ligatures:none",
		"font-kerning:none",
	} {
		plain := shapedInk(t, "", "Hello 123 fi")
		got := shapedInk(t, style, "Hello 123 fi")
		if got.Empty() {
			t.Fatalf("%s: no ink", style)
		}
		// These features leave the default glyph forms and advances alone.
		if got != plain {
			t.Fatalf("%s: ink %v differs from the unfeatured run %v", style, got, plain)
		}
	}
	// small-caps changes the text itself (the engine uppercases it), so it
	// only has to paint.
	if shapedInk(t, "font-variant:small-caps", "Hello").Empty() {
		t.Fatal("small-caps painted no ink")
	}
}

// TestSoftwarePainterReproducesAutospaceGap: layout opens 0.125 em (about 2 px
// for 18 px text) at each ideograph and Latin boundary, and the painter must
// advance by the same amount.
func TestSoftwarePainterReproducesAutospaceGap(t *testing.T) {
	t.Parallel()
	const content = "日本語abc"
	plain := shapedInk(t, "text-autospace:no-autospace", content)
	spaced := shapedInk(t, "text-autospace:normal", content)
	if plain.Empty() || spaced.Empty() {
		t.Fatal("no ink")
	}
	if grow := spaced.Dx() - plain.Dx(); grow < 1 || grow > 5 {
		t.Fatalf("autospace grew the run by %d px (plain %v, spaced %v), want about one 0.125 em gap", grow, plain, spaced)
	}
	// With no boundary there is nothing to open.
	if a, b := shapedInk(t, "text-autospace:no-autospace", "abc"), shapedInk(t, "text-autospace:normal", "abc"); a != b {
		t.Fatalf("latin-only run changed under autospace: %v vs %v", a, b)
	}
}

// TestSoftwarePainterExpandsTabsWithoutTofu: literal tab characters under
// white-space:pre must advance by tab-size without drawing missing-glyph boxes
// in the gap between columns.
func TestSoftwarePainterExpandsTabsWithoutTofu(t *testing.T) {
	t.Parallel()
	const html = `<body style="margin:0;background:#fff"><div style="position:absolute;left:10px;top:10px;color:#000;font-size:18px;white-space:pre;tab-size:8">A` + "\t" + `B</div></body>`
	d, err := render.DisplayList(context.Background(), html, 300, 60)
	if err != nil {
		t.Fatal(err)
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatalf("software painter: %v", err)
	}
	// Find horizontal spans of ink. A and B must be separated by a clean blank
	// gap (no missing-glyph tofu box between them).
	var spans [][2]int
	inSpan := false
	start := 0
	for x := 0; x < 300; x++ {
		hasInk := false
		for y := 0; y < 60; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 0xfa00 || g < 0xfa00 || b < 0xfa00 {
				hasInk = true
				break
			}
		}
		if hasInk && !inSpan {
			start = x
			inSpan = true
		} else if !hasInk && inSpan {
			spans = append(spans, [2]int{start, x})
			inSpan = false
		}
	}
	if inSpan {
		spans = append(spans, [2]int{start, 300})
	}
	if len(spans) != 2 {
		t.Fatalf("expected 2 ink spans (A and B) with clean gap, got %d spans: %v", len(spans), spans)
	}
	if gap := spans[1][0] - spans[0][1]; gap < 20 {
		t.Fatalf("expected tab gap >= 20px for tab-size:8, got %dpx (spans %v)", gap, spans)
	}
}
