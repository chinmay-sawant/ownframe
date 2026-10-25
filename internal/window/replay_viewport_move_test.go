package window

import (
	"image"
	"testing"
)

func TestViewportExposedCoversOnlyNewPixels(t *testing.T) {
	previous := image.Rect(0, 0, 100, 200)
	for _, next := range []image.Rectangle{
		image.Rect(0, 80, 100, 280), image.Rect(20, 80, 120, 280),
		image.Rect(-20, -80, 80, 120), image.Rect(0, 1000, 100, 1200), previous,
	} {
		strips := viewportExposed(next, previous)
		for y := next.Min.Y; y < next.Max.Y; y++ {
			for x := next.Min.X; x < next.Max.X; x++ {
				point := image.Pt(x, y)
				covers := 0
				for _, strip := range strips {
					if point.In(strip) {
						covers++
					}
				}
				want := 1
				if point.In(previous) {
					want = 0
				}
				if covers != want {
					t.Fatalf("%v: pixel %v replayed %d times, want %d", next, point, covers, want)
				}
			}
		}
	}
}
