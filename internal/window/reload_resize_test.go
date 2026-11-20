package window

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"testing"
	"time"
)

func TestHotReloadPollsOnTheInterval(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 {
		t.Fatalf("polls = %d, want 1", app.polls)
	}

	if app.redraws != 2 {
		t.Fatalf("redraws = %d, want one reload", app.redraws)
	}

	if s.display == nil || s.seq != 2 {
		t.Fatalf("display = %v, seq = %d, want the new generation", s.display, s.seq)
	}
}

func TestHotReloadWaitsForTheInterval(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	s.lastPoll = time.Now()

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 0 {
		t.Fatalf("polls = %d, want 0 inside the interval", app.polls)
	}

	s.lastPoll = time.Now().Add(-time.Second)

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 {
		t.Fatalf("polls = %d, want 1 after the interval", app.polls)
	}
}

func TestReloadResolvesHoverFromTheCursor(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	s.cursorX, s.cursorY = 10, 20

	if err := s.pollReload(); err != nil {
		t.Fatal(err)
	}

	if app.hovers != 1 {
		t.Fatalf("hovers = %d, want 1", app.hovers)
	}

	if app.hoverX != 10 || app.hoverY != 20 {
		t.Fatalf("hover = %v, %v", app.hoverX, app.hoverY)
	}
}

func TestReloadNotePrintsOncePerDistinctError(t *testing.T) {
	t.Parallel()

	s := &shell{}
	err := errors.New("/tmp/index.html: parse: unexpected EOF")

	line, show := s.reloadNote(err)
	if !show || line != "hot reload: /tmp/index.html: parse: unexpected EOF" {
		t.Fatalf("line = %q, show = %v", line, show)
	}

	if _, show := s.reloadNote(err); show {
		t.Fatal("the same error printed twice")
	}

	s.lastNote = ""

	if _, show := s.reloadNote(err); !show {
		t.Fatal("the error did not print after a success")
	}
}

func TestHotReloadErrorKeepsDrawing(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.err = errors.New("/tmp/index.html: stat: no such file")

	s := newShell(context.Background(), app)
	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 1 || s.display == nil {
		t.Fatalf("polls = %d, display = %v", app.polls, s.display)
	}

	s.lastPoll = time.Now().Add(-time.Second)

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if app.polls != 2 {
		t.Fatalf("polls = %d, want 2", app.polls)
	}

	if app.redraws != 1 {
		t.Fatal("an error reloaded the page")
	}
}

func TestReloadClampsScroll(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1
	app.boxes = []layout.Box{{Y: 500, H: 100, W: 10}}

	s := newShell(context.Background(), app)
	s.scrollY = 550

	if err := s.Update(); err != nil {
		t.Fatal(err)
	}

	if s.scrollY != 0 {
		t.Fatalf("scrollY = %d, want 0 after the shorter page", s.scrollY)
	}
}

func TestResizeRedrawsOnTheFirstChangedFrame(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 2560, maxH: 2560,
	}
	game := NewGame(context.Background(), app)
	shell, ok := game.(*shell)
	if !ok {
		t.Fatalf("game = %T", game)
	}

	shell.Layout(600, 720)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}

	if app.width != 600 || app.height != 720 {
		t.Fatalf("size = %d x %d", app.width, app.height)
	}

	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws after a settled frame = %d, want 1", app.redraws)
	}
}

func TestResizeClampsBeforeRedraw(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 800, maxH: 900,
	}
	shell := newShell(context.Background(), app)

	shell.Layout(50, 50)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.width != 320 || app.height != 400 {
		t.Fatalf("clamped size = %d x %d", app.width, app.height)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}
}

func TestResizeDefersCommitDuringTitlebarDrag(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 2560, maxH: 2560,
	}
	shell := newShell(context.Background(), app)

	shell.windowDrag.active = true
	shell.Layout(600, 720)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 0 {
		t.Fatalf("redraws during drag = %d, want 0", app.redraws)
	}

	if app.width != 480 || app.height != 640 {
		t.Fatalf("size during drag = %d x %d, want unchanged", app.width, app.height)
	}

	shell.windowDrag.active = false
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws after release = %d, want 1", app.redraws)
	}

	if app.width != 600 || app.height != 720 {
		t.Fatalf("size after release = %d x %d, want 600 x 720", app.width, app.height)
	}
}

// replayScreen hands out one display list per generation.
type replayScreen struct {
	fakeScreen
}

func (r *replayScreen) Display() *layout.Display {
	return &layout.Display{Width: r.width, Height: r.height}
}

func TestMotionKeepsTheDrawnArtifact(t *testing.T) {
	t.Parallel()

	app := &replayScreen{fakeScreen: fakeScreen{width: 480, height: 640, minW: 1, minH: 1}}
	shell := newShell(context.Background(), app)

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	drawn := shell.display
	shell.lastRelayout = time.Now()

	shell.Layout(600, 700)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if shell.display != drawn || shell.fallback {
		t.Fatal("a throttled motion frame rebuilt the drawn artifact")
	}
}

func TestResizeCommitsOncePerSettledSize(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 480, height: 640,
		minW: 1, minH: 1,
		maxW: 2560, maxH: 2560,
	}
	shell := newShell(context.Background(), app)

	// Thirty window events between two updates settle on the last size.
	for i := 0; i < 30; i++ {
		shell.Layout(500+i, 700+i)
	}

	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}

	if app.width != 529 || app.height != 729 {
		t.Fatalf("size = %d x %d, want 529 x 729", app.width, app.height)
	}

	// Motion inside the throttle keeps the previous frame on screen.
	shell.lastRelayout = time.Now()

	for i := 0; i < 30; i++ {
		shell.Layout(600+i, 800+i)
	}

	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 1 {
		t.Fatalf("throttled redraws = %d, want 1", app.redraws)
	}

	// The next update has no event, so the size has settled.
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.redraws != 2 {
		t.Fatalf("settled redraws = %d, want 2", app.redraws)
	}

	if app.width != 629 || app.height != 829 {
		t.Fatalf("settled size = %d x %d, want 629 x 829", app.width, app.height)
	}
}

// growScreen keeps the previous display until Redraw builds the new one, so
// a test can catch a cursor mapped through the stale artifact.
type growScreen struct {
	*fakeScreen
	display *layout.Display
	hoverX  float64
}

func (g *growScreen) Display() *layout.Display { return g.display }

func (g *growScreen) Hover(_ context.Context, x, y float64) error {
	g.hoverX = x

	return nil
}

func (g *growScreen) Redraw(context.Context) error {
	g.redraws++
	g.display = &layout.Display{Width: g.width, Height: g.height}

	return nil
}

func TestResizeGrowMapsThroughTheCommittedFrame(t *testing.T) {
	t.Parallel()

	app := &growScreen{
		fakeScreen: &fakeScreen{width: 300, height: 500, minW: 1, minH: 1},
		display:    &layout.Display{Width: 300, Height: 500},
	}
	s := newShell(context.Background(), app)
	s.cursorX, s.cursorY = 550, 40

	s.Layout(600, 500)
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if app.hoverX != 550 {
		t.Fatalf("hover x = %v, want 550 through the committed frame", app.hoverX)
	}
}

// moveScreen keeps one control against the right edge, so a window resize
// moves it out from under or back under a fixed cursor.
type moveScreen struct {
	fakeScreen
	hover   string
	pressed string
}

func (m *moveScreen) Redraw(context.Context) error {
	m.redraws++
	m.boxes = []layout.Box{{ID: "close", X: float64(m.width - 80), Y: 20, W: 60, H: 40}}

	return nil
}

func (m *moveScreen) Hover(_ context.Context, x, y float64) error {
	return m.set(&m.hover, m.hit(x, y))
}

func (m *moveScreen) Press(_ context.Context, x, y float64) error {
	return m.set(&m.pressed, m.hit(x, y))
}

func (m *moveScreen) set(id *string, next string) error {
	if *id == next {
		return nil
	}

	*id = next
	m.redraws++

	return nil
}

func (m *moveScreen) hit(x, y float64) string {
	for _, b := range m.boxes {
		if x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
			return b.ID
		}
	}

	return ""
}

func TestResizeReResolvesHoverAndPress(t *testing.T) {
	t.Parallel()

	app := &moveScreen{fakeScreen: fakeScreen{
		width: 600, height: 500,
		minW: 1, minH: 1,
		maxW: 2560, maxH: 2560,
	}}
	app.boxes = []layout.Box{{ID: "close", X: 520, Y: 20, W: 60, H: 40}}
	app.hover = "close"
	app.pressed = "close"

	shell := newShell(context.Background(), app)
	shell.cursorX, shell.cursorY = 550, 40
	shell.mouseDown = true

	shell.Layout(300, 500)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.hover != "" || app.pressed != "" {
		t.Fatalf("hover = %q, pressed = %q after the control moved away", app.hover, app.pressed)
	}

	// Widen back: the control returns under the cursor.
	shell.Layout(600, 500)
	if err := shell.resize(); err != nil { // moving, inside the throttle
		t.Fatal(err)
	}

	if err := shell.resize(); err != nil { // settled
		t.Fatal(err)
	}

	if app.hover != "close" || app.pressed != "close" {
		t.Fatalf("hover = %q, pressed = %q after the control returned", app.hover, app.pressed)
	}
}

// freeScreen clamps to the minimum only, the way page.Page does now, and
// reports the max as a window bound.
type freeScreen struct {
	fakeScreen
}

func (f *freeScreen) Clamp(width, height int) (int, int) {
	if width < f.minW {
		width = f.minW
	}

	if height < f.minH {
		height = f.minH
	}

	return width, height
}

func (f *freeScreen) SetSize(width, height int) {
	f.width, f.height = f.Clamp(width, height)
}

func (f *freeScreen) MaxSize() (int, int) {
	return f.maxW, f.maxH
}

func TestResizeLaysOutPastTheMax(t *testing.T) {
	t.Parallel()

	app := &freeScreen{fakeScreen: fakeScreen{
		width: 480, height: 640,
		minW: 320, minH: 400,
		maxW: 800, maxH: 900,
	}}
	shell := newShell(context.Background(), app)

	shell.Layout(2000, 1500)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if app.width != 2000 || app.height != 1500 {
		t.Fatalf("size = %d x %d, want 2000 x 1500", app.width, app.height)
	}

	if shell.stretched() {
		t.Fatal("stretched with the frame at the window size")
	}
}

func TestWindowBoundsReadsMaxSize(t *testing.T) {
	t.Parallel()

	bounded := &freeScreen{fakeScreen: fakeScreen{maxW: 800, maxH: 900}}
	if maxW, maxH := windowBounds(bounded); maxW != 800 || maxH != 900 {
		t.Fatalf("bounds = %d x %d, want 800 x 900", maxW, maxH)
	}

	plain := &fakeScreen{}
	if maxW, maxH := windowBounds(plain); maxW != -1 || maxH != -1 {
		t.Fatalf("plain bounds = %d x %d, want -1 x -1", maxW, maxH)
	}
}

// scrollScreen is tall above 500 px wide, so a shrink makes the content
// shorter than the current offset.
type scrollScreen struct {
	fakeScreen
}

func (s *scrollScreen) Redraw(context.Context) error {
	s.redraws++
	s.boxes = nil

	if s.width > 500 {
		s.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: float64(s.width), H: 1200}}
	}

	return nil
}

func (s *scrollScreen) Display() *layout.Display {
	return &layout.Display{Width: s.width, Height: s.height}
}

func TestScrollClampsAfterAShrink(t *testing.T) {
	t.Parallel()

	app := &scrollScreen{fakeScreen: fakeScreen{width: 600, height: 300, minW: 1, minH: 1}}
	app.boxes = []layout.Box{{ID: "tall", X: 0, Y: 0, W: 600, H: 1200}}

	shell := newShell(context.Background(), app)

	if _, contentH := shell.contentSize(); contentH != 1200 {
		t.Fatalf("content height = %d, want 1200", contentH)
	}

	shell.scrollY = 900 // the bottom of a 600 x 300 window

	shell.Layout(400, 300)
	if err := shell.resize(); err != nil {
		t.Fatal(err)
	}

	if err := shell.syncImage(); err != nil {
		t.Fatal(err)
	}

	if shell.scrollY != 0 {
		t.Fatalf("scrollY = %d after the shrink, want 0", shell.scrollY)
	}
}

func TestNativeLandscapeStillScrollsVertically(t *testing.T) {
	ctx := context.Background()
	p, err := page.New(page.Config{
		HTML:  `<body style="margin:0"><div style="height:1400px;background:red"></div></body>`,
		Width: 420, Height: 934, MinWidth: 320, MinHeight: 480,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(ctx, p); err != nil {
		t.Fatal(err)
	}
	s := newShell(ctx, p)
	s.Layout(885, 420)
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}
	if s.stretched() {
		t.Fatal("native landscape viewport is stretched instead of scrollable")
	}
	w, h := s.frameSize()
	if err := s.touchMove(touchUpdate{dy: -120}, nil, w, h); err != nil {
		t.Fatal(err)
	}
	if s.scrollY != 120 || s.scrollX != 0 {
		t.Fatalf("vertical swipe became offset %d,%d", s.scrollX, s.scrollY)
	}
}
