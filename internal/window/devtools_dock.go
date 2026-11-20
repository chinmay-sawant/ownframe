package window

import (
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"math"
)

const (
	// devDockWidth is the widest the right-hand panel opens, on a roomy
	// screen.
	devDockWidth = 340.0

	// devDockMin is the narrowest the panel may become while resizing.
	devDockMin = 220.0

	// devDockPageMin is the page strip the panel must leave beside it on a
	// desktop screen with room for that strip.
	devDockPageMin = 320.0

	// devDockRoomyW is the window width at which the panel opens at its full
	// devDockWidth. Narrower windows give the panel a share of the window
	// instead, so the page keeps a usable strip.
	devDockRoomyW = 1180

	// devDockShare is the fraction of a window narrower than devDockRoomyW
	// the panel takes, between devDockMin and devDockWidth.
	devDockShare = 0.28

	// devDockGrab is the half-width of the resize edge beside the panel.
	devDockGrab = 4.0

	// devDockGrabWidth spans the resize edge on both sides.
	devDockGrabWidth = 2 * devDockGrab
)

// devDockDefaultWidth is the panel width on a window with no stored drag
// width: devDockWidth on a roomy screen, else devDockShare of the window
// between devDockMin and devDockWidth, rounded down. It never exceeds the
// window, so a narrow screen still shows the panel alone.
func devDockDefaultWidth(screenW int) float64 {
	if screenW >= devDockRoomyW {
		return devDockWidth
	}

	w := float64(screenW) * devDockShare
	if w > devDockWidth {
		w = devDockWidth
	}

	if w < devDockMin {
		w = devDockMin
	}

	w = math.Floor(w)

	if maxW := float64(screenW); w > maxW {
		w = maxW
	}

	return w
}

// devDockRect places the devtools panel on the right edge, full height. A
// width larger than the window is clamped, so a phone shows the panel alone.
func devDockRect(screenW, screenH int, width float64) devRect {
	if width < 0 {
		width = 0
	}

	if width > float64(screenW) {
		width = float64(screenW)
	}

	return devRect{X: float64(screenW) - width, Y: 0, W: width, H: float64(screenH)}
}

// devDockWidth returns the dock width for this screen: the stored drag width
// clamped so the page keeps devDockPageMin beside it, else the responsive
// default. A screen narrower than devDockMin keeps the old clamp: the width
// is capped at the screen, so the default fills it.
func (s *shell) devDockWidth() float64 {
	screenW := float64(s.screenW)

	w := s.dev.dockW
	if w <= 0 {
		w = devDockDefaultWidth(s.screenW)
	}

	if screenW < devDockMin {
		if w > screenW {
			w = screenW
		}

		return w
	}

	if maxW := screenW - devDockPageMin; w > maxW {
		w = maxW
	}

	if w < devDockMin {
		w = devDockMin
	}

	return w
}

// devInRect reports whether a point is inside a rectangle.
func devInRect(r devRect, x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

// devOnEdge reports the resize strip just left of the dock. It is off on a
// screen the dock already fills.
func devOnEdge(panel devRect, x, y int) bool {
	if panel.X <= devDockGrab {
		return false
	}

	edge := devRect{X: panel.X - devDockGrab, W: devDockGrabWidth, H: panel.H}

	return devInRect(edge, float64(x), float64(y))
}

// devOwnsPoint reports whether the dock owns the pointer: inside the panel
// or on its resize edge. A custom-chrome draggable strip must not start a
// window drag there, or the header close key never receives the press.
func (s *shell) devOwnsPoint(x, y int) bool {
	if !s.dev.on {
		return false
	}

	panel := s.dev.panel
	if panel.W <= 0 || panel.H <= 0 {
		if s.dev.mobile {
			panel = devMobileRect(s.screenW, s.screenH, s.dev.mobileExpanded)
		} else {
			panel = devDockRect(s.screenW, s.screenH, s.devDockWidth())
		}
	}

	if devInRect(panel, float64(x), float64(y)) {
		return true
	}

	return !s.dev.mobile && devOnEdge(panel, x, y)
}

// devRect is a rectangle in screen pixels.
type devRect struct {
	X, Y, W, H float64
}

// devHit is one clickable region of the panel: a screen rect and the action
// a click runs.
type devHit struct {
	rect devRect
	act  devAct
	arg  int
	key  string
}

// devScreenRect converts a rectangle in page space to screen pixels. The
// page draws through the pinch zoom, so a page point lands at
// point*zoom - scroll. Stretched: scale by screenW/frameW and screenH/frameH
// as well, the factors drawReplayScaled uses.
func devScreenRect(r devRect, scrollX, scrollY int, zoom float64, stretched bool, frameW, frameH, screenW, screenH int) devRect {
	if zoom <= 0 {
		zoom = 1
	}

	if !stretched {
		return devRect{
			X: r.X*zoom - float64(scrollX),
			Y: r.Y*zoom - float64(scrollY),
			W: r.W * zoom,
			H: r.H * zoom,
		}
	}

	sx, sy := 1.0, 1.0

	if frameW > 0 && screenW > 0 {
		sx = float64(screenW) / float64(frameW)
	}

	if frameH > 0 && screenH > 0 {
		sy = float64(screenH) / float64(frameH)
	}

	return devRect{
		X: r.X * sx * zoom,
		Y: r.Y * sy * zoom,
		W: r.W * sx * zoom,
		H: r.H * sy * zoom,
	}
}

// devScreen converts a page-space rectangle for this frame.
func (s *shell) devScreen(r devRect) devRect {
	frameW, frameH := s.frameSize()
	viewW, viewH := s.viewSize()

	return devScreenRect(r, s.scrollX, s.scrollY, s.zoom(), s.stretched(), frameW, frameH, viewW, viewH)
}

const (
	devPanelPad = 8.0
	devHeaderH  = 26.0
	devTabsH    = 28.0
	devFooterH  = 22.0
)

// devLayout places the dock, builds the active tab's rows, and records this
// frame's clickable regions.
func (s *shell) devLayout() {
	if s.dev.mobile {
		s.devMobileLayout()
		return
	}
	s.dev.panel = devDockRect(s.screenW, s.screenH, s.devDockWidth())
	s.dev.content = devRect{
		X: s.dev.panel.X + devPanelPad,
		Y: s.dev.panel.Y + devHeaderH + devTabsH + devPanelPad,
		W: s.dev.panel.W - 2*devPanelPad,
		H: s.dev.panel.H - devHeaderH - devTabsH - devFooterH - 2*devPanelPad,
	}

	s.dev.rows = s.devTabRows()
	s.devClampScroll()
	s.devHits()
}

// devHits records the tab strip and every actionable content row. Rows are
// placed at the scroll offset the last frame drew, so the pointer reads the
// same layout it sees.
func (s *shell) devHits() {
	lineH := devLineHeight()
	s.dev.hits = s.dev.hits[:0]

	x := s.dev.panel.X + devPanelPad

	s.devPickerHit()

	cls, _ := devHeaderRight(s.dev.panel)
	s.dev.hits = append(s.dev.hits, devHit{rect: cls, act: devActClose})

	for i, name := range devTabNames {
		w := devTextWidth(name)
		s.dev.hits = append(s.dev.hits, devHit{
			rect: devRect{X: x, Y: s.dev.panel.Y + devHeaderH, W: w + devTabPad, H: devTabsH},
			act:  devActTab,
			arg:  i,
		})
		x += w + devTabPad
	}

	y := s.dev.content.Y - s.dev.scroll

	for _, row := range s.dev.rows {
		rect := devRect{X: s.dev.panel.X, Y: y, W: s.dev.panel.W, H: lineH}
		y += lineH

		if row.act == devActNone || rect.Y+lineH < s.dev.content.Y || rect.Y > s.dev.content.Y+s.dev.content.H {
			continue
		}

		s.dev.hits = append(s.dev.hits, devHit{rect: rect, act: row.act, arg: row.arg, key: row.key})
	}
}

// devTabRows builds the content of the active tab.
func (s *shell) devTabRows() []devRow {
	switch s.dev.tab {
	case devTabFrame:
		return s.devFrameRows(s.dev.content.W)
	case devTabOps:
		return s.devOpsRows()
	default:
		return s.devElementRows()
	}
}

const (
	devMobileHeader     = 48.0
	devMobileSafeBottom = 32.0
	devMobileFraction   = 0.4
)

func devMobileRect(w, h int, expanded bool) devRect {
	available := math.Max(0, float64(h)-devMobileSafeBottom)
	height := math.Min(available, devMobileHeader)
	if expanded {
		height = math.Min(available, math.Max(devMobileHeader, float64(h)*devMobileFraction))
	}
	return devRect{Y: available - height, W: math.Max(0, float64(w)), H: height}
}

func (s *shell) devMobileLayout() {
	s.dev.panel = devMobileRect(s.screenW, s.screenH, s.dev.mobileExpanded)
	r := s.dev.panel
	s.dev.content = devRect{X: devPanelPad, Y: r.Y + devMobileHeader, W: math.Max(0, r.W-2*devPanelPad), H: math.Max(0, r.H-devMobileHeader)}
	s.dev.rows = nil
	if s.dev.mobileExpanded {
		s.dev.rows = s.devMobileRows(s.dev.content.W)
	}
	s.devClampScroll()
	closeWidth := math.Min(r.W, devMobileHeader)
	headerHeight := math.Min(r.H, devMobileHeader)
	s.dev.hits = []devHit{
		{rect: devRect{X: r.X + r.W - closeWidth, Y: r.Y, W: closeWidth, H: headerHeight}, act: devActClose},
		{rect: devRect{X: r.X, Y: r.Y, W: r.W - closeWidth, H: headerHeight}, act: devActMobileExpand},
	}
}

func (s *shell) devMobileRows(width float64) []devRow {
	rows := []devRow{}
	header := func(label string) { rows = append(rows, devRow{line: devText(label, devDim)}) }
	value := func(label string, n any) { rows = append(rows, devFrameRow(width, label, fmt.Sprint(n), devNumInk)) }
	value("FPS", fmt.Sprintf("%.1f", ebiten.ActualFPS()))
	value("TPS", fmt.Sprintf("%.1f", ebiten.ActualTPS()))
	s.addPerfRows(&rows, width, header, value)
	s.addMemoryRows(&rows, width, header, value)
	header("Pipeline")
	value("Redraws", s.dev.stats.Redraws)
	value("Layouts", s.dev.stats.Layouts)
	value("Repaints", s.dev.stats.Repaints)
	s.addPipelinePerfRows(&rows, width, header, value, s.dev.stats)
	return rows
}

func (s *shell) drawDevMobile(screen *ebiten.Image) {
	r := s.dev.panel
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devPanelBg, false)
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(math.Min(r.H, devMobileHeader)), devHeaderBg, false)
	label := "Metrics  |  tap to expand"
	if s.dev.mobileExpanded {
		label = "Metrics  |  tap to collapse"
	}
	devTextAt(screen, label, r.X+devPanelPad, r.Y+(devMobileHeader-devFaceHeight())/devCenterDiv, devFg)
	devTextAt(screen, "X", r.X+r.W-devMobileHeader/2-devTextWidth("X")/2, r.Y+(devMobileHeader-devFaceHeight())/devCenterDiv, devFg)
	if s.dev.mobileExpanded {
		s.drawDevContent(screen)
		s.drawDevScrollbar(screen)
	}
}

// devClosePad is the pointer slop around the X and Picker labels. The F12
// hint draws between them but takes no hit, so a press aimed at Picker or
// the hint cannot close the dock.
const devClosePad = 6.0

// devHeaderRight places the header buttons from the right edge: the Picker
// toggle, the F12 hint, and the X key. Only the X key closes the dock; the
// F12 label is a hint with dead space on both sides. The picker hit covers
// only its own label, so the gaps around F12 do nothing. It returns the
// close rect and the picker text origin so the draw and the hits agree.
func devHeaderRight(r devRect) (cls devRect, px float64) {
	wX := devTextWidth("X")
	wPicker := devTextWidth("Picker")
	xX := r.X + r.W - devPanelPad - wX
	fx := xX - devHeaderGap - devTextWidth("F12")
	px = fx - devHeaderGap - wPicker
	cls = devRect{X: xX - devClosePad, Y: r.Y, W: wX + devClosePad*2, H: devHeaderH}

	if cls.X < r.X {
		cls.W -= r.X - cls.X
		cls.X = r.X
	}

	if cls.X+cls.W > r.X+r.W {
		cls.W = r.X + r.W - cls.X
	}

	return cls, px
}

// devHeaderF12X is the draw origin of the F12 hint. It takes no hit.
func devHeaderF12X(r devRect) float64 {
	wX := devTextWidth("X")
	xX := r.X + r.W - devPanelPad - wX

	return xX - devHeaderGap - devTextWidth("F12")
}

// devPickerRect is the tight hit around the Picker label drawn at px.
func devPickerRect(px, panelY float64) devRect {
	return devRect{X: px - devClosePad, Y: panelY, W: devTextWidth("Picker") + devClosePad*2, H: devHeaderH}
}

// devHover records the box under the cursor without calling the page.
func (s *shell) devHover(px, py float64) {
	box, ok := s.boxAt(px, py)
	s.dev.hovered, s.dev.haveHov = box, ok
}

// devBoxAt returns the last box that contains x, y, the innermost element.
func devBoxAt(boxes []layout.Box, x, y float64) (layout.Box, bool) {
	var found layout.Box
	ok := false

	for _, b := range boxes {
		if x < b.X || y < b.Y || x > b.X+b.W || y > b.Y+b.H {
			continue
		}

		found = b
		ok = true
	}

	return found, ok
}

// devSameBox reports whether two picks name the same element.
func devSameBox(a, b layout.Box) bool {
	if a.ID != "" || b.ID != "" {
		return a.ID == b.ID
	}

	return a.Tag == b.Tag && a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}

// devPanelPress runs the action of the hit under the pointer. A click on
// empty panel space does nothing, so a pinned box stays pinned.
func (s *shell) devPanelPress(x, y int) {
	for _, hit := range s.dev.hits {
		if devInRect(hit.rect, float64(x), float64(y)) {
			s.devRun(hit)

			return
		}
	}
}

// devOpHitText is the displayed row for one operation pick.
func (s *shell) devOpHitText(arg int) string {
	for _, row := range s.devOpsRows() {
		if row.act == devActOp && row.arg == arg {
			return row.line.plain()
		}
	}

	return ""
}

// devRun applies one panel hit. An Ops row pick writes its text to the
// OS clipboard; the write always follows that explicit click, and there
// is no setting to turn it off, so keep the dock closed for sensitive
// output. See documentation/devtools.md and security.md.
func (s *shell) devRun(hit devHit) {
	s.dev.panelCache.valid = false
	switch hit.act {
	case devActPause:
		s.dev.paused = !s.dev.paused
	case devActRefresh:
		switch s.dev.refreshGap {
		case devPanelFastGap:
			s.dev.refreshGap = devPanelRefreshGap
		case devPanelSlowGap:
			s.dev.refreshGap = devPanelFastGap
		default:
			s.dev.refreshGap = devPanelSlowGap
		}
	case devActMobileExpand:
		if s.dev.mobile {
			s.dev.mobileExpanded = !s.dev.mobileExpanded
			s.dev.scroll = 0
		}
	case devActTab:
		s.dev.tab = devTabAt(hit.arg)
		s.dev.scroll = 0
	case devActOpsToggle:
		s.dev.ops = !s.dev.ops
	case devActJSON:
		if s.dev.collapsed == nil {
			s.dev.collapsed = map[string]bool{}
		}

		s.dev.collapsed[hit.key] = !s.dev.collapsed[hit.key]
	case devActOp:
		text := s.devOpHitText(hit.arg)
		s.dev.opPin = false
		if s.dev.haveOp && s.dev.opPick == hit.arg {
			s.dev.haveOp = false
			if text != "" {
				clipboard.Write(text)
			}

			return
		}

		s.dev.opPick, s.dev.haveOp = hit.arg, true
		if text != "" {
			clipboard.Write(text)
		}
	case devActPicker:
		s.devSetPicker(!s.devPickerOn())
	case devActClose:
		_ = s.devToggle()
	}
}

// devResize follows the pointer with the dock's left edge. The page viewport
// follows on the next update, throttled like a window drag.
func (s *shell) devResize(x int) {
	s.moving = true

	w := float64(s.screenW - x)
	if w < devDockMin {
		w = devDockMin
	}

	if w > float64(s.screenW) {
		w = float64(s.screenW)
	}

	s.dev.dockW = w
}

// devScrollLines is the lines one wheel notch moves.
const devScrollLines = 3.0

// devWheel routes the wheel to the dock while the pointer is over it. It
// reports whether the dock ate the wheel, so the page below stays put.
func (s *shell) devWheel() bool {
	if !s.dev.on || !devInRect(s.dev.panel, float64(s.cursorX), float64(s.cursorY)) {
		return false
	}

	_, wheelY := ebiten.Wheel()
	s.dev.scroll -= wheelY * devLineHeight() * devScrollLines
	s.devClampScroll()

	return true
}

// devClampScroll keeps the dock content inside its viewport.
func (s *shell) devClampScroll() {
	max := float64(len(s.dev.rows))*devLineHeight() - s.dev.content.H
	if max < 0 {
		max = 0
	}

	if s.dev.scroll > max {
		s.dev.scroll = max
	}

	if s.dev.scroll < 0 {
		s.dev.scroll = 0
	}
}

// devTouch routes finger input while the devtools dock is open. Lifts
// without a drag fire the panel action; drags scroll the dock; every other
// finger joins the touch pipeline for the content beside the dock.
func (s *shell) devTouch(frameW, frameH int) error {
	var all []touchPos
	for _, id := range ebiten.AppendTouchIDs(nil) {
		x, y := ebiten.TouchPosition(id)
		y -= s.titlebarInset()
		if y < 0 {
			continue
		}
		all = append(all, touchPos{id: id, x: x, y: y})
	}
	content, taps, dy := devPanelStep(all, s.dev.panel, &s.dev.panelTrack)
	if dy != 0 {
		s.dev.scroll += float64(dy)
		s.devClampScroll()
	}
	for _, p := range taps {
		s.devPanelPress(p.x, p.y)
	}
	return s.applyTouches(content, false, frameW, frameH)
}

// devPanelStep sorts panel fingers from content fingers. It returns the
// content fingers, taps that lifted still, and upward scroll pixels.
// A zero-size panel owns nothing. A tap fires at the release point, and any
// move past the slop on either axis turns the finger into a scroll so a
// slide off a button never acts on the landing point.
func devPanelStep(all []touchPos, panel devRect, tr *panelTrack) (content, taps []touchPos, dy int) {
	live := panel.W > 0 && panel.H > 0
	next := make([]panelFinger, 0, len(tr.fingers)+1)
	for _, p := range all {
		if !live || !devInRect(panel, float64(p.x), float64(p.y)) {
			content = append(content, p)
			continue
		}
		i := panelFind(tr.fingers, p.id)
		if i < 0 {
			next = append(next, panelFinger{id: p.id, x: p.x, y: p.y, lastX: p.x, lastY: p.y})
			continue
		}
		f := tr.fingers[i]
		if absInt(p.x-f.x) > touchSlop || absInt(p.y-f.y) > touchSlop {
			f.drag = true
		}
		if f.drag {
			dy += f.lastY - p.y
		}
		f.lastX = p.x
		f.lastY = p.y
		next = append(next, f)
	}
	var ids []ebiten.TouchID
	for _, p := range all {
		ids = append(ids, p.id)
	}
	for _, f := range tr.fingers {
		if !hasTouch(ids, f.id) && !f.drag {
			taps = append(taps, touchPos{id: f.id, x: f.lastX, y: f.lastY})
		}
	}
	tr.fingers = next
	return content, taps, dy
}

// devTab names the panel's top-level views.
type devTab int

const (
	devTabFrame devTab = iota
	devTabElements
	devTabOps
)

// devTabNames are the labels in the tab strip, in order. The Frame tab is
// first, so it is the one the dock opens on.
var devTabNames = [...]string{"Frame", "Elements", "Ops"}

// devTabAt clamps an index to a known tab.
func devTabAt(i int) devTab {
	if i < 0 {
		return devTabFrame
	}

	if i >= len(devTabNames) {
		return devTabOps
	}

	return devTab(i)
}

// devTabHint is the footer hint for one tab.
func devTabHint(tab devTab) string {
	switch tab {
	case devTabFrame:
		return "frame and pipeline counters"
	case devTabOps:
		return "o toggles the outlines"
	default:
		return "click a box in the page to pin it"
	}
}

// devTabUnderline is the active tab underline thickness and rise.
const devTabUnderline = 2.0

// drawDevTabs paints the tab strip. The active tab gets the blue underline.
func (s *shell) drawDevTabs(screen *ebiten.Image) {
	r := s.dev.panel
	y := r.Y + devHeaderH
	vector.FillRect(screen, float32(r.X), float32(y), float32(r.W), devTabsH, devHeaderBg, false)
	vector.FillRect(screen, float32(r.X), float32(y+devTabsH-1), float32(r.W), 1, devBorder, false)

	x := r.X + devPanelPad
	top := y + (devTabsH-devFaceHeight())/devCenterDiv

	for i, name := range devTabNames {
		ink := devDim
		if devTab(i) == s.dev.tab {
			ink = devFg
		}

		devTextAt(screen, name, x, top, ink)

		w, _ := text.Measure(name, badgeFace, 0)
		if devTab(i) == s.dev.tab {
			vector.FillRect(screen, float32(x), float32(y+devTabsH-devTabUnderline), float32(w), devTabUnderline, devAccent, false)
		}

		x += w + devTabPad
	}
}
