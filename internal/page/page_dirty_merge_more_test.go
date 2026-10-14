package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Changes scattered across the frame must still fall back to the frame.
func TestDiffScatteredChangesFallBack(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{range .}}<div style="position:absolute;left:{{.X}}px;top:{{.Y}}px;width:20px;height:20px;background:{{.C}}"></div>{{end}}</body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	type spot struct {
		X, Y int
		C    string
	}

	spots := []spot{{0, 0, "#eee"}, {300, 0, "#eee"}, {0, 180, "#eee"}, {300, 180, "#eee"}, {150, 90, "#eee"}}

	screen.SetData(spots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	for i := range spots {
		spots[i].C = "#111"
	}

	screen.SetData(spots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || !isFullFrame(rects, mergeW, mergeH) {
		t.Fatalf("scattered %v ok %v", rects, ok)
	}
}

// One change covering more than a third of the frame must fall back.
func TestDiffThirdAreaFallsBack(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0"><div style="width:320px;height:100px;background:{{.C}}"></div></body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ C string }{C: "#eee"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	screen.SetData(struct{ C string }{C: "#111"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || !isFullFrame(rects, mergeW, mergeH) {
		t.Fatalf("third area %v ok %v", rects, ok)
	}
}
