package window

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"strings"
	"testing"
)

// TestDevToolsToggleSwallowsF12 checks the Phase 3 exit: F12 flips the
// overlay and the page's key handler never sees the press or the release.
func TestDevToolsToggleSwallowsF12(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	swallow, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 press swallow = %v, err = %v", swallow, err)
	}

	if d.on {
		t.Fatal("f12 did not close the overlay")
	}
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}

	if d.hovers != 1 {
		t.Fatalf("close hovers = %d, want 1", d.hovers)
	}

	if down, up := s.pageKeyStep(ebiten.KeyF12, true); down || up {
		t.Fatal("the page saw the f12 press")
	}

	swallow, err = s.devKeyStep(ebiten.KeyF12, false, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 release swallow = %v, err = %v", swallow, err)
	}

	if down, up := s.pageKeyStep(ebiten.KeyF12, false); down || up {
		t.Fatal("the page saw the f12 release")
	}
}

// TestDevToolsToggleIgnoresRepeatPulse checks a server repeat reported as a
// release and a press in the next frame does not toggle twice.
func TestDevToolsToggleIgnoresRepeatPulse(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if _, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if d.on {
		t.Fatal("the first press did not toggle")
	}

	if _, err := s.devKeyStep(ebiten.KeyF12, false, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if _, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if d.on {
		t.Fatal("the repeat pulse toggled again")
	}
}

type devCursorScreen struct {
	*devScreen
	shape host.Shape
}

func (d *devCursorScreen) CursorShape() host.Shape { return d.shape }

func TestDevToolsNormalHoverUpdatesCursor(t *testing.T) {
	d := &devCursorScreen{devScreen: newDevScreen(), shape: host.ShapePointer}
	s := newDevShell(d.devScreen)
	s.app = d
	s.devLayout()
	// The routing under test runs after startup, so commit a frame: the
	// first update defers the page hover until the first paint.
	s.frameCommit()
	s.cursor = ebiten.CursorShapeText
	s.setCursor = func(ebiten.CursorShapeType) {}
	if err := s.devPointer(20, 20, 20, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapePointer || d.hovers != 1 {
		t.Fatalf("normal content hover kept stale cursor: shape=%v hovers=%d", s.cursor, d.hovers)
	}

	if err := s.devPointer(int(s.dev.panel.X)+50, 10, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatal("dock chrome retained the page cursor")
	}
	if err := s.devPointer(int(s.dev.panel.X), 100, 0, 0); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeEWResize {
		t.Fatal("dock resize edge did not show its resize cursor")
	}
	s.devSetPicker(true)
	if err := s.devPointer(20, 20, 20, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault || !s.dev.haveHov || d.hovers != 1 {
		t.Fatal("picker did not use its own hover without forwarding to the page")
	}
	s.devSetPicker(false)
	d.shape = host.ShapeDefault
	if err := s.devPointer(21, 20, 21, 20); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault || d.hovers != 2 {
		t.Fatal("disabling picker did not restore normal page hover")
	}
}

// devTestDisplay is a small display list with one fill and one text run.
func devTestDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 100, H: 50},
			{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, Text: "hi"},
			{Kind: layout.DisplayOpNoop, X: 0, Y: 0, W: 1, H: 1},
		},
		Order:          []int{0, 1, 2},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}

// TestDevDockOnTheRight checks the dock covers the right edge at full
// height and clamps a width wider than the window.
func TestDevDockOnTheRight(t *testing.T) {
	t.Parallel()

	if got := devDockRect(800, 600, 340); got != (devRect{X: 460, Y: 0, W: 340, H: 600}) {
		t.Fatalf("dock = %+v, want the right 340 columns", got)
	}

	if got := devDockRect(300, 200, 340); got != (devRect{X: 0, Y: 0, W: 300, H: 200}) {
		t.Fatalf("clamped dock = %+v, want the whole screen", got)
	}
}

// TestDevBadgeClearsTheDock checks the fallback badge shifts left of the
// dock instead of hiding under it.
func TestDevBadgeClearsTheDock(t *testing.T) {
	t.Parallel()

	dock := devDockRect(800, 600, devDockWidth)
	x, y, w, h := badgeRectInset(800, badgeLabel, dock.W)
	badge := devRect{X: x, Y: y, W: w, H: h}

	if badge.X+badge.W > dock.X {
		t.Fatalf("badge %+v overlaps dock %+v", badge, dock)
	}
}

// TestDevDockResponsiveWidth checks the default dock takes its full 340 px
// only on a roomy screen, gives a narrow window a share instead, and leaves
// the page at least devDockPageMin wide wherever the screen has room.
func TestDevDockResponsiveWidth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		screenW int
		dock    float64
	}{
		{1500, devDockWidth},
		{1180, devDockWidth},
		{1100, 308},
		{1000, 280},
		{800, 224},
		{600, devDockMin},
		{420, devDockMin},
	}

	for _, c := range cases {
		s := newDevShell(newDevScreen())
		s.screenW, s.screenH = c.screenW, 900
		s.dev.on = true

		got := s.devDockWidth()
		if got != c.dock {
			t.Fatalf("%d px screen: dock = %v, want %v", c.screenW, got, c.dock)
		}

		if got < devDockMin && c.screenW >= int(devDockMin) {
			t.Fatalf("%d px screen: dock = %v, under the %v minimum", c.screenW, got, devDockMin)
		}

		if got > float64(c.screenW) {
			t.Fatalf("%d px screen: dock = %v, wider than the screen", c.screenW, got)
		}

		pageW, pageH := s.viewSize()
		if pageW != c.screenW-int(got) || pageH != 900 {
			t.Fatalf("%d px screen: page = %dx%d, want %dx900",
				c.screenW, pageW, pageH, c.screenW-int(got))
		}

		if c.screenW >= int(devDockMin+devDockPageMin) && pageW < int(devDockPageMin) {
			t.Fatalf("%d px screen: page = %d, want at least %v",
				c.screenW, pageW, devDockPageMin)
		}

		panel := devDockRect(s.screenW, s.screenH, got)
		if panel.X < 0 || panel.W != got || panel.X+panel.W != float64(c.screenW) {
			t.Fatalf("%d px screen: panel = %+v", c.screenW, panel)
		}
	}
}

// TestDevDockWidthClamps checks the stored drag width never goes under the
// minimum or squeezes the page under 320 px on a screen with room for it.
func TestDevDockWidthClamps(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if got := s.devDockWidth(); got != devDockWidth {
		t.Fatalf("default width = %v, want %v", got, devDockWidth)
	}

	s.dev.dockW = 100
	if got := s.devDockWidth(); got != devDockMin {
		t.Fatalf("narrow width = %v, want %v", got, devDockMin)
	}

	s.dev.dockW = 9000
	if got := s.devDockWidth(); got != float64(s.screenW)-devDockPageMin {
		t.Fatalf("wide width = %v, want the page at %v", got, devDockPageMin)
	}

	s.screenW = 800
	s.dev.dockW = 9000
	if got := s.devDockWidth(); got != float64(s.screenW)-devDockPageMin {
		t.Fatalf("small screen drag = %v, want the page at %v", got, devDockPageMin)
	}

	s.dev.dockW = 0
	s.screenW = 200
	if got := s.devDockWidth(); got != float64(s.screenW) {
		t.Fatalf("screen under the minimum = %v, want it filled", got)
	}
}

// TestDevDrawOverlaySmoke draws the overlay into an offscreen image in
// normal and stretched modes. The stroke hook records the computed screen
// rects, so the test finds each outline without reading pixels back.
func TestDevDrawOverlaySmoke(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devTestDisplay()
	s.dev.ops = true
	s.devSetPicker(true)
	s.dev.haveHov = true
	s.dev.hovered = d.boxes[0]
	s.dev.havePin = true
	s.dev.pinned = d.boxes[0]

	var strokes []devRect
	s.dev.stroke = func(_ *ebiten.Image, r devRect, _ color.RGBA) {
		strokes = append(strokes, r)
	}

	img := ebiten.NewImage(1920, 1080)
	s.drawDevTools(img)

	box := d.boxes[0]
	want := s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})
	// The ops view strokes the display's two painted ops first, then the
	// hover and pin boxes.
	if len(strokes) != 4 || strokes[2] != want || strokes[3] != want {
		t.Fatalf("stretched outlines = %v, want the boxes %v last", strokes, want)
	}

	s.screenW, s.screenH = 800, 600
	s.drawDevTools(img)

	want = s.devScreen(devRect{X: box.X, Y: box.Y, W: box.W, H: box.H})
	if len(strokes) != 8 || strokes[6] != want || strokes[7] != want {
		t.Fatalf("normal outlines = %v, want the boxes %v last", strokes[4:], want)
	}

	// A page with no boxes still draws its dock on the right edge.
	d.boxes = nil
	s.dev.haveHov, s.dev.havePin = false, false
	s.drawDevTools(img)

	if s.dev.panel.W <= 0 || s.dev.panel.H <= 0 {
		t.Fatalf("panel = %+v, want a drawn panel", s.dev.panel)
	}

	if s.dev.panel.X+s.dev.panel.W != float64(s.screenW) || s.dev.panel.Y != 0 {
		t.Fatalf("panel = %+v, want the right dock", s.dev.panel)
	}
}

// TestDevElementRowsJSON checks the pinned box renders as valid JSON with
// the element fields.
func TestDevElementRowsJSON(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devTestDisplay()
	s.dev.havePin, s.dev.pinned = true, d.boxes[0]

	joined := devRowsText(s.devElementRows())
	if !strings.Contains(joined, `"tag": "div"`) || !strings.Contains(joined, `"id": "known"`) {
		t.Fatalf("element rows = %q", joined)
	}

	start := strings.Index(joined, "{")
	end := strings.LastIndex(joined, "}") + 1
	if start < 0 || end <= start || !json.Valid([]byte(joined[start:end])) {
		t.Fatalf("element JSON is not valid: %q", joined)
	}
}

// TestDevElementRowsTarget checks the pinned box wins over the hovered one
// and an empty pick explains how to select.
func TestDevElementRowsTarget(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if text := devRowsText(s.devElementRows()); !strings.Contains(text, "Select an element") {
		t.Fatalf("empty rows = %q", text)
	}

	s.dev.haveHov, s.dev.hovered = true, layout.Box{Tag: "span", ID: "hover"}
	if text := devRowsText(s.devElementRows()); !strings.Contains(text, "hovering  span#hover") {
		t.Fatalf("hover rows = %q", text)
	}

	s.dev.havePin, s.dev.pinned = true, layout.Box{Tag: "div", ID: "pin"}
	if text := devRowsText(s.devElementRows()); !strings.Contains(text, "pinned  div#pin") {
		t.Fatalf("pin rows = %q", text)
	}
}

// TestDevScreenRectNormal checks the scroll subtraction of the one
// conversion function.
func TestDevScreenRectNormal(t *testing.T) {
	t.Parallel()

	got := devScreenRect(devRect{X: 16, Y: 32, W: 120, H: 48}, 10, 20, 1, false, 800, 600, 800, 600)
	want := devRect{X: 6, Y: 12, W: 120, H: 48}

	if got != want {
		t.Fatalf("screen rect = %+v, want %+v", got, want)
	}
}

// TestDevScreenRectZoom checks the pinch zoom multiplies before the scroll
// subtraction, matching the draw path.
func TestDevScreenRectZoom(t *testing.T) {
	t.Parallel()

	got := devScreenRect(devRect{X: 16, Y: 32, W: 120, H: 48}, 10, 20, 1.5, false, 800, 600, 800, 600)
	want := devRect{X: 14, Y: 28, W: 180, H: 72}

	if got != want {
		t.Fatalf("screen rect = %+v, want %+v", got, want)
	}
}

// TestDevScreenRectStretched checks the scale factors drawReplayScaled uses.
func TestDevScreenRectStretched(t *testing.T) {
	t.Parallel()

	got := devScreenRect(devRect{X: 16, Y: 32, W: 120, H: 48}, 0, 0, 1, true, 400, 300, 800, 600)
	want := devRect{X: 32, Y: 64, W: 240, H: 96}

	if got != want {
		t.Fatalf("screen rect = %+v, want %+v", got, want)
	}
}

// TestDevHeaderHitsTouch checks only the X key closes the dock. The Picker
// hit covers only its label, the F12 hint takes no hit, and dead space sits
// between them so a press aimed at Picker or the hint cannot close.
func TestDevHeaderHitsTouch(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.devLayout()

	var picker, closer *devHit
	for i := range s.dev.hits {
		switch s.dev.hits[i].act {
		case devActPicker:
			hit := s.dev.hits[i]
			picker = &hit
		case devActClose:
			hit := s.dev.hits[i]
			closer = &hit
		}
	}

	if picker == nil || closer == nil {
		t.Fatalf("hits = %+v, want picker and close", s.dev.hits)
	}

	if picker.rect.X+picker.rect.W >= closer.rect.X {
		t.Fatalf("picker [%v %v] overlaps close [%v %v]",
			picker.rect.X, picker.rect.X+picker.rect.W, closer.rect.X, closer.rect.X+closer.rect.W)
	}

	f12X := devHeaderF12X(s.dev.panel)
	f12Center := f12X + devTextWidth("F12")/2
	if devInRect(closer.rect, f12Center, closer.rect.Y+closer.rect.H/2) {
		t.Fatalf("F12 center %v in close rect %+v", f12Center, closer.rect)
	}
	if devInRect(picker.rect, f12Center, closer.rect.Y+closer.rect.H/2) {
		t.Fatalf("F12 center %v in picker rect %+v", f12Center, picker.rect)
	}

	s.devPanelPress(int(f12Center), int(closer.rect.Y+closer.rect.H/2))
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if !s.dev.on || !d.on {
		t.Fatal("pressing the F12 hint closed the dock")
	}

	xX := s.dev.panel.X + s.dev.panel.W - devPanelPad - devTextWidth("X")
	s.devPanelPress(int(xX+devTextWidth("X")/2), int(closer.rect.Y+closer.rect.H/2))
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.on {
		t.Fatal("pressing X did not close the dock")
	}
}

// TestDevOwnsPointBlocksTitlebarDrag checks the dock owns its header and
// resize edge, so a custom-chrome draggable strip cannot start a window
// drag over the close key.
func TestDevOwnsPointBlocksTitlebarDrag(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.screenW, s.screenH = 1500, 900
	s.dev.on = true
	s.devLayout()

	cls, _ := devHeaderRight(s.dev.panel)
	xCenter := s.dev.panel.X + s.dev.panel.W - devPanelPad - devTextWidth("X")/2
	yCenter := s.dev.panel.Y + devHeaderH/2
	if !s.devOwnsPoint(int(xCenter), int(yCenter)) {
		t.Fatalf("close key %v,%v not owned by dock %+v", xCenter, yCenter, s.dev.panel)
	}
	_ = cls

	edgeX := int(s.dev.panel.X - 1)
	if !s.devOwnsPoint(edgeX, int(yCenter)) {
		t.Fatalf("resize edge %v not owned by dock %+v", edgeX, s.dev.panel)
	}

	pageX := int(s.dev.panel.X) - int(devDockGrabWidth) - 10
	if s.devOwnsPoint(pageX, int(yCenter)) {
		t.Fatalf("page titlebar %v owned by dock %+v", pageX, s.dev.panel)
	}

	s.dev.on = false
	if s.devOwnsPoint(int(xCenter), int(yCenter)) {
		t.Fatal("closed dock still owns the header point")
	}
}

// TestDevLayoutPlacesTheDock checks the dock hugs the right edge, the
// content rect has room, and every tab gets a hit.
func TestDevLayoutPlacesTheDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.devLayout()

	if s.dev.panel.X != float64(s.screenW)-devDockWidth || s.dev.panel.H != float64(s.screenH) {
		t.Fatalf("panel = %+v", s.dev.panel)
	}

	if s.dev.content.W <= 0 || s.dev.content.H <= 0 {
		t.Fatalf("content = %+v", s.dev.content)
	}

	tabs := 0

	for _, hit := range s.dev.hits {
		if hit.act == devActTab {
			tabs++
		}
	}

	if tabs != len(devTabNames) {
		t.Fatalf("tab hits = %d, want %d", tabs, len(devTabNames))
	}
}

// TestDevLayoutTabContent checks each tab builds its own rows.
func TestDevLayoutTabContent(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.dev.tab = devTabElements
	if text := devRowsText(s.devTabRows()); !strings.Contains(text, "Select an element") {
		t.Fatalf("elements tab = %q", text)
	}

	s.dev.tab = devTabFrame
	if text := devRowsText(s.devTabRows()); !strings.Contains(text, "PIPELINE") {
		t.Fatalf("frame tab = %q", text)
	}

	s.display = devTestDisplay()
	s.dev.tab = devTabOps
	if text := devRowsText(s.devTabRows()); !strings.Contains(text, "PAINT ORDER") {
		t.Fatalf("ops tab = %q", text)
	}
}

func TestMobileMetricsLayoutAndCollapse(t *testing.T) {
	for _, size := range [][2]int{{420, 934}, {934, 420}, {200, 60}} {
		s := newDevShell(newDevScreen())
		s.screenW, s.screenH = size[0], size[1]
		s.dev.on, s.dev.mobile = true, true
		s.devLayout()
		collapsed := s.dev.panel
		width, height := s.viewOf(size[0], size[1])
		if width != size[0] || height < 1 || collapsed.X != 0 || collapsed.W != float64(size[0]) {
			t.Fatalf("mobile viewport narrowed: %dx%d panel=%+v", width, height, collapsed)
		}
		if collapsed.Y < 0 || collapsed.Y+collapsed.H > float64(size[1]) {
			t.Fatalf("panel outside screen: %+v", collapsed)
		}
		if size[1] > 80 && collapsed.H != devMobileHeader {
			t.Fatalf("collapsed touch target is not 48dp: %+v", collapsed)
		}
		s.devRun(devHit{act: devActMobileExpand})
		s.devLayout()
		if !s.dev.mobileExpanded || s.dev.panel.H < collapsed.H {
			t.Fatal("mobile panel did not expand")
		}
		if size[1] > 120 && s.dev.panel.H > float64(size[1])*devMobileFraction {
			t.Fatalf("panel occupies over 40%%: %+v", s.dev.panel)
		}
		if s.dockClipX(1000) != 1000 {
			t.Fatal("mobile panel incorrectly clipped page width")
		}
		s.devRun(devHit{act: devActMobileExpand})
		s.devLayout()
		if s.dev.panel != collapsed {
			t.Fatalf("collapse did not restore geometry: %+v", s.dev.panel)
		}
	}
}

func TestMobileMetricsTouchHeaderDoesNotReachPage(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.screenW, s.screenH = 420, 934
	s.dev.on, s.dev.mobile = true, true
	s.devLayout()
	touch := touchPos{id: 1, x: 20, y: int(s.dev.panel.Y) + 20}
	content, taps, _ := devPanelStep([]touchPos{touch}, s.dev.panel, &s.dev.panelTrack)
	if len(content) != 0 || len(taps) != 0 {
		t.Fatal("metrics press leaked to application")
	}
	_, taps, _ = devPanelStep(nil, s.dev.panel, &s.dev.panelTrack)
	if len(taps) != 1 {
		t.Fatal("header lift did not produce a tap")
	}
	s.devPanelPress(taps[0].x, taps[0].y)
	if !s.dev.mobileExpanded {
		t.Fatal("header did not expand panel")
	}
	content, _, _ = devPanelStep([]touchPos{{id: 2, x: 20, y: 20}}, s.dev.panel, &s.dev.panelTrack)
	if len(content) != 1 {
		t.Fatal("application touch was swallowed outside mobile panel")
	}
}

func TestDesktopMetricsLayoutRemainsSideDock(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.screenW, s.screenH = 420, 934
	s.dev.on = true
	s.dev.mobile = false
	s.devLayout()
	// 420 px narrows the default dock from 340 to the 220 px minimum, so
	// the page still gets 200 px beside it.
	if s.dev.panel != (devRect{X: 200, W: 220, H: 934}) {
		t.Fatalf("desktop dock changed: %+v", s.dev.panel)
	}
	if width, height := s.viewOf(420, 934); width != 200 || height != 934 {
		t.Fatalf("desktop viewport changed: %dx%d", width, height)
	}
	s.devRun(devHit{act: devActMobileExpand})
	if s.dev.mobileExpanded {
		t.Fatal("mobile action changed desktop state")
	}
}

func TestMobileMetricsCloseTouchRestoresViewport(t *testing.T) {
	for _, expanded := range []bool{false, true} {
		d := newDevScreen()
		s := newDevShell(d)
		s.screenW, s.screenH = 420, 934
		s.dev.mobile, s.dev.mobileExpanded = true, expanded
		s.perfOption = true
		s.devLayout()
		close := s.dev.hits[0]
		if close.act != devActClose || close.rect.W != 48 || close.rect.H != 48 {
			t.Fatalf("close button lacks 48dp target: %+v", close)
		}
		touch := touchPos{id: 1, x: int(close.rect.X) + 24, y: int(close.rect.Y) + 24}
		content, _, _ := devPanelStep([]touchPos{touch}, s.dev.panel, &s.dev.panelTrack)
		if len(content) != 0 {
			t.Fatal("close press reached page")
		}
		_, taps, _ := devPanelStep(nil, s.dev.panel, &s.dev.panelTrack)
		if len(taps) != 1 {
			t.Fatal("close touch did not produce tap")
		}
		s.devPanelPress(taps[0].x, taps[0].y)
		if err := s.devSync(); err != nil {
			t.Fatal(err)
		}
		if s.dev.on || d.DevTools() {
			t.Fatal("X left devtools open")
		}
		if !s.perf {
			t.Fatal("closing devtools disabled opted-in metrics")
		}
		if w, h := s.viewSize(); w != 420 || h != 934 {
			t.Fatalf("viewport not restored: %dx%d", w, h)
		}
		if d.clicked != 0 || len(s.dev.panelTrack.fingers) != 0 {
			t.Fatal("close leaked input or retained fingers")
		}
	}
}

// TestDevThumbWinsOverDockEdge checks the overlapping strip between the page
// scrollbar thumb and the dock resize edge scrolls instead of resizing. The
// thumb occupies the last 8 px of the page view and the edge straddles the
// dock boundary by 4 px on each side, so the right half of the thumb is also
// the edge. pointerScrollbar must consume that press before devPointer can
// start a resize.
func TestDevThumbWinsOverDockEdge(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.boxes = []layout.Box{
		{ID: "known", Tag: "div", X: 10, Y: 10, W: 100, H: 40},
		{ID: "tall", Tag: "div", X: 0, Y: 0, W: 100, H: 2000},
	}
	s := newDevShell(d)
	s.screenW, s.screenH = 800, 600
	s.dev.on = true
	s.devLayout()

	viewW, viewH := s.viewSize()
	// Match the page to the docked viewport so scrollbars are enabled:
	// a stretched page disables them and the thumb never shows.
	d.SetSize(viewW, viewH)
	s.devLayout()
	viewW, viewH = s.viewSize()
	contentW, contentH := s.contentSize()
	if !scrollbarVisible(contentH, viewH) {
		t.Fatalf("content %dx%d view %dx%d shows no vertical thumb", contentW, contentH, viewW, viewH)
	}

	panelX := int(s.dev.panel.X)
	thumbX := viewW - 2
	if !devOnEdge(s.dev.panel, thumbX, 10) {
		t.Fatalf("x=%d not on dock edge %+v", thumbX, s.dev.panel)
	}
	if _, _, ok := s.scrollbarGrab(thumbX, 10); !ok {
		t.Fatalf("x=%d y=10 not on scrollbar thumb (view %dx%d)", thumbX, viewW, viewH)
	}

	if !s.pointerScrollbar(thumbX, 10, true, true) {
		t.Fatal("thumb press did not consume the overlapping edge press")
	}
	if s.dev.resizing {
		t.Fatal("thumb press started a dock resize")
	}
	if s.dragAxis == dragNone {
		t.Fatal("thumb press did not start a scrollbar drag")
	}
	_ = panelX
}

// TestDevPanelSlideOffCancelsTap checks a touch that lands on a dock button
// but slides past the slop never fires at its landing point.
func TestDevPanelSlideOffCancelsTap(t *testing.T) {
	t.Parallel()

	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	land := touchPos{id: 7, x: 1700, y: 100}
	if _, _, _ = devPanelStep([]touchPos{land}, panel, tr); len(tr.fingers) != 1 {
		t.Fatal("landing finger not tracked")
	}
	moved := touchPos{id: 7, x: 1700 + touchSlop + 2, y: 100}
	if _, _, _ = devPanelStep([]touchPos{moved}, panel, tr); !tr.fingers[0].drag {
		t.Fatal("horizontal slide past slop did not mark drag")
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 0 {
		t.Fatalf("slid finger tapped: %v", taps)
	}
}

// TestDevPanelTapFiresAtRelease checks a tap within the slop fires at its
// release point instead of its landing point.
func TestDevPanelTapFiresAtRelease(t *testing.T) {
	t.Parallel()

	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	if _, _, _ = devPanelStep([]touchPos{{id: 9, x: 1700, y: 100}}, panel, tr); len(tr.fingers) != 1 {
		t.Fatal("landing finger not tracked")
	}
	if _, _, _ = devPanelStep([]touchPos{{id: 9, x: 1702, y: 101}}, panel, tr); tr.fingers[0].drag {
		t.Fatal("move within slop marked drag")
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 1 {
		t.Fatalf("taps = %v, want one", taps)
	}
	if taps[0].x != 1702 || taps[0].y != 101 {
		t.Fatalf("tap = %v, want the release point 1702,101", taps[0])
	}
}

// TestDevToolsLeavesPageAlone is the plan's guard: the same page with the
// overlay off and on keeps its PNG bytes and generation, and the overlay
// never enters the display list or the box list.
func TestDevToolsLeavesPageAlone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := page.New(page.Config{
		HTML:   `<div id="a" style="background:#1a56db;border-radius:8px">hello</div>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	display := p.Display()
	if display == nil {
		t.Fatal("the page fell back to a bitmap")
	}

	before := p.PNG()
	if len(before) == 0 {
		t.Fatal("no PNG before the overlay")
	}

	gen := p.Generation()
	ops := len(display.Ops)
	boxes := len(p.Boxes())

	s := &shell{app: p, ctx: ctx, screenW: 320, screenH: 200}

	off := ebiten.NewImage(320, 200)
	s.Draw(off)

	p.SetDevTools(true)
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}

	s.devRefresh()
	s.display = display
	s.dev.ops = true

	if boxes > 0 {
		s.dev.haveHov, s.dev.hovered = true, p.Boxes()[0]
		s.dev.havePin, s.dev.pinned = true, p.Boxes()[0]
	}

	on := ebiten.NewImage(320, 200)
	s.Draw(on)

	if p.Generation() != gen {
		t.Fatalf("generation = %d, want %d", p.Generation(), gen)
	}

	if !bytes.Equal(before, p.PNG()) {
		t.Fatal("the overlay changed the page PNG")
	}

	if len(p.Display().Ops) != ops || len(p.Boxes()) != boxes {
		t.Fatalf("ops = %d, boxes = %d, want %d and %d",
			len(p.Display().Ops), len(p.Boxes()), ops, boxes)
	}
}

// TestDevPanelTabsSwitch checks a click on a tab hit changes the active tab
// and clears the content scroll.
func TestDevPanelTabsSwitch(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.scroll = 40

	s.devRun(devHit{act: devActTab, arg: int(devTabFrame)})
	if s.dev.tab != devTabFrame {
		t.Fatalf("tab = %v, want Frame", s.dev.tab)
	}

	if s.dev.scroll != 0 {
		t.Fatalf("scroll = %v, want 0 after a tab switch", s.dev.scroll)
	}
}

// TestDevPanelRunActions checks the toggle, JSON, and op rows.
func TestDevPanelRunActions(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devTestDisplay()

	s.devRun(devHit{act: devActOpsToggle})
	if !s.dev.ops {
		t.Fatal("the toggle row did not turn the outlines on")
	}

	s.devRun(devHit{act: devActJSON, key: "rect"})
	if !s.dev.collapsed["rect"] {
		t.Fatal("the JSON row did not collapse rect")
	}

	s.devRun(devHit{act: devActOp, arg: 1})
	if !s.dev.haveOp || s.dev.opPick != 1 {
		t.Fatalf("op pick = %v/%d, want op 1", s.dev.haveOp, s.dev.opPick)
	}

	s.devRun(devHit{act: devActOp, arg: 1})
	if s.dev.haveOp {
		t.Fatal("the same op row did not clear the pick")
	}
}

// TestDevPanelPressHits checks a press lands on the recorded hit rects and
// that empty panel space does nothing.
func TestDevPanelPressHits(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.panel = devRect{X: 100, W: 200, H: 400}
	s.dev.hits = []devHit{{
		rect: devRect{X: 100, Y: 30, W: 200, H: 16},
		act:  devActTab,
		arg:  int(devTabOps),
	}}

	s.devPanelPress(120, 35)
	if s.dev.tab != devTabOps {
		t.Fatalf("tab = %v, want Ops", s.dev.tab)
	}

	s.devPanelPress(120, 300)
	if s.dev.tab != devTabOps {
		t.Fatal("an empty panel click changed the tab")
	}
}

// TestDevCloseHitClosesDock checks the header close key shuts the overlay
// through the page flag.
func TestDevCloseHitClosesDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.panel = devRect{X: 100, W: 400, H: 600}
	cls, _ := devHeaderRight(s.dev.panel)
	s.dev.hits = []devHit{{rect: cls, act: devActClose}}

	s.devPanelPress(int(cls.X+cls.W-4), 10)
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.on {
		t.Fatal("the close key did not shut the dock")
	}
}

// TestDevResizeFollowsThePointer checks the drag width and its clamps.
func TestDevResizeFollowsThePointer(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.devResize(s.screenW - 500)
	if s.dev.dockW != 500 {
		t.Fatalf("dock width = %v, want 500", s.dev.dockW)
	}

	s.devResize(s.screenW - 1)
	if s.dev.dockW != devDockMin {
		t.Fatalf("narrow drag = %v, want %v", s.dev.dockW, devDockMin)
	}
}

// TestDevClampScroll checks the content scroll stays within the rows.
func TestDevClampScroll(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.rows = make([]devRow, 200)
	s.dev.content = devRect{H: 100}

	s.dev.scroll = -5
	s.devClampScroll()

	if s.dev.scroll != 0 {
		t.Fatalf("negative scroll = %v, want 0", s.dev.scroll)
	}

	s.dev.scroll = 1e9
	s.devClampScroll()

	want := float64(200)*devLineHeight() - 100
	if s.dev.scroll != want {
		t.Fatalf("clamped scroll = %v, want %v", s.dev.scroll, want)
	}
}

// TestDevClampScrollShortContent checks a tab shorter than the viewport
// cannot scroll.
func TestDevClampScrollShortContent(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.rows = make([]devRow, 2)
	s.dev.content = devRect{H: 400}
	s.dev.scroll = 30

	s.devClampScroll()

	if s.dev.scroll != 0 {
		t.Fatalf("short content scroll = %v, want 0", s.dev.scroll)
	}
}

// No window or GPU loop runs here: a pointer action must leave the drawn
// viewport unchanged until the next sync, which precedes resize.
func TestDevToolsToggleDefersGeometryUntilSync(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	beforeW, beforeH := s.viewSize()
	if err := s.devToggle(); err != nil {
		t.Fatal(err)
	}
	if d.DevTools() {
		t.Fatal("close request did not reach the page")
	}
	if w, h := s.viewSize(); w != beforeW || h != beforeH || d.hovers != 0 {
		t.Fatal("pointer close changed geometry before the next resize pass")
	}
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.hovers != 1 {
		t.Fatal("sync did not close the dock and clear hover exactly once")
	}
	if w, h := s.viewSize(); w != s.screenW || h != s.screenH {
		t.Fatalf("closed viewport = %dx%d", w, h)
	}
}

// TestDevToolsKeepsF12AvailableAfterPageClose guards the walk hook case: the
// page opens the overlay, then closes it before the first frame. F12 must
// still reopen the dock.
func TestDevToolsKeepsF12AvailableAfterPageClose(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.available, s.dev.on = false, false

	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if !s.dev.on || !s.dev.available {
		t.Fatal("page opt-in did not make the overlay available")
	}

	d.on = false
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || !s.dev.available {
		t.Fatal("page close dropped F12 availability")
	}

	swallow, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 after page close: swallow=%v err=%v", swallow, err)
	}
	if !d.on {
		t.Fatal("f12 did not reopen the overlay")
	}
}

// TestDevTabOpensOnFrame checks the dock starts on the Frame tab, since the
// tab order puts Frame first.
func TestDevTabOpensOnFrame(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if s.dev.tab != devTabFrame {
		t.Fatalf("tab = %v, want Frame", s.dev.tab)
	}

	if devTabNames[0] != "Frame" {
		t.Fatalf("first tab = %q, want Frame", devTabNames[0])
	}
}

// TestDevTabAtClamps checks an out-of-range index lands on a real tab.
func TestDevTabAtClamps(t *testing.T) {
	t.Parallel()

	if got := devTabAt(-1); got != devTabFrame {
		t.Fatalf("devTabAt(-1) = %v, want Frame", got)
	}

	if got := devTabAt(len(devTabNames)); got != devTabOps {
		t.Fatalf("devTabAt(%d) = %v, want Ops", len(devTabNames), got)
	}
}

// TestDevPanelStepTapsOnLift fires a panel tap when a still finger lifts,
// and never for content fingers or a zero-size panel.
func TestDevPanelStepTapsOnLift(t *testing.T) {
	t.Parallel()
	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	all := []touchPos{{id: 1, x: 1700, y: 100}, {id: 2, x: 500, y: 500}}
	content, taps, dy := devPanelStep(all, panel, tr)
	if len(content) != 1 || content[0].id != 2 {
		t.Fatalf("content = %v, want the finger at 500,500", content)
	}
	if len(taps) != 0 || dy != 0 {
		t.Fatalf("landing fired taps = %v dy = %v", taps, dy)
	}
	content, taps, _ = devPanelStep(nil, panel, tr)
	if len(taps) != 1 || taps[0].id != 1 {
		t.Fatalf("lift taps = %v, want the still panel finger", taps)
	}
	if len(content) != 0 {
		t.Fatalf("lift content = %v, want none", content)
	}
	content, taps, _ = devPanelStep(all, devRect{}, &panelTrack{})
	if len(content) != 2 || len(taps) != 0 {
		t.Fatal("a zero panel must not own touches")
	}
}

// TestDevPanelStepDragScrolls scrolls on a dragged panel finger and fires
// no tap for it when it lifts.
func TestDevPanelStepDragScrolls(t *testing.T) {
	t.Parallel()
	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	devPanelStep([]touchPos{{id: 3, x: 1700, y: 500}}, panel, tr)
	_, _, dy := devPanelStep([]touchPos{{id: 3, x: 1700, y: 460}}, panel, tr)
	if dy != 40 {
		t.Fatalf("drag dy = %v, want 40 up", dy)
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 0 {
		t.Fatalf("dragged lift tapped: %v", taps)
	}
}
