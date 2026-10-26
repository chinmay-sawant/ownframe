package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	prewarmMaxPixels  = 12 << 20
	prewarmBandHeight = 128
)

// viewportWarmup spreads cold raster work over frames. Large documents warm
// a bounded neighbourhood and start another sweep when scrolling leaves it.
type viewportWarmup struct {
	generation uint64
	content    image.Point
	view       image.Point
	area       image.Rectangle
	next       image.Point
	scratch    *ebiten.Image
}

func (w *viewportWarmup) take(content, view, scroll image.Point, generation uint64) image.Rectangle {
	if content.X <= 0 || content.Y <= view.Y || view.X <= 0 || view.Y <= 0 {
		return image.Rectangle{}
	}
	reset := w.content != content || w.view != view || w.generation != generation
	if reset || scroll.Y < w.area.Min.Y || scroll.Y+view.Y > w.area.Max.Y {
		top := 0
		height := content.Y
		if int64(content.X)*int64(content.Y) > prewarmMaxPixels {
			top = max(0, scroll.Y-view.Y)
			height = max(view.Y, prewarmMaxPixels/content.X)
		}
		w.content, w.view, w.generation = content, view, generation
		w.area = image.Rect(0, top, content.X, min(content.Y, top+height))
		w.next = w.area.Min
	}
	if w.next.Y >= w.area.Max.Y {
		return image.Rectangle{}
	}
	r := image.Rect(
		w.next.X,
		w.next.Y,
		min(w.next.X+view.X, content.X),
		min(w.next.Y+prewarmBandHeight, w.area.Max.Y),
	)
	w.next.X += view.X
	if w.next.X >= content.X {
		w.next.X = 0
		w.next.Y += prewarmBandHeight
	}
	return r
}

// prewarmViewport submits at most one narrow band per frame. Keeping the
// scratch image small also bounds the graphics queue's cold-frame growth.
func (s *shell) prewarmViewport(d *layout.Display, bounds image.Rectangle) bool {
	if d == nil {
		return false
	}
	contentW, contentH := s.contentSize()
	w := &s.viewport.warmup
	r := w.take(
		image.Pt(contentW, contentH),
		bounds.Size(),
		image.Pt(s.scrollX, s.scrollY),
		s.layerGeneration(),
	)
	if r.Empty() {
		if w.scratch != nil {
			w.scratch.Deallocate()
			w.scratch = nil
		}
		return false
	}
	size := image.Pt(min(contentW, bounds.Dx()), prewarmBandHeight)
	if w.scratch == nil || w.scratch.Bounds().Size() != size {
		if w.scratch != nil {
			w.scratch.Deallocate()
		}
		// A transient render target: keep it out of the sprite atlas.
		w.scratch = ebiten.NewImageWithOptions(
			image.Rectangle{Max: size}, &ebiten.NewImageOptions{Unmanaged: true},
		)
	}
	clip, ok := w.scratch.SubImage(image.Rectangle{Max: r.Size()}).(*ebiten.Image)
	if !ok {
		return false
	}
	clip.Clear()
	clip.Fill(s.pageBackground())
	s.renderer.DrawVisibleUnfixedZBackground(clip, d, -float64(r.Min.X), -float64(r.Min.Y), s.viewportPinZ())
	return true
}
