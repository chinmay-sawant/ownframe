package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

type shell struct {
	renderer    replay.Renderer
	transparent bool
	interactive func(int, int) bool
	draggable   func(int, int) bool
	perf        bool
	perfHooks   perfHooks
	windowDrag  windowDrag
	passthrough bool
	app         host.Screen
	// ctx is run-scoped, set once in NewGame and never per request.
	// Ebiten Update/Draw take no ctx, so storage is forced.
	ctx context.Context
	// cancel releases ctx; set once with ctx, never replaced.
	cancel       context.CancelFunc
	img          *ebiten.Image
	scaleBuf     *ebiten.Image
	display      *layout.Display
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
	commits    uint64
	skipped    uint64
	// placementID remembers the window rect across launches; empty disables.
	placementID        string
	lastPlacement      windowPlacement
	lastPlacementValid bool
	lastPlacementSave  time.Time
	// dragTrace logs one drag line per applied move. Dev-only: set from
	// OWNFRAME_DRAG_DEBUG with the ownframe_debug build tag.
	dragTrace bool
}
