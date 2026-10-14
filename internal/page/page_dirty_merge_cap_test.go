package page_test

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// coveredBy reports whether one of the rects contains b.
func coveredBy(rects []image.Rectangle, b image.Rectangle) bool {
	for _, r := range rects {
		if b.In(r) {
			return true
		}
	}

	return false
}

// Twelve dots past the merge gap in a small area must merge under the
// window's eight-rect cap instead of overflowing it.
func TestDiffRectListUnderWindowCap(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{range .}}<div style="position:absolute;left:{{.X}}px;top:{{.Y}}px;width:2px;height:2px;background:{{.C}}"></div>{{end}}</body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	type dot struct {
		X, Y int
		C    string
	}

	dots := make([]dot, 0, 12)
	for y := 0; y < 60; y += 20 {
		for x := 0; x < 60; x += 20 {
			dots = append(dots, dot{x, y, "#eee"})
		}
	}

	screen.SetData(dots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	for i := range dots {
		dots[i].C = "#111"
	}

	screen.SetData(dots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || len(rects) > 8 {
		t.Fatalf("rects %d ok %v over the window cap", len(rects), ok)
	}

	if isFullFrame(rects, mergeW, mergeH) {
		t.Fatal("small dots fell back to the frame")
	}

	for _, d := range dots {
		b := image.Rect(d.X, d.Y, d.X+2, d.Y+2)
		if !coveredBy(rects, b) {
			t.Fatalf("rects %v miss dot %v", rects, b)
		}
	}
}
