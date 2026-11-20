package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/debugtag"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// NewGame returns the screen loop Run and BindMobile use.
// The ctx is run-scoped: stored once with its cancel, never per request.
func NewGame(ctx context.Context, app host.Screen) ebiten.Game {
	return newShell(ctx, app)
}

// newShell builds the shell NewGame and NewGameWithOptions share, so
// neither needs a type assertion on the game value.
func newShell(ctx context.Context, app host.Screen) *shell {
	ebiten.SetVsyncEnabled(false)
	if ctx == nil {
		ctx = context.Background()
	}
	ctxDone := ctx.Done()
	runCtx, cancel := context.WithCancel(ctx)

	width, height := 0, 0
	if app != nil {
		width, height = app.Size()
	}

	game := &shell{
		app:        app,
		ctx:        runCtx,
		ctxDone:    ctxDone,
		cancel:     cancel,
		chords:     newChordWatch(),
		watched:    newKeyWatch(),
		pageZoom:   1,
		dev:        devState{watch: newKeyWatch(), mobile: mobileDevTools},
		pendingW:   width,
		pendingH:   height,
		screenW:    width,
		screenH:    height,
		perf:       perfEnabled(app),
		a11yOrigin: ebiten.WindowPosition,
		// Drag tracing stays out of shipped builds: the build tag gates
		// it like the pacing debug output, and the variable gates a run.
		dragTrace: debugtag.Enabled && os.Getenv("OWNFRAME_DRAG_DEBUG") != "",
	}
	game.wirePerf()
	game.imeInit()
	game.initA11y()
	game.configureTouchScroll()

	return game
}

type shell struct {
	renderer    replay.Renderer
	transparent bool
	interactive func(int, int) bool
	draggable   func(int, int) bool
	perf        bool
	perfOption  bool
	perfHooks   perfHooks
	// debugEnabled serves the profiling snapshot on DebugAddr. Set once
	// before the loop when a debug build starts its listener; the game
	// thread reads it, the HTTP thread only loads debugSnapshot.
	debugEnabled bool
	// debugSnapshot is the last published profiling copy. The game thread
	// stores fresh immutable copies, the HTTP thread loads them, so no
	// mutex guards the frame path.
	debugSnapshot atomic.Pointer[windowDebugState]
	debugDisplay  *layout.Display
	// debugLast throttles snapshot publishes. Game thread only.
	debugLast    time.Time
	windowDrag   windowDrag
	nativeCursor nativeWindowDrag
	passthrough  bool
	app          host.Screen
	// ctx is run-scoped, set once in NewGame and never per request.
	// Ebiten Update/Draw take no ctx, so storage is forced.
	ctx context.Context
	// ctxDone is the caller's done channel. A background context has none,
	// so no watcher is started for hosts that cannot observe cancellation.
	ctxDone <-chan struct{}
	// cancel releases ctx; set once with ctx, never replaced.
	cancel         context.CancelFunc
	img            *ebiten.Image
	scaleBuf       *ebiten.Image
	display        *layout.Display
	viewportSticky bool
	// drawSyncErr is a sync failure from Draw. Update reports it, because
	// Draw cannot return an error.
	drawSyncErr  error
	fallback     bool
	bitmapView   bitmapView
	orderCache   orderCache
	seq          uint64
	contentW     int
	contentH     int
	contentGen   uint64
	chars        []rune
	chords       chordWatch
	watched      keyWatch
	pendingW     int
	pendingH     int
	screenW      int
	screenH      int
	scrollX      int
	scrollY      int
	scrollMotion touchScrollMotion
	// redrawX/Y are offsets the display was built with.
	redrawX  int
	redrawY  int
	dragAxis int
	dragGrab float64
	hold     longPressWatch

	mouseDown  bool
	pageZoom   float64
	dragActive bool
	gestureState
	pace             paceState
	menu             menuState
	cursor           ebiten.CursorShapeType
	setCursor        func(ebiten.CursorShapeType)
	readFullscreen   func() bool
	applyFullscreen  func(bool)
	moving           bool
	lastRelayout     time.Time
	cursorX, cursorY int
	hover            hoverState

	lastPoll time.Time
	lastNote string
	viewport viewportState
	partial  partialState
	frame    frameState
	// frameOptIn is set in Update; Draw may run without Update.
	frameOptIn bool
	dev        devState
	ime        imeState
	a11y       a11yHost
	a11yOrigin func() (int, int)
	commits    uint64
	skipped    uint64
	// placementID remembers the window rect across launches; empty disables.
	placementID        string
	lastPlacement      windowPlacement
	lastPlacementValid bool
	lastPlacementSave  time.Time
	// dragTrace logs one drag line per applied move. Dev-only: set from
	// OWNFRAME_DRAG_DEBUG with the ownframe_debug build tag.
	dragTrace   bool
	windowTrace windowTraceState
	titlebar    titlebarState
}

// gestureState holds the shell's drag and key-chord tracking. The shell embeds
// this state; the game loop owns its fields.
type gestureState struct {
	gesture       pointerDrag
	touchCanceled bool
	fingers       touchGesture
	tabEaten      bool
	f11Eaten      bool
	clicks        clickWatch
	dragX, dragY  float64
}

// roles narrows shell.app to what each call site needs. Entry points
// keep host.Screen for compat; pointer, keyboard, and clipboard code
// routes through one role so fakes implement roles, not the surface.
func (s *shell) ptr() host.Pointer    { return s.app }
func (s *shell) kbd() host.Keyboard   { return s.app }
func (s *shell) clip() host.Clipboard { return s.app }

// cleanLog strips CR/LF so reload text cannot forge stderr lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}

func (s *shell) updatePassthrough() {
	if s.interactive == nil {
		return
	}
	x, y := s.cursorPosition()
	pass := !s.windowDrag.active && !s.interactive(x, y)
	if pass != s.passthrough {
		ebiten.SetWindowMousePassthrough(pass)
		s.passthrough = pass
	}
}

// syncImage follows the page after each Redraw. A display-list page keeps
// its operations and drops any bitmap; any other page keeps one Ebiten image
// built from its picture. The old image is disposed when it is replaced.
func (s *shell) syncImage() error {
	s.syncA11y()
	if s.built() {
		return nil
	}

	front := host.SelectFront(s.app)
	if front.Display != nil {
		s.setDisplay(front.Display)
		s.pullScroll()

		return nil
	}

	if front.Image == nil {
		return errNoImage
	}

	next := ebiten.NewImageFromImage(front.Image)
	s.disposeImage()

	if s.display != nil {
		s.renderer.EvictDisplay(s.display)
	}
	s.img = next
	s.display = nil
	s.viewportSticky = false
	s.fallback = true
	s.seq = front.Generation
	s.pullScroll()

	return nil
}

// built reports that the cached artifact came from the page's current
// generation at the size the page committed, so no rebuild is needed.
func (s *shell) built() bool {
	if s.app.Generation() != s.seq || (s.img == nil && s.display == nil) {
		return false
	}

	w, h := s.app.Size()
	cw, ch := s.canvasSize()

	return cw == w && ch == h
}

func (s *shell) setDisplay(display *layout.Display) {
	s.viewportSticky = layout.HasViewportSticky(display)
	if s.display != nil && s.display != display {
		s.renderer.EvictDisplay(s.display)
	}
	s.disposeImage()
	s.display = display
	s.img = nil
	s.fallback = false
	s.seq = s.app.Generation()
	s.markRedrawOffset()
}

// markRedrawOffset records the scroll offset the current display was built
// with, so viewport-pinned layers can be drawn against it.
func (s *shell) markRedrawOffset() {
	if observer, ok := s.app.(host.ScrollObserver); ok {
		s.redrawX, s.redrawY = observer.ScrollOffset()
	}
}

func (s *shell) disposeImage() {
	if s.img != nil {
		s.img.Deallocate()
		s.img = nil
	}
}
