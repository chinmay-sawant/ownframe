package window

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"sync"
	"time"
)

// drawDevTools paints the overlay over the frame: the operation outlines,
// the chosen operation, the hovered and pinned boxes, and the right dock.
// It only reads shell state, so the page picture never changes.
func (s *shell) drawDevTools(screen *ebiten.Image) {
	refresh := s.refreshDevPanel(time.Now())

	if s.dev.ops {
		s.drawDevOps(screen)
	}

	if s.dev.haveOp {
		s.drawDevOpPick(screen)
	}

	s.drawDevPick(screen)
	s.drawCachedDevPanel(screen, refresh)
}

// devOutlineWidth strokes a box; devLabelPadX/TDy pad a label.
const (
	devOutlineWidth = 2.0
	devLabelPadX    = 8.0
	devLabelTextDy  = 2.0
)

// devOutline strokes one box and labels it. The stroke hook lets a test read
// the computed rect without reading pixels.
func (s *shell) devOutline(screen *ebiten.Image, box layout.Box, ink color.RGBA) {
	r := s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})

	if s.dev.stroke != nil {
		s.dev.stroke(screen, r, ink)
	} else {
		vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devOutlineWidth, ink, false)
	}

	// The 2 px stroke reaches 1 px past each edge.
	lr := s.devLabel(screen, r, devBoxLabel(box))
	s.frameNote(frameRect(r.X-1, r.Y-1, r.W+devOutlineWidth, r.H+devOutlineWidth).Union(frameRect(lr.X, lr.Y, lr.W, lr.H)))
}

// devBoxLabel is the tag, id, and size line drawn over a box.
func devBoxLabel(box layout.Box) string {
	name := box.Tag
	if box.ID != "" {
		name += "#" + box.ID
	}

	return fmt.Sprintf("%s  %gx%g", name, box.W, box.H)
}

// devLabel draws a box label above the outline, or below it when the top
// edge would clip, and returns the rect it painted.
func (s *shell) devLabel(screen *ebiten.Image, r devRect, label string) devRect {
	w, h := text.Measure(label, badgeFace, 0)
	x, y := r.X, r.Y-h-devLabelInset

	if y < 0 {
		y = r.Y + r.H + devLabelInset
	}

	vector.FillRect(screen, float32(x), float32(y), float32(w+devLabelPadX), float32(h+devLabelInset), badgeInk, false)

	var op text.DrawOptions
	op.GeoM.Translate(x+devLabelInset, y+devLabelTextDy)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, label, badgeFace, &op)

	return devRect{X: x, Y: y, W: w + devLabelPadX, H: h + devLabelInset}
}

// The devtools dock follows the Chrome DevTools dark theme: a dark grey dock,
// a lighter header and tab strip, a blue accent, and syntax colours for the
// JSON view. Each hex packs 0xRRGGBBAA.
const (
	devPanelBgHex  = 0x202124ff
	devHeaderBgHex = 0x292a2dff
	devBorderHex   = 0x3c4043ff
	devFgHex       = 0xe8eaedff
	devDimHex      = 0x9aa0a6ff
	devAccentHex   = 0x8ab4f8ff
	devErrInkHex   = 0xf28b82ff
	devKeyInkHex   = 0x9cdcfeff
	devStrInkHex   = 0xce9178ff
	devNumInkHex   = 0xb5cea8ff
	devBoolInkHex  = 0x569cd6ff
	devPunctInkHex = 0xd4d4d4ff
	devRowHoverHex = 0x35363aff
	devRowPickHex  = 0x3d4d6dff
	devFillInkHex  = 0x1a56dbff
	devStrokeHex   = 0xe5484dff
	devLineInkHex  = 0xd97706ff
	devTextInkHex  = 0x176b45ff
	devImageInkHex = 0x7c3aedff
	devGridInkHex  = 0x0891b2ff
)

var (
	devPanelBg   = devRGBA(devPanelBgHex)
	devHeaderBg  = devRGBA(devHeaderBgHex)
	devBorder    = devRGBA(devBorderHex)
	devFg        = devRGBA(devFgHex)
	devDim       = devRGBA(devDimHex)
	devAccent    = devRGBA(devAccentHex)
	devErrInk    = devRGBA(devErrInkHex)
	devKeyInk    = devRGBA(devKeyInkHex)
	devStrInk    = devRGBA(devStrInkHex)
	devNumInk    = devRGBA(devNumInkHex)
	devBoolInk   = devRGBA(devBoolInkHex)
	devPunctInk  = devRGBA(devPunctInkHex)
	devRowHover  = devRGBA(devRowHoverHex)
	devRowPick   = devRGBA(devRowPickHex)
	devFillInk   = devRGBA(devFillInkHex)
	devStrokeInk = devRGBA(devStrokeHex)
	devLineInk   = devRGBA(devLineInkHex)
	devTextInk   = devRGBA(devTextInkHex)
	devImageInk  = devRGBA(devImageInkHex)
	devGridInk   = devRGBA(devGridInkHex)
	devHoverInk  = devRGBA(devStrokeHex)
	devPinInk    = devRGBA(devFillInkHex)
)

// devMeasure memoizes panel text widths and the badge face line height.
// Panel labels repeat on every frame and text.Measure shapes its run per
// call, so the memo turns a frame of label measures into map hits.
var devMeasure = struct {
	mu    sync.Mutex
	width map[string]float64
	faceH float64
}{}

// devTextWidth returns the badge-face width of s, measuring it once.
func devTextWidth(s string) float64 {
	devMeasure.mu.Lock()
	defer devMeasure.mu.Unlock()

	if w, ok := devMeasure.width[s]; ok {
		return w
	}

	w, _ := text.Measure(s, badgeFace, 0)
	if devMeasure.width == nil {
		devMeasure.width = map[string]float64{}
	}
	devMeasure.width[s] = w

	return w
}

// devFaceHeight returns the badge face line height, measured once.
func devFaceHeight() float64 {
	devMeasure.mu.Lock()
	defer devMeasure.mu.Unlock()

	if devMeasure.faceH == 0 {
		_, devMeasure.faceH = text.Measure("Ag", badgeFace, 0)
	}

	return devMeasure.faceH
}

// drawDevPanel paints the dock: the background, the header, the tab strip,
// the content, the scrollbar, and the footer hint.
func (s *shell) drawDevPanel(screen *ebiten.Image) {
	if s.dev.mobile {
		s.drawDevMobile(screen)
		return
	}
	r := s.dev.panel
	if r.W <= 0 || r.H <= 0 {
		return
	}

	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devPanelBg, false)
	vector.FillRect(screen, float32(r.X), float32(r.Y), 1, float32(r.H), devBorder, false)

	s.drawDevHeader(screen)
	s.drawDevTabs(screen)
	s.drawDevContent(screen)
	s.drawDevScrollbar(screen)
	s.drawDevFooter(screen)
}

// drawDevHeader paints the title bar, the picker toggle, and the close key hint.
func (s *shell) drawDevHeader(screen *ebiten.Image) {
	r := s.dev.panel
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), devHeaderH, devHeaderBg, false)

	top := r.Y + (devHeaderH-devFaceHeight())/devCenterDiv
	devTextAt(screen, "Developer Tools", r.X+devPanelPad, top, devFg)

	label := "Picker"
	ink := devDim
	if s.devPickerOn() {
		ink = devAccent
	}
	_, px := devHeaderRight(r)
	devTextAt(screen, label, px, top, ink)

	devTextAt(screen, "F12", devHeaderF12X(r), top, devDim)
	xX := r.X + r.W - devPanelPad - devTextWidth("X")
	devTextAt(screen, "X", xX, top, devDim)
}

// drawDevFooter paints the hint line at the bottom of the dock.
func (s *shell) drawDevFooter(screen *ebiten.Image) {
	r := s.dev.panel
	y := r.Y + r.H - devFooterH
	vector.FillRect(screen, float32(r.X), float32(y), float32(r.W), 1, devBorder, false)
	devTextAt(screen, devTabHint(s.dev.tab), r.X+devPanelPad, y+(devFooterH-devFaceHeight())/devCenterDiv, devDim)
}

const (
	devPanelRefreshGap = time.Second
	devPanelFastGap    = 250 * time.Millisecond
	devPanelSlowGap    = 5 * time.Second
)

// The panel owns one texture. Metrics refresh at 1 Hz by default; input and geometry
// changes invalidate it immediately. Page outlines are drawn separately.
type devPanelCache struct {
	img   *ebiten.Image
	key   devPanelKey
	next  time.Time
	valid bool
}

type devPanelKey struct {
	width, height    int
	dockWidth        float64
	tab              devTab
	scroll           float64
	mobile, expanded bool
	picker, ops      bool
	haveOp, havePin  bool
	opPick           int
	cursorX, cursorY int
	gen              uint64
	display          *layout.Display
	selection        devElementJSON
	paused           bool
	refreshGap       time.Duration
}

func (s *shell) devPanelKey() devPanelKey {
	key := devPanelKey{
		paused: s.dev.paused, refreshGap: s.dev.refreshGap,
		width: s.screenW, height: s.screenH, dockWidth: s.devDockWidth(),
		tab: s.dev.tab, scroll: s.dev.scroll,
		mobile: s.dev.mobile, expanded: s.dev.mobileExpanded,
		picker: s.dev.picker, ops: s.dev.ops,
		haveOp: s.dev.haveOp, havePin: s.dev.havePin, opPick: s.dev.opPick,
	}
	if !s.dev.mobile && s.dev.tab != devTabFrame {
		key.gen, key.display = s.devGeneration(), s.display
	}
	if devInRect(s.dev.panel, float64(s.cursorX), float64(s.cursorY)) {
		key.cursorX, key.cursorY = s.cursorX, s.cursorY
	}
	if box, ok := s.devBox(); ok {
		key.selection = devElementJSON{
			Tag: box.Tag, ID: box.ID, Action: box.Action, Text: box.Text,
			Rect: devElementRect{X: box.X, Y: box.Y, Width: box.W, Height: box.H},
		}
	}
	return key
}

func (s *shell) refreshDevPanel(now time.Time) bool {
	c := &s.dev.panelCache
	key := s.devPanelKey()
	if c.valid && c.key == key && (c.next.IsZero() || now.Before(c.next)) {
		return false
	}
	s.devLayout()
	c.key = s.devPanelKey()
	c.next = time.Time{}
	if (s.dev.mobile || s.dev.tab == devTabFrame) && !s.dev.paused {
		gap := s.dev.refreshGap
		if gap == 0 {
			gap = devPanelRefreshGap
		}
		c.next = now.Add(gap)
	}
	c.valid = true
	return true
}

func (s *shell) drawCachedDevPanel(screen *ebiten.Image, refresh bool) {
	r := s.dev.panel
	bounds := frameRect(r.X, r.Y, r.W, r.H).Intersect(screen.Bounds())
	if bounds.Empty() {
		return
	}
	c := &s.dev.panelCache
	if c.img == nil || c.img.Bounds() != bounds {
		if c.img != nil {
			c.img.Deallocate()
		}
		c.img = ebiten.NewImageWithOptions(bounds, &ebiten.NewImageOptions{Unmanaged: true})
		refresh = true
	}
	if refresh {
		c.img.Clear()
		s.drawDevPanel(c.img)
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	screen.DrawImage(c.img, opts)
	s.frameNote(bounds)
}

func (c *devPanelCache) close() {
	if c.img != nil {
		c.img.Deallocate()
	}
	*c = devPanelCache{}
}

// panelFinger is one finger inside the devtools dock: its landing point,
// its latest position, and whether it dragged past the tap slop. A finger
// that lifts without dragging taps at its release point; a dragged finger
// scrolls and never fires.
type panelFinger struct {
	id    ebiten.TouchID
	x, y  int
	lastX int
	lastY int
	drag  bool
}

// panelTrack is the dock's finger state across frames.
type panelTrack struct {
	fingers []panelFinger
}

// panelFind returns the tracked finger with an id, or -1.
func panelFind(fs []panelFinger, id ebiten.TouchID) int {
	for i, f := range fs {
		if f.id == id {
			return i
		}
	}
	return -1
}

// devSpan is one run on a panel line: text in one ink, or a filled swatch
// when swatch is positive. w is the measured text width when the builder
// knows it; a zero w measures while drawing. A right span starts at the
// line's right edge, which needs its w.
type devSpan struct {
	text   string
	ink    color.RGBA
	swatch float64
	w      float64
	right  bool
}

// devLine is one panel line, drawn left to right.
type devLine struct {
	spans []devSpan
}

// devAct is what a panel row does when a click lands on it.
type devAct int

const (
	devActNone         devAct = iota
	devActTab                 // select the tab named by arg
	devActOpsToggle           // flip the operation outlines
	devActJSON                // collapse or expand the JSON node named by key
	devActOp                  // outline the operation at index arg
	devActPicker              // toggle element picker mode
	devActClose               // close the dock
	devActMobileExpand        // expand or collapse the mobile metrics sheet
	devActPause
	devActRefresh
)

// devRow is one content line plus the click action it carries. w is the line
// width the builder measured against, for right-aligned spans.
type devRow struct {
	line devLine
	act  devAct
	arg  int
	key  string
	w    float64
}

// devSwatchSize is the side of a colour chip in a panel line.
const devSwatchSize = 9.0

// devSwatchGap is the space between a swatch and the text after it.
const devSwatchGap = 6.0

// devText builds a one-span line.
func devText(s string, ink color.RGBA) devLine {
	return devLine{spans: []devSpan{{text: s, ink: ink}}}
}

// devPlain builds a line in the default panel ink.
func devPlain(s string) devLine {
	return devText(s, devFg)
}

// devSwatch builds a colour chip span.
func devSwatch(ink color.RGBA) devSpan {
	return devSpan{ink: ink, swatch: devSwatchSize}
}

// plain is the line without colour, for labels and tests.
func (l devLine) plain() string {
	out := ""
	for _, span := range l.spans {
		out += span.text
	}

	return out
}

// devRowBack paints the row band: blue on the chosen operation, grey under
// the pointer on any actionable row.
func (s *shell) devRowBack(screen *ebiten.Image, row devRow, y, lineH float64) {
	picked := row.act == devActOp && s.dev.haveOp && row.arg == s.dev.opPick
	if !picked && !(row.act != devActNone && s.devRowHot(y, lineH)) {
		return
	}

	ink := devRowHover
	if picked {
		ink = devRowPick
	}

	vector.FillRect(screen, float32(s.dev.panel.X), float32(y), float32(s.dev.panel.W), float32(lineH), ink, false)
}

// devDrawLine draws one row's spans in order, swatches included. width is the
// line's content width, so a right-aligned span starts at the right edge.
func (s *shell) devDrawLine(screen *ebiten.Image, line devLine, x, y, width float64) {
	start := x

	for _, span := range line.spans {
		if span.swatch > 0 {
			top := y + (devLineHeight()-span.swatch)/devCenterDiv
			vector.FillRect(screen, float32(x), float32(top), float32(span.swatch), float32(span.swatch), span.ink, false)
			x += span.swatch + devSwatchGap

			continue
		}

		if span.right {
			x = start + width - span.w
		}

		devTextAt(screen, span.text, x, y, span.ink)

		if span.w > 0 {
			x += span.w

			continue
		}

		w, _ := text.Measure(span.text, badgeFace, 0)
		x += w
	}
}

// drawDevScrollbar paints a thumb when the rows overflow the viewport.
func (s *shell) drawDevScrollbar(screen *ebiten.Image) {
	c := s.dev.content
	total := float64(len(s.dev.rows)) * devLineHeight()
	if total <= c.H || c.H <= 0 {
		return
	}

	const thick = 4.0

	length := c.H * c.H / total
	if length < devScrollMin {
		length = devScrollMin
	}

	pos := c.Y + (c.H-length)*s.dev.scroll/(total-c.H)
	x := s.dev.panel.X + s.dev.panel.W - thick - devScrollMargin
	vector.FillRect(screen, float32(x), float32(pos), thick, float32(length), devBorder, false)
}
