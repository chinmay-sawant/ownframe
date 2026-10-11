package replay

import (
	"context"
	"image"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// BenchmarkDrawRectScan times the op filter on the platform example, once
// for a counter-sized dirty rect and once for the whole frame.
func BenchmarkDrawRectScan(b *testing.B) {
	html, err := os.ReadFile("../../examples/platform/platform/platform.html")
	if err != nil {
		b.Fatal(err)
	}

	display, err := render.DisplayListState(context.Background(), string(html), 480, 640, render.State{})
	if err != nil {
		b.Fatal(err)
	}

	box, found := boxByID(display.Boxes, "count")
	if !found {
		b.Fatal("no #count box")
	}

	ppt := display.PointsPerPixel
	dirty := image.Rect(int(box.X)-8, int(box.Y)-8, int(box.X+box.W)+8, int(box.Y+box.H)+8)
	full := image.Rect(0, 0, display.Width, display.Height)

	b.Logf("ops=%d dirty=%v hits=%d", len(display.Ops), dirty, countTouches(display, ppt, dirty))

	b.Run("dirty", func(b *testing.B) { scanOps(b, display, ppt, dirty) })
	b.Run("full", func(b *testing.B) { scanOps(b, display, ppt, full) })
}

func countTouches(display *layout.Display, ppt float64, rect image.Rectangle) int {
	n := 0

	for i := range display.Ops {
		if opTouches(&display.Ops[i], ppt, rect) {
			n++
		}
	}

	return n
}

func scanOps(b *testing.B, display *layout.Display, ppt float64, rect image.Rectangle) {
	for i := 0; i < b.N; i++ {
		for j := range display.Ops {
			opTouches(&display.Ops[j], ppt, rect)
		}
	}
}

func boxByID(boxes []layout.Box, id string) (layout.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return layout.Box{}, false
}
