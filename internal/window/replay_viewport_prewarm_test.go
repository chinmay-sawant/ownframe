package window

import (
	"image"
	"testing"
)

func TestViewportWarmupEventuallyCoversThePage(t *testing.T) {
	var warmup viewportWarmup
	content, view := image.Pt(2400, 4000), image.Pt(1000, 600)
	whole := image.Rectangle{Max: content}
	var pixels int
	for {
		r := warmup.take(content, view, image.Point{}, 1)
		if r.Empty() {
			break
		}
		if !r.In(whole) || r.Dx() > view.X || r.Dy() > prewarmBandHeight {
			t.Fatalf("warmup exceeded the per-frame band: %v", r)
		}
		pixels += r.Dx() * r.Dy()
		if pixels > content.X*content.Y {
			t.Fatal("warmup repeated content instead of finishing")
		}
	}
	if pixels != content.X*content.Y {
		t.Fatalf("warmed %d pixels, want the whole page", pixels)
	}
	if r := warmup.take(content, view, image.Point{}, 1); !r.Empty() {
		t.Fatalf("completed warmup restarted: %v", r)
	}
	if r := warmup.take(content, view, image.Point{}, 2); r.Min != (image.Point{}) || r.Empty() {
		t.Fatalf("new content did not restart warmup: %v", r)
	}
}

func TestViewportWarmupFollowsLargePageScroll(t *testing.T) {
	var warmup viewportWarmup
	content, view := image.Pt(2000, 100000), image.Pt(1000, 600)
	warmup.take(content, view, image.Point{}, 1)
	if warmup.area.Dx()*warmup.area.Dy() > prewarmMaxPixels {
		t.Fatal("large-page warmup exceeds its pixel budget")
	}
	scroll := image.Pt(0, 90000)
	r := warmup.take(content, view, scroll, 1)
	wanted := image.Rectangle{Min: scroll, Max: scroll.Add(view)}
	if r.Empty() || !wanted.In(warmup.area) {
		t.Fatalf("warmup did not follow the viewport: band=%v area=%v", r, warmup.area)
	}
}
