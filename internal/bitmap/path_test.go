package bitmap_test

import (
	"context"
	"image"
	"image/color"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cornerShapePage is the shaped box the replay golden uses: a 120 by 80 CSS
// pixel box at (20, 20), a 4 px border, and a 24 px bevel. The engine paints it
// as two DisplayOpPath ops instead of a radiused fill and stroke pair.
const cornerShapePage = `<body style="margin:0;height:140px;background:#ffffff">` +
	`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// bitmapPath returns the first corner path of a display, by fill or border.
func bitmapPath(t *testing.T, display *layout.Display, stroke bool) *layout.DisplayOp {
	t.Helper()

	for index := range display.Ops {
		op := &display.Ops[index]
		if op.Kind != layout.DisplayOpPath {
			continue
		}

		if corner, ok := op.CornerPathValue(); ok && corner != nil && corner.Stroke == stroke {
			return op
		}
	}

	t.Fatalf("no corner path with stroke=%v", stroke)

	return nil
}

// bitmapPixel reads one painted pixel as a non-premultiplied colour.
func bitmapPixel(t *testing.T, img image.Image, x, y int) color.NRGBA {
	t.Helper()

	got, ok := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	if !ok {
		t.Fatal("pixel model changed")
	}

	return got
}

// TestBitmapPaintsCornerShapePaths is the fallback regression: before the path
// painter, one DisplayOpPath made Paint return an unsupported-operation error,
// so a shaped box had no fallback picture at all. The samples sit away from the
// antialiased edges.
func TestBitmapPaintsCornerShapePaths(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerShapePage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	bitmapPath(t, display, false)
	bitmapPath(t, display, true)

	img, err := bitmap.Paint(display)
	if err != nil {
		t.Fatalf("bitmap paint: %v", err)
	}

	fill := color.NRGBA{R: 0x30, G: 0x60, B: 0xc0, A: 0xff}
	border := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	samples := []struct {
		name string
		x, y int
		want color.NRGBA
	}{
		{"fill centre", 84, 64, fill},
		{"fill under the top border", 84, 26, fill},
		{"left border", 20, 84, border},
		{"bevel stroke", 32, 32, border},
		{"inside the bevel", 36, 36, fill},
		{"outside the bevel", 28, 28, white},
		{"right of the box", 160, 64, white},
	}

	for _, sample := range samples {
		if got := bitmapPixel(t, img, sample.x, sample.y); got != sample.want {
			t.Errorf("%s at (%d,%d) = %+v, want %+v", sample.name, sample.x, sample.y, got, sample.want)
		}
	}
}

// TestBitmapPathHonoursOpacity checks that the shape fades with the element
// opacity, the same field an animation writes.
func TestBitmapPathHonoursOpacity(t *testing.T) {
	t.Parallel()

	faded := `<body style="margin:0;height:140px;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
		`corner-shape:bevel;opacity:.5"></div></body>`

	display, err := render.DisplayList(context.Background(), faded, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	img, err := bitmap.Paint(display)
	if err != nil {
		t.Fatalf("bitmap paint: %v", err)
	}

	// #3060c0 at half alpha over white.
	want := color.NRGBA{R: 151, G: 175, B: 223, A: 0xff}
	if got := bitmapPixel(t, img, 84, 64); got != want {
		t.Errorf("faded fill = %+v, want %+v", got, want)
	}
}

// TestBitmapPathFollowsTheMatrix checks that the shaped outline paints through
// the op matrix: a scale about the box centre reaches a point outside the
// upright box.
func TestBitmapPathFollowsTheMatrix(t *testing.T) {
	t.Parallel()

	scaled := `<body style="margin:0;height:140px;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
		`corner-shape:bevel;transform:scale(2)"></div></body>`

	display, err := render.DisplayList(context.Background(), scaled, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	img, err := bitmap.Paint(display)
	if err != nil {
		t.Fatalf("bitmap paint: %v", err)
	}

	fill := color.NRGBA{R: 0x30, G: 0x60, B: 0xc0, A: 0xff}
	if got := bitmapPixel(t, img, 10, 64); got != fill {
		t.Errorf("scaled fill at (10,64) = %+v, want %+v", got, fill)
	}
}

// TestBitmapPathStrokeScalesWithTheMatrix checks the border width rule: the
// stroke is a round pen in the op's local space, so the matrix narrows it with
// the shape instead of leaving it at the untransformed width.
func TestBitmapPathStrokeScalesWithTheMatrix(t *testing.T) {
	t.Parallel()

	scaled := `<body style="margin:0;height:140px;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
		`corner-shape:bevel;transform:scale(.75)"></div></body>`

	display, err := render.DisplayList(context.Background(), scaled, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	img, err := bitmap.Paint(display)
	if err != nil {
		t.Fatalf("bitmap paint: %v", err)
	}

	fill := color.NRGBA{R: 0x30, G: 0x60, B: 0xc0, A: 0xff}
	border := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}

	// The border box shrinks to x 36..132, centred on (84, 64), so the left
	// border is a 3 px band centred on x = 36.
	if got := bitmapPixel(t, img, 36, 64); got != border {
		t.Errorf("scaled border = %+v, want %+v", got, border)
	}
	if got := bitmapPixel(t, img, 39, 64); got != fill {
		t.Errorf("fill inside the scaled border = %+v, want %+v", got, fill)
	}
}

// TestBitmapRoundBoxKeepsTheRectKinds is the untouched-page regression: a box
// without a corner shape must keep the fill rect and stroke rect pair and paint
// the same picture, so the new path painter only sees shaped boxes.
func TestBitmapRoundBoxKeepsTheRectKinds(t *testing.T) {
	t.Parallel()

	source := `<body style="margin:0;height:100px;background:#ffffff">` +
		`<div style="position:absolute;left:20px;top:20px;width:80px;height:60px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px"></div></body>`

	display, err := render.DisplayList(context.Background(), source, 120, 100)
	if err != nil {
		t.Fatal(err)
	}

	for index := range display.Ops {
		if display.Ops[index].Kind == layout.DisplayOpPath {
			t.Fatalf("a round box emitted a path op at %d", index)
		}
	}

	img, err := bitmap.Paint(display)
	if err != nil {
		t.Fatalf("bitmap paint: %v", err)
	}

	fill := color.NRGBA{R: 0x30, G: 0x60, B: 0xc0, A: 0xff}
	border := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}

	if got := bitmapPixel(t, img, 60, 50); got != fill {
		t.Errorf("round-box fill = %+v, want %+v", got, fill)
	}
	if got := bitmapPixel(t, img, 20, 50); got != border {
		t.Errorf("round-box border = %+v, want %+v", got, border)
	}
}
