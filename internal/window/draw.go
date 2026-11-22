package window

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
	"image"
	"image/color"
	"math"
	"time"
)

func (s *shell) Draw(screen *ebiten.Image) {
	content, compose := s.titlebarCanvas(screen)
	defer compose()
	screen = content
	// Update syncs before the wheel handler, and that handler can redraw.
	// Paint the display it produced. Replaying the previous one into the new
	// damage leaves the tail of a moved thumb on the retained screen.
	if !s.built() {
		if err := s.syncImage(); err != nil && !errors.Is(err, errNoImage) {
			s.drawSyncErr = err
		}
	}

	s.devFrameTime()
	defer s.perfFrameSample()
	s.frameBegin()
	defer s.frameCommit()

	if s.fallback {
		defer s.drawBadge(screen)
	}

	start := time.Now()
	s.drawContent(screen)
	s.drawScrollbars(screen)
	s.dev.draw = time.Since(start)
	if s.perf {
		s.dev.stages.draw = s.dev.draw
	}

	tail := time.Now()
	if s.dev.on {
		s.drawDevTools(screen)
	}
	s.drawMenu(screen)
	if s.perf {
		s.dev.stages.present = time.Since(tail)
	}
}

// drawContent paints the page and reports whether the whole screen is
// painted, which every branch does: the tiled replay blits its own tiles, so
// the caller owes no buffer copy. Each full branch lays the page background
// down first, because the screen keeps the previous frame.
func (s *shell) drawContent(screen *ebiten.Image) bool {
	if s.viewLocked() {
		s.partial.live = false
		screen.Fill(s.pageBackground())
		s.drawFitted(screen)

		return true
	}

	if s.bitmapView.img != nil {
		s.partial.live = false
		s.dropDisplayBuffers()
		screen.Fill(s.pageBackground())
		screen.DrawImage(s.bitmapView.img, nil)

		return true
	}

	zoom := s.zoom()

	if s.display != nil {
		if s.viewportSticky {
			s.drawStickyContent(screen)
			return true
		}
		if s.stretched() || zoom != 1 {
			s.partial.live = false
			screen.Fill(s.pageBackground())
			s.drawReplayScaled(screen)

			return true
		}

		return s.drawReplayPartial(screen)
	}

	s.partial.live = false
	screen.Fill(s.pageBackground())
	s.drawBitmapPage(screen, zoom)

	return true
}

// drawBitmapPage draws the cached bitmap over the already filled screen: its
// own pixels for a page larger than the window, scaled to the window
// otherwise. A size that has not been laid out yet is scaled until the new
// picture arrives.
func (s *shell) drawBitmapPage(screen *ebiten.Image, zoom float64) {
	// The bitmap frame never reads the display-list buffers.
	s.dropDisplayBuffers()
	img := s.img
	if img == nil {
		return
	}

	bounds := img.Bounds()
	viewW, viewH := s.viewSize()
	if bounds.Dx() == 0 || bounds.Dy() == 0 || viewW == 0 || viewH == 0 {
		return
	}

	var op ebiten.DrawImageOptions
	if !s.stretched() {
		if zoom != 1 {
			op.Filter = ebiten.FilterLinear
			op.GeoM.Scale(zoom, zoom)
		}

		op.GeoM.Translate(-float64(s.scrollX), -float64(s.scrollY))
		screen.DrawImage(img, &op)

		return
	}

	op.Filter = ebiten.FilterLinear
	op.GeoM.Scale(
		float64(viewW)/float64(bounds.Dx())*zoom,
		float64(viewH)/float64(bounds.Dy())*zoom,
	)
	screen.DrawImage(img, &op)
}

// channelMax mirrors internal/masking: the max 8-bit color channel value.
const (
	channelMax    = 255
	opaqueAlpha16 = 0xffff
)

// fillBackground clears dst only when bg is translucent and fills it with bg,
// avoiding a redundant Clear + Fill pair on opaque page backgrounds.
func fillBackground(dst *ebiten.Image, bg color.Color) {
	if _, _, _, a := bg.RGBA(); a < opaqueAlpha16 {
		dst.Clear()
	}
	dst.Fill(bg)
}

// pageBackground returns the color to paint under the page: the page's own
// background fill for a display list, or the corner pixel of the bitmap.
// White when neither is known.
func (s *shell) pageBackground() color.Color {
	if s.transparent {
		return color.Transparent
	}

	if s.display != nil {
		if c, ok := displayBackground(s.display); ok {
			return c
		}
	}

	if s.img != nil {
		b := s.img.Bounds()
		if !b.Empty() {
			return s.img.At(b.Min.X, b.Min.Y)
		}
	}

	return color.White
}

// displayBackground finds the first fill in paint order that starts at the
// canvas top-left and spans its width: the html or body background.
func displayBackground(display *layout.Display) (color.Color, bool) {
	if display.PointsPerPixel <= 0 || display.Width <= 0 {
		return nil, false
	}

	widthPts := float64(display.Width) * display.PointsPerPixel
	order := display.Order
	n := len(order)
	if n == 0 {
		n = len(display.Ops)
	}

	for k := 0; k < n; k++ {
		index := k
		if len(order) != 0 {
			index = order[k]
		}

		if index < 0 || index >= len(display.Ops) {
			continue
		}

		op := &display.Ops[index]
		if op.Kind != layout.DisplayOpFillRect || op.X > 0.5 || op.Y > 0.5 {
			continue
		}

		if op.X+op.W < widthPts-0.5 {
			continue
		}

		return opColor(op), true
	}

	return nil, false
}

// opColor returns the op's color composited over white.
func opColor(op *layout.DisplayOp) color.Color {
	a := op.Alpha
	if a <= 0 || a > 1 {
		a = 1
	}

	return color.RGBA{
		R: uint8(clamp01(op.R*a+1-a) * channelMax),
		G: uint8(clamp01(op.G*a+1-a) * channelMax),
		B: uint8(clamp01(op.B*a+1-a) * channelMax),
		A: channelMax,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}

const badgeLabel = "bitmap fallback"

// badgePad is the badge box pad and screen margin.
const badgePad = 8.0

// badgeInkHex packs badgeInk as 0xRRGGBBAA.
const badgeInkHex = 0x1c1915e6

var (
	badgeFace = text.NewGoXFace(basicfont.Face7x13)
	badgeInk  = devRGBA(badgeInkHex)
)

// badgeRect returns the top-right box that holds label on a screenW-wide
// frame with no dock.
func badgeRect(screenW float64, label string) (x, y, width, height float64) {
	return badgeRectInset(screenW, label, 0)
}

// badgeRectInset is badgeRect with the right edge inset pixels from the
// screen edge, so the devtools dock leaves the badge visible.
func badgeRectInset(screenW float64, label string, inset float64) (x, y, width, height float64) {
	labelW, labelH := text.Measure(label, badgeFace, 0)

	width = labelW + badgePad
	height = labelH + badgePad
	x = screenW - width - badgePad - inset
	if x < badgePad {
		x = badgePad
	}

	y = badgePad

	return x, y, width, height
}

// drawBadge paints the bitmap fallback notice on top of the frame. The
// devtools dock pushes it left.
func (s *shell) drawBadge(screen *ebiten.Image) {
	inset := 0.0
	if s.dev.on && !s.dev.mobile {
		inset = s.dev.panel.W
	}

	x, y, width, height := badgeRectInset(float64(s.screenW), badgeLabel, inset)

	vector.FillRect(
		screen,
		float32(x), float32(y), float32(width), float32(height),
		badgeInk, false,
	)

	var op text.DrawOptions
	op.GeoM.Translate(x+devLabelInset, y+devLabelInset)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, badgeLabel, badgeFace, &op)
	s.frameNote(frameRect(x, y, width, height))
}

// errViewportNotRGBA reports a viewport paint that is not NRGBA pixels.
// bitmap.Paint always produces NRGBA, so this is unreachable in practice;
// the checked assertion keeps the type gate honest.
var errViewportNotRGBA = errors.New("ownframe: viewport pixels are not NRGBA")

type bitmapView struct {
	img  *ebiten.Image
	gen  uint64
	area image.Rectangle
	zoom float64
}

func (s *shell) prepareBitmapViewport() error {
	// The provider is window-only; the browser sink serves PNG bytes and
	// never reads retained operations. See host.FallbackDisplayProvider.
	raw, ok := s.app.(host.FallbackDisplayProvider)
	if !s.fallback || !ok || raw.FallbackDisplay() == nil || s.stretched() || s.viewLocked() {
		if s.bitmapView.img != nil {
			s.bitmapView.img.Deallocate()
		}
		s.bitmapView = bitmapView{}
		return nil
	}
	viewW, viewH := s.viewSize()
	area := image.Rect(s.scrollX, s.scrollY, s.scrollX+viewW, s.scrollY+viewH)
	gen, zoom := s.app.Generation(), s.zoom()
	v := &s.bitmapView
	if v.img != nil && v.gen == gen && v.area == area && v.zoom == zoom {
		return nil
	}
	img, err := bitmap.Viewport(raw.FallbackDisplay(), area, zoom)
	if err != nil {
		return err
	}
	if img == nil {
		return nil
	}
	if v.img == nil || v.img.Bounds().Size() != area.Size() {
		if v.img != nil {
			v.img.Deallocate()
		}
		v.img = ebiten.NewImage(area.Dx(), area.Dy())
	}
	pixels, ok := img.(*image.NRGBA)
	if !ok {
		return errViewportNotRGBA
	}
	v.img.WritePixels(pixels.Pix)
	v.gen, v.area, v.zoom = gen, area, zoom
	return nil
}

// maxBlitRects caps one frame's screen blits; past it blitPlan collapses to
// one whole-buffer copy.
const maxBlitRects = 8

// blitPlan decides one frame's screen blits. full means the caller painted the
// whole screen, so no blits are needed. Rects are in buffer coordinates, which
// blitTiles offsets by the scroll and trims to the dock's visible page. A
// repaintBuffer plan means the screen repaints in full; repaintBlit
// contributes nothing. Overlapping rects merge; past the cap one union covers
// them all.
func blitPlan(plans []repaintPlan, chrome []image.Rectangle) (full bool, rects []image.Rectangle) {
	return blitPlanBuf(nil, plans, chrome)
}

func blitPlanBuf(buf []image.Rectangle, plans []repaintPlan, chrome []image.Rectangle) (full bool, rects []image.Rectangle) {
	rects = buf[:0]
	for _, plan := range plans {
		if plan.mode == repaintBuffer {
			return true, nil
		}

		if plan.mode == repaintRect && !plan.rect.Empty() {
			rects = append(rects, plan.rect)
		}
	}

	for _, rect := range chrome {
		if !rect.Empty() {
			rects = append(rects, rect)
		}
	}

	rects = mergeBlitRects(rects)
	if len(rects) > maxBlitRects {
		if cap(rects) > 0 {
			u := unionRects(rects)
			rects = rects[:1]
			rects[0] = u
			return false, rects
		}

		return false, []image.Rectangle{unionRects(rects)}
	}

	return false, rects
}

// mergeBlitRects unions overlapping rects until no pair overlaps. A frame
// brings at most the plan cap plus the chrome rects, so the restarting scan
// stays cheap.
func mergeBlitRects(rects []image.Rectangle) []image.Rectangle {
	for changed := true; changed; {
		changed = false

		for i := 0; i < len(rects) && !changed; i++ {
			for j := i + 1; j < len(rects); j++ {
				if rects[i].Overlaps(rects[j]) {
					rects[i] = rects[i].Union(rects[j])
					rects = append(rects[:j], rects[j+1:]...)
					changed = true

					break
				}
			}
		}
	}

	return rects
}

// unionRects is the smallest rect covering every rect. blitPlan uses it past
// the cap: one copy of the whole damaged area.
func unionRects(rects []image.Rectangle) image.Rectangle {
	if len(rects) == 0 {
		return image.Rectangle{}
	}

	box := rects[0]
	for _, rect := range rects[1:] {
		box = box.Union(rect)
	}

	return box
}

// chromeRectsBuf moves the previous frame's chrome rects from screen space
// into buffer space, reusing buf. The current scroll says which buffer pixels
// belong under the screen rect now, so a restore repairs stale chrome even
// after a scroll.
func chromeRectsBuf(buf []image.Rectangle, screen []image.Rectangle, scroll image.Point) []image.Rectangle {
	if len(screen) == 0 {
		return nil
	}

	out := buf[:0]
	for _, rect := range screen {
		out = append(out, rect.Add(scroll))
	}

	return out
}

// frameStale reports a frame whose retained pixels cannot be trusted: the
// first frame, a focus regain after the OS may have dropped the surface, or a
// screen size change, or a host that clears its surface every frame.
// It records focus and size for the next frame. Retained desktop surfaces
// can still lose pixels after an occlusion without a focus change.
func (s *shell) frameStale(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	stale := ebiten.IsScreenClearedEveryFrame() || !s.frameSeen() || s.partial.screen != size || focus && !s.partial.focus
	s.partial.focus, s.partial.screen = focus, size

	return stale
}

// frameState holds the chrome rects painted on the screen outside the page
// buffer: the devtools overlays, an open menu, each scrollbar thumb, and the
// fallback badge. A partial frame first restores the previous frame's rects
// from the buffer, so closed or moved chrome leaves no stale pixels. Rects
// are screen pixels; a blit offsets them into the buffer by the scroll.
//
// The slices swap backing arrays in frameCommit, so the steady state
// allocates nothing.
type frameState struct {
	restore []image.Rectangle // chrome rects drawn last frame
	current []image.Rectangle // chrome rects collected this frame
	seen    bool              // a frameCommit has run
}

// frameBegin starts this frame's chrome collection.
func (s *shell) frameBegin() {
	s.frame.current = s.frame.current[:0]
}

// frameNote records one chrome rect actually painted this frame. An empty
// rect is ignored because nothing was painted there.
func (s *shell) frameNote(r image.Rectangle) {
	if r.Empty() {
		return
	}

	s.frame.current = append(s.frame.current, r)
}

// frameRestore returns the chrome rects of the last committed frame. The
// caller must only read the slice.
func (s *shell) frameRestore() []image.Rectangle {
	return s.frame.restore
}

// frameSeen reports whether a frame has been committed.
func (s *shell) frameSeen() bool {
	return s.frame.seen
}

// frameCommit makes this frame's rects the restore set and recycles the old
// restore slice as the empty current slice.
func (s *shell) frameCommit() {
	s.frame.restore, s.frame.current = s.frame.current, s.frame.restore[:0]
	s.frame.seen = true
}

// frameRect rounds a painted float rectangle outward to whole pixels, so a
// partially covered edge pixel is restored too.
func frameRect(x, y, w, h float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(x)), int(math.Floor(y)),
		int(math.Ceil(x+w)), int(math.Ceil(y+h)),
	)
}
