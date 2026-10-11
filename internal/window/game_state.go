package window

import (
	"context"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

type shell struct {
	transparent  bool
	interactive  func(int, int) bool
	draggable    func(int, int) bool
	perf         bool
	perfHooks    perfHooks
	windowDrag   windowDrag
	passthrough  bool
	app          host.Screen
	ctx          context.Context
	img          *ebiten.Image
	fitBuf       *ebiten.Image
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
	// redrawX and redrawY are the scroll offsets the current display was
	// built with; viewport-pinned layers draw against them.
	redrawX  int
	redrawY  int
	dragAxis int
	dragGrab float64
	hold     longPressWatch

	gesture          pointerDrag
	mouseDown        bool
	touchCanceled    bool
	fingers          touchGesture
	pageZoom         float64
	tabEaten         bool
	f11Eaten         bool
	clicks           clickWatch
	dragActive       bool
	dragX, dragY     float64
	menu             menuState
	cursor           ebiten.CursorShapeType
	setCursor        func(ebiten.CursorShapeType)
	readFullscreen   func() bool
	applyFullscreen  func(bool)
	moving           bool
	lastRelayout     time.Time
	cursorX, cursorY int

	replayBuf *ebiten.Image

	lastPoll time.Time
	lastNote string
	viewport viewportState
	partial  partialState
	dev      devState
	ime      imeState
	commits  uint64
	skipped  uint64
}
