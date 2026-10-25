package window

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// devState is the window's side of the devtools overlay: the mode flags, the
// boxes under the cursor and pinned, the panel rects, and the frame numbers.
// The page owns the on/off flag, so Config.DevTools and SetDevTools work.
type devState struct {
	mobile         bool
	mobileExpanded bool
	on             bool
	available      bool
	picker         bool
	ops            bool
	tab            devTab
	dockW          float64
	scroll         float64
	collapsed      map[string]bool
	rows           []devRow
	opsCache       devOpsCache
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
