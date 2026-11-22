package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type viewportKey struct {
	generation uint64
	bounds     image.Rectangle
}

type viewportState struct {
	buf      *ebiten.Image
	spare    *ebiten.Image
	key      viewportKey
	area     image.Rectangle
	scroll   image.Point
	valid    bool
	blitOnly bool
	warmup   viewportWarmup
}

func (s *shell) drawViewport(dst *ebiten.Image, d *layout.Display, taker dirtyTaker) {
	s.dropContentBuffers()
	bounds := dst.Bounds()
	if bounds.Empty() {
		s.viewport.blitOnly = false
		return
	}
	// setViewportArea may dispose the buffer and zero this state.
	had := s.viewport.valid && s.viewport.buf != nil
	prevScroll := s.viewport.scroll
	scroll := image.Pt(s.scrollX, s.scrollY)
	key := viewportKey{generation: s.seq, bounds: bounds}
	// Bake this frame's retained layers before reading dirty state. The
	// background below leaves their operations as holes, and the layers are
	// composited after the blit, so a moving layer never repaints the page
	// under it.
	specs := s.layerSpecs(d)
	s.renderer.BeginViewportLayers(d, s.layerGeneration(), specs, s.viewportPinZ())
	s.syncPromotedLayers()
	// Paint dirty rects separately so a moving chip costs its own box instead
	// of the bounding box of every moving chip.
	rects, changed := taker.TakeDirtyRects()
	// A promotion opens a hole in the background and a demotion closes one.
	// Ordinary layer motion needs neither, because the pixels under a layer
	// stay background.
	if topology := s.renderer.LayerTopologyDamage(); len(topology) > 0 {
		rects = append(rects, topology...)
		// Overscan can contain an unpromoted pose painted many ticks ago.
		// Current layer bounds cannot locate it after an offscreen animation
		// moves, so a membership change invalidates the cached background.
		s.viewport.valid = false
	}
	changed = changed || len(rects) > 0
	wanted := bounds.Add(scroll)
	full := !s.viewport.valid || s.viewport.key.bounds != bounds || !wanted.In(s.viewport.area)
	full = full || (s.viewport.key.generation != key.generation && !changed)
	if full {
		previous := s.viewport.area
		reuse := s.viewport.valid && s.viewport.key.bounds == bounds &&
			(s.viewport.key.generation == key.generation || changed)
		buffer := s.viewport.buf
		s.setViewportArea(bounds)
		if reuse && s.viewport.buf == buffer && previous.Overlaps(s.viewport.area) {
			s.moveViewport(d, previous)
			if changed {
				s.paintViewportRects(d, rects)
			}
		} else {
			s.paintViewport(d, s.viewport.area, s.viewportPinZ())
		}
	} else if changed && len(rects) > 0 {
		s.paintViewportRects(d, rects)
	}
	if had && scroll != prevScroll {
		// Offscreen motion keeps sampling without page damage. Refresh the
		// strips entering the screen even when the overscan already covers them.
		for _, strip := range viewportExposed(wanted, bounds.Add(prevScroll)) {
			s.paintViewport(d, strip, s.viewportPinZ())
		}
	}
	noPaint := !full && !changed
	s.viewport.blitOnly = had && noPaint && scroll == prevScroll
	s.viewport.scroll = scroll
	s.viewport.key, s.viewport.valid = key, true
	var options ebiten.DrawImageOptions
	options.Blend = ebiten.BlendCopy
	options.GeoM.Translate(
		float64(s.viewport.area.Min.X-s.scrollX),
		float64(s.viewport.area.Min.Y-s.scrollY),
	)
	dst.DrawImage(s.viewport.buf, &options)
	// Layers composite onto dst in page coordinates, not into the buffer, so
	// they take the scroll offset alone. Adding the buffer origin here put
	// every retained layer one overscan band low once the page scrolled.
	if clip := s.pinClip(dst); clip != nil {
		s.renderer.DrawLayers(clip, d, -float64(s.scrollX), -float64(s.scrollY))
		s.renderer.DrawFixed(clip, d)
		s.renderer.DrawPinnedZ(clip, d, s.viewportPinZ(), -float64(s.redrawX), -float64(s.redrawY))
	}
	if s.prewarmViewport(d, bounds) {
		s.viewport.blitOnly = false
	}
}

func (s *shell) disposeViewport() {
	if s.viewport.spare != nil {
		s.viewport.spare.Deallocate()
	}
	if s.viewport.warmup.scratch != nil {
		s.viewport.warmup.scratch.Deallocate()
	}
	if s.viewport.buf != nil {
		s.viewport.buf.Deallocate()
	}
	s.viewport = viewportState{}
}

func (s *shell) dropContentBuffers() {
	if s.partial.tiles != nil {
		s.partial.tiles.dispose()
		s.partial = partialState{}
	}

	if s.scaleBuf != nil {
		s.scaleBuf.Deallocate()
		s.scaleBuf = nil
	}
}

// moveViewport keeps the already painted overlap when scrolling crosses the
// overscan. Only the newly exposed strips need a replay of the display list.
func (s *shell) moveViewport(d *layout.Display, previous image.Rectangle) {
	v := &s.viewport
	if v.spare == nil {
		bounds := v.buf.Bounds()
		v.spare = ebiten.NewImageWithOptions(bounds, &ebiten.NewImageOptions{Unmanaged: true})
	}
	v.buf, v.spare = v.spare, v.buf
	v.buf.Clear()
	var options ebiten.DrawImageOptions
	options.GeoM.Translate(float64(previous.Min.X-v.area.Min.X), float64(previous.Min.Y-v.area.Min.Y))
	v.buf.DrawImage(v.spare, &options)
	for _, strip := range viewportExposed(v.area, previous) {
		s.paintViewport(d, strip, s.viewportPinZ())
	}
}

func viewportExposed(next, previous image.Rectangle) [4]image.Rectangle {
	shared := next.Intersect(previous)
	if shared.Empty() {
		return [4]image.Rectangle{next}
	}
	return [4]image.Rectangle{
		image.Rect(next.Min.X, next.Min.Y, next.Max.X, shared.Min.Y),
		image.Rect(next.Min.X, shared.Max.Y, next.Max.X, next.Max.Y),
		image.Rect(next.Min.X, shared.Min.Y, shared.Min.X, shared.Max.Y),
		image.Rect(shared.Max.X, shared.Min.Y, next.Max.X, shared.Max.Y),
	}
}

const viewportOverscan = 256

// viewportSides counts the top and bottom overscan bands.
const viewportSides = 2

func (s *shell) setViewportArea(bounds image.Rectangle) {
	gap := min(viewportOverscan, bounds.Dy())
	width, height := bounds.Dx(), bounds.Dy()+viewportSides*gap
	if limit := ebiten.MaxImageSize(); limit > 0 {
		height = min(height, limit)
	}
	if s.viewport.buf == nil || s.viewport.buf.Bounds().Size() != image.Pt(width, height) {
		s.disposeViewport()
		// This image alternates between a render target and a scroll source.
		// Keep it out of the sprite atlas so those switches never migrate it.
		s.viewport.buf = ebiten.NewImageWithOptions(
			image.Rect(0, 0, width, height), &ebiten.NewImageOptions{Unmanaged: true},
		)
	}
	top := max(0, s.scrollY-gap)
	s.viewport.area = image.Rect(s.scrollX, top, s.scrollX+width, top+height)
}

// viewportDirtyBounds is the fallback region a viewport repaint covers: the
// bounding box of the list. paintViewportRects uses it only when the list is
// longer than one frame should scan; a short list paints rect by rect.
func viewportDirtyBounds(rects []image.Rectangle) image.Rectangle {
	if len(rects) == 0 {
		return image.Rectangle{}
	}
	box := rects[0]
	for _, r := range rects[1:] {
		box = box.Union(r)
	}
	return box
}

// paintViewportRects repaints each dirty rect in turn. A moving chip then
// costs its own box, not the bounding box of every moving card on screen. A
// list past the frame cap falls back to one bounding-box pass.
func (s *shell) paintViewportRects(d *layout.Display, rects []image.Rectangle) {
	if len(rects) == 0 {
		return
	}
	if len(rects) > maxRepaintRects {
		s.paintViewport(d, viewportDirtyBounds(rects), s.viewportPinZ())

		return
	}
	for _, rect := range rects {
		s.paintViewport(d, rect, s.viewportPinZ())
	}
}

func (s *shell) paintViewport(d *layout.Display, rect image.Rectangle, pinZ int) {
	rect = rect.Intersect(s.viewport.area).Sub(s.viewport.area.Min)
	if rect.Empty() {
		return
	}
	clip, ok := s.viewport.buf.SubImage(rect).(*ebiten.Image)
	if !ok {
		return
	}
	s.partial.paints++
	s.partial.paintPx += rect.Dx() * rect.Dy()
	fillBackground(clip, s.pageBackground())
	s.renderer.DrawVisibleUnfixedZBackground(clip, d, -float64(s.viewport.area.Min.X), -float64(s.viewport.area.Min.Y), pinZ)
}

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

func (w *viewportWarmup) take(content, view, scroll image.Point, generation uint64, windowing bool) image.Rectangle {
	if content.X <= 0 || content.Y <= view.Y || view.X <= 0 || view.Y <= 0 {
		return image.Rectangle{}
	}
	reset := w.content != content || w.view != view || w.generation != generation
	if reset || scroll.Y < w.area.Min.Y || scroll.Y+view.Y > w.area.Max.Y {
		top := 0
		height := content.Y
		if windowing {
			// A windowed page renders only the viewport slice between
			// full-height spacers, so sweep the viewport plus overscan
			// instead of the full spacer height.
			gap := min(viewportOverscan, view.Y)
			top = max(0, scroll.Y-gap)
			height = min(content.Y-top, view.Y+viewportSides*gap)
		} else if int64(content.X)*int64(content.Y) > prewarmMaxPixels {
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
		s.windowing(),
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
	fillBackground(clip, s.pageBackground())
	s.renderer.DrawVisibleUnfixedZBackground(clip, d, -float64(r.Min.X), -float64(r.Min.Y), s.viewportPinZ())
	return true
}
