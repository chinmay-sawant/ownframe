package window

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image"
	"image/color"
	"time"
)

// devState is the window's side of the devtools overlay: the mode flags, the
// boxes under the cursor and pinned, the panel rects, and the frame numbers.
// The page owns the on/off flag, so Config.DevTools and SetDevTools work.
type devState struct {
	mobile         bool
	mobileExpanded bool
	on             bool
	paused         bool
	refreshGap     time.Duration
	available      bool
	picker         bool
	ops            bool
	tab            devTab
	dockW          float64
	scroll         float64
	collapsed      map[string]bool
	rows           []devRow
	opsCache       devOpsCache
	panelCache     devPanelCache
	content        devRect
	hovered        layout.Box
	haveHov        bool
	pinned         layout.Box
	havePin        bool
	opPick         int
	haveOp         bool
	opPin          bool
	resizing       bool
	gen            uint64
	stats          host.Stats
	frame          time.Duration
	draw           time.Duration
	lastAt         time.Time
	ring           perfRing
	stages         perfStage
	sampler        perfSampler
	perfAvg        time.Duration
	perfP95        time.Duration
	perfP99        time.Duration
	perfLong       int
	runHeap        uint64
	runRSS         uint64
	runAlloc       uint64
	runGoroutines  int
	watch          keyWatch
	eaten          [len(devKeys)]bool
	panel          devRect
	hits           []devHit
	forward        bool
	// panelTrack holds dock fingers: lifts tap, drags scroll.
	panelTrack panelTrack
	// stroke is the outline hook. Nil strokes with vector.StrokeRect; tests
	// replace it to read the computed screen rects without pixels.
	stroke func(screen *ebiten.Image, r devRect, ink color.RGBA)
}

// devInspector returns the screen's inspector, or nil for a plain screen.
func (s *shell) devInspector() host.Inspector {
	insp, _ := s.app.(host.Inspector)

	return insp
}

// devRefresh samples the page stats and drops a box a new generation no
// longer has. It runs after the frame's picture is synced.
func (s *shell) devRefresh() {
	insp := s.devInspector()
	if insp == nil || !s.dev.on {
		return
	}

	if sink, ok := s.app.(drawTimer); ok {
		sink.SetDrawTime(s.dev.draw)
	}

	s.dev.stats = insp.Stats()

	gen := s.app.Generation()
	if gen == s.dev.gen {
		return
	}

	s.dev.gen = gen
	s.dev.haveOp = false
	if !s.dev.haveHov && !s.dev.havePin {
		return
	}
	boxes := s.boxes()

	if s.dev.haveHov && !devHasBox(boxes, s.dev.hovered) {
		s.dev.haveHov = false
	}

	if s.dev.havePin && !devHasBox(boxes, s.dev.pinned) {
		s.dev.havePin = false
	}
}

// devClose clears the picks and the dock fingers, and sends one hover for
// the current cursor, so a stale :hover does not stick after the overlay
// closes.
func (s *shell) devClose() error {
	s.devSetPicker(false)
	s.dev.panelTrack = panelTrack{}

	x, y := s.cursorPosition()
	frameW, frameH := s.frameSize()
	viewW, viewH := s.viewSize()
	px, py := contentPoint(x, y, s.scrollX, s.scrollY, s.stretched(), frameW, frameH, viewW, viewH)

	return s.ptr().Hover(s.ctx, px, py)
}

// devHasBox reports whether boxes still holds the same element. An id is
// enough when it has one; a box without an id matches its own rectangle.
func devHasBox(boxes []layout.Box, want layout.Box) bool {
	for _, b := range boxes {
		if want.ID != "" {
			if b.ID == want.ID {
				return true
			}

			continue
		}

		if b.Tag == want.Tag && b.X == want.X && b.Y == want.Y && b.W == want.W && b.H == want.H {
			return true
		}
	}

	return false
}

// devSync follows the page's flag once per frame, so Config.DevTools and a
// SetDevTools call open or close the overlay without a key.
func (s *shell) devSync() error {
	insp := s.devInspector()
	if insp == nil {
		s.dev.on = false
		s.perf = s.perfOption || perfEnabled(s.app)

		return nil
	}

	on := insp.DevTools()
	s.perf = s.perfOption || perfEnabled(s.app) || on
	if on {
		// A page that opened the overlay once keeps F12 available after it
		// closes, including a close before the first frame.
		s.dev.available = true
	}
	if on == s.dev.on {
		return nil
	}

	s.dev.on = on
	s.dev.panelCache.close()
	if !on {
		return s.devClose()
	}

	return nil
}

// devOpsCache keeps the Ops tab rows for one page generation. The tab has one
// row per painted operation, so a still generation must not rebuild the list.
// The outline toggle changes a row without a new generation, so it is a key
// too. Generation zero never caches, so a fake screen in a test stays fresh.
type devOpsCache struct {
	rows    []devRow
	gen     uint64
	toggle  bool
	valid   bool
	display *layout.Display
}

// devOpsRows returns the Ops tab rows, rebuilt only when the page generation,
// the display, or the outline toggle moved.
func (s *shell) devOpsRows() []devRow {
	gen := s.devGeneration()
	c := &s.dev.opsCache

	if gen != 0 && c.valid && c.gen == gen && c.toggle == s.dev.ops && c.display == s.display {
		return c.rows
	}

	rows := s.devOpsBuildRows()
	if gen != 0 {
		*c = devOpsCache{rows: rows, gen: gen, toggle: s.dev.ops, valid: true, display: s.display}
	}

	return rows
}

// devGeneration is the page's build generation, zero before a screen exists.
func (s *shell) devGeneration() uint64 {
	if s.app == nil {
		return 0
	}

	return s.app.Generation()
}

// devOwns reports whether the overlay owns one key under these modifiers.
func (s *shell) devOwns(key ebiten.Key, mods modifiers) bool {
	insp := s.devInspector()
	if insp == nil {
		return false
	}

	// Keep the overlay available only for screens that opted in through
	// Config.DevTools or SetDevTools. A plain screen never sees F12.
	available := s.dev.available || s.dev.on || insp.DevTools()
	if !available {
		return false
	}
	s.dev.available = true

	switch key {
	case ebiten.KeyF12:
		return true
	case ebiten.KeyI:
		return mods.Control && mods.Shift
	case ebiten.KeyO:
		return s.dev.on && !mods.command()
	}

	return false
}

func devKeyIndex(key ebiten.Key) int {
	for i, k := range devKeys {
		if k == key {
			return i
		}
	}

	return -1
}

// devToggle requests the change through the page flag. devSync applies it
// before resize on the next update, including toggles from pointer handling.
func (s *shell) devToggle() error {
	insp := s.devInspector()
	if insp == nil {
		return nil
	}

	insp.SetDevTools(!insp.DevTools())
	s.dev.available = true

	return nil
}

// devEats reports whether keyEvents must leave one key to the overlay.
func (s *shell) devEats(key ebiten.Key) bool {
	i := devKeyIndex(key)

	return i >= 0 && s.dev.eaten[i]
}

// pageKeyStep is the keyEvents step for one key. A key the overlay owns
// never reaches the page watch, so its press and release stay invisible.
func (s *shell) pageKeyStep(key ebiten.Key, pressed bool) (bool, bool) {
	if s.devEats(key) {
		return false, false
	}

	return s.watched.step(key, pressed)
}

// devFrameTime records the wall time since the previous Draw call.
func (s *shell) devFrameTime() {
	now := time.Now()
	if !s.dev.lastAt.IsZero() {
		s.dev.frame = now.Sub(s.dev.lastAt)
	}

	s.dev.lastAt = now
}

// devMS formats a duration for the panel. A zero duration reads as zero
// instead of a negative or rounded value.
func devMS(d time.Duration) string {
	if d <= 0 {
		return "0.0ms"
	}

	return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond))
}

// devErr prints an empty error as a dash for a counter row.
func devErr(err string) string {
	if err == "" {
		return "-"
	}

	return err
}

// devClip shortens a string to n runes, with dots when it was cut.
func devClip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}

	return string(r[:n]) + "..."
}

// devBoxName is tag or tag#id for labels and the target row.
func devBoxName(box layout.Box) string {
	if box.ID != "" {
		return box.Tag + "#" + box.ID
	}

	return box.Tag
}

// Shared devtools UI numbers used by two or more files. File-local
// numbers stay in their own file, except devScrollMin and
// devScrollMargin, kept here to keep the scrollbar file under the cap.
const (
	// devTabPad is the horizontal pad around a tab name.
	devTabPad = 20.0
	// devHeaderGap is the gap between header buttons. The Picker hit covers
	// only its own label and the X hit covers only its key; the F12 hint
	// between them takes no hit, so dead space sits on both sides of it.
	devHeaderGap = 8.0
	// devLabelInset is the label text inset and outline gap.
	devLabelInset = 4.0
	// devCenterDiv centers by halving a height difference.
	devCenterDiv = 2.0
	// devScrollMin floors the scrollbar thumb.
	devScrollMin = 24.0
	// devScrollMargin insets the scrollbar thumb from the edge.
	devScrollMargin = 2.0
)

const (
	devShiftR = 24
	devShiftG = 16
	devShiftB = 8
)

// devRGBA builds an opaque-ish ink from packed 0xRRGGBBAA.
func devRGBA(hex uint32) color.RGBA {
	return color.RGBA{
		R: uint8(hex >> devShiftR),
		G: uint8(hex >> devShiftG), //nolint:gosec // reason: shift keeps channel in byte range.
		B: uint8(hex >> devShiftB), //nolint:gosec // reason: shift keeps channel in byte range.
		A: uint8(hex),              //nolint:gosec // reason: low byte holds alpha.
	}
}

// devOpCounts is the per-kind operation count the panel prints.
type devOpCounts struct {
	Fill   int
	Stroke int
	Line   int
	Text   int
	Image  int
	Grid   int
	Path   int
}

func (c devOpCounts) total() int {
	return c.Fill + c.Stroke + c.Line + c.Text + c.Image + c.Grid + c.Path
}

// devCountOps counts the operations that paint, in one pass.
func devCountOps(display *layout.Display) devOpCounts {
	var c devOpCounts
	if display == nil {
		return c
	}

	for i := range display.Ops {
		switch display.Ops[i].Kind {
		case layout.DisplayOpFillRect:
			c.Fill++
		case layout.DisplayOpStrokeRect:
			c.Stroke++
		case layout.DisplayOpLine:
			c.Line++
		case layout.DisplayOpText, layout.DisplayOpBullet:
			c.Text++
		case layout.DisplayOpImage:
			c.Image++
		case layout.DisplayOpGridRun:
			c.Grid++
		case layout.DisplayOpPath:
			c.Path++
		}
	}

	return c
}

// devOpsInBox counts the painted operations inside one box.
func devOpsInBox(display *layout.Display, box layout.Box) int {
	count := 0

	for i := range display.Ops {
		op := &display.Ops[i]
		if _, ok := devOpInk(op.Kind); !ok {
			continue
		}

		r := devOpRect(op, display.PointsPerPixel)
		if r.X >= box.X && r.Y >= box.Y && r.X+r.W <= box.X+box.W && r.Y+r.H <= box.Y+box.H {
			count++
		}
	}

	return count
}

// devLineGap is the leading between panel lines.
const devLineGap = 2.0

// devLineHeight is the badge face line height plus a little leading.
func devLineHeight() float64 {
	return devFaceHeight() + devLineGap
}

// devTextAt draws one run at an exact panel point.
func devTextAt(screen *ebiten.Image, label string, x, y float64, ink color.RGBA) {
	var op text.DrawOptions
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(screen, label, badgeFace, &op)
}

// drawDevContent paints the active tab's rows, clipped to the content rect,
// with a hover band and a blue band on the chosen operation.
func (s *shell) drawDevContent(screen *ebiten.Image) {
	c := s.dev.content
	if c.W <= 0 || c.H <= 0 {
		return
	}

	clip, ok := screen.SubImage(image.Rect(int(c.X), int(c.Y), int(c.X+c.W), int(c.Y+c.H))).(*ebiten.Image)
	if !ok {
		return
	}
	lineH := devLineHeight()
	y := c.Y - s.dev.scroll

	for _, row := range s.dev.rows {
		if y+lineH > c.Y && y < c.Y+c.H {
			s.devRowBack(screen, row, y, lineH)
			s.devDrawLine(clip, row.line, c.X, y, row.w)
		}

		y += lineH
	}
}

// devRowHot reports the pointer inside one row band.
func (s *shell) devRowHot(y, lineH float64) bool {
	cy := float64(s.cursorY)

	return cy >= y && cy < y+lineH &&
		s.cursorX >= int(s.dev.panel.X) && s.cursorX <= int(s.dev.panel.X+s.dev.panel.W)
}

// devKeys are the keys the overlay owns while a page has an inspector.
var devKeys = [...]ebiten.Key{ebiten.KeyF12, ebiten.KeyI, ebiten.KeyO}

// drawTimer is the optional hook a screen implements to receive the window's
// draw time. host.Inspector keeps its three-method shape.
type drawTimer interface {
	SetDrawTime(time.Duration)
}

// devtoolsKeys steps the overlay's keys before the page sees them. A key the
// overlay owns is swallowed until it goes up, and a swallowed key also drops
// the printable input of the frame, so "o" toggles the operation view
// without typing.
func (s *shell) devtoolsKeys(mods modifiers) (bool, error) {
	ate := false

	for _, key := range devKeys {
		swallowed, err := s.devKeyStep(key, ebiten.IsKeyPressed(key), mods)
		if err != nil {
			return false, err
		}

		ate = ate || swallowed
	}

	return ate, nil
}

// devKeyStep advances the overlay watch for one key and reports whether the
// key belongs to the overlay this frame.
func (s *shell) devKeyStep(key ebiten.Key, pressed bool, mods modifiers) (bool, error) {
	i := devKeyIndex(key)
	if i < 0 {
		return false, nil
	}

	if !s.dev.eaten[i] && !s.devOwns(key, mods) {
		return false, nil
	}

	down, _ := s.dev.watch.step(key, pressed)
	if !pressed {
		swallowed := s.dev.eaten[i]
		s.dev.eaten[i] = false

		return swallowed, nil
	}

	s.dev.eaten[i] = true
	if !down {
		return true, nil
	}

	return true, s.devKeyFired(key)
}

// devKeyFired runs the action of one overlay key.
func (s *shell) devKeyFired(key ebiten.Key) error {
	if key == ebiten.KeyO {
		s.dev.ops = !s.dev.ops
		s.dev.tab = devTabOps

		return nil
	}

	return s.devToggle()
}
