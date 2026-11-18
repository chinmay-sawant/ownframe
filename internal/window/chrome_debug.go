package window

import (
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/debugtag"
	"github.com/hajimehoshi/ebiten/v2"
)

const windowTraceInterval = 100 * time.Millisecond

type windowTraceState struct {
	latestEbiten      workarea
	latestNative      workarea
	monitor           string
	maximized         bool
	known             bool
	nativeKnown       bool
	nativeOK          bool
	dirty             bool
	environmentLogged bool
	lastSample        time.Time
	lastLog           time.Time
}

func windowTraceEnabled() bool {
	return debugtag.Enabled && os.Getenv("OWNFRAME_DRAG_DEBUG") != ""
}

func windowTracef(format string, args ...any) {
	if !windowTraceEnabled() {
		return
	}
	fmt.Fprintf(os.Stderr, "[window] "+format+"\n", args...)
}

func manualMaxTrace() string {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	return fmt.Sprintf("active=%t known=%t learning=%t restoring=%t monitor=%q saved=(%d,%d %dx%d) workarea=(%d,%d %dx%d)",
		manualMax.active, manualMax.known, manualMax.learning, manualMax.restoring,
		manualMax.learnedMon, manualMax.sx, manualMax.sy, manualMax.sw, manualMax.sh,
		manualMax.wx, manualMax.wy, manualMax.ww, manualMax.wh)
}

func traceMaximizeState(event string) {
	if !windowTraceEnabled() {
		return
	}

	x, y := windowPosition()
	w, h := ebiten.WindowSize()
	windowTracef("maximize %s decorated=%t manualWanted=%t WSLg=%t wmMaximized=%t monitor=%q rect=(%d,%d %dx%d) manual={%s}",
		event, ebiten.IsWindowDecorated(), manualMaximizeWanted(), isWSLg(),
		ebiten.IsWindowMaximized(), currentMonitorName(), x, y, w, h, manualMaxTrace())
}

func (s *shell) traceWindowGeometry() {
	if !s.dragTrace {
		return
	}

	x, y := ebiten.WindowPosition()
	w, h := ebiten.WindowSize()
	maximized := ebiten.IsWindowMaximized()
	ebitenRect := workarea{x: x, y: y, w: w, h: h}
	state := &s.windowTrace
	now := time.Now()
	urgent := !state.known || ebitenRect.w != state.latestEbiten.w || ebitenRect.h != state.latestEbiten.h || maximized != state.maximized
	if !state.known || ebitenRect != state.latestEbiten || maximized != state.maximized {
		state.latestEbiten = ebitenRect
		state.maximized = maximized
		state.known = true
		state.dirty = true
	}

	var nativeRect workarea
	nativeOK := false
	layout := []monitorRect(nil)
	if state.lastSample.IsZero() || now.Sub(state.lastSample) >= windowTraceInterval {
		nativeRect, nativeOK = nativeWindowGeometry()
		layout = monitorLayout()
		monitor := ""
		if nativeOK {
			if mon, ok := monitorForWindow(nativeRect.x, nativeRect.y, nativeRect.w, nativeRect.h, layout); ok {
				monitor = mon.name
			}
		} else if mon, ok := monitorForWindow(x, y, w, h, layout); ok {
			monitor = mon.name
		} else if ebitenMonitor := ebiten.Monitor(); ebitenMonitor != nil {
			monitor = ebitenMonitor.Name()
		}

		if !state.nativeKnown || nativeOK != state.nativeOK || nativeOK && nativeRect != state.latestNative {
			state.latestNative = nativeRect
			state.nativeOK = nativeOK
			state.nativeKnown = true
			state.dirty = true
		}
		state.monitor = monitor
		state.lastSample = now
	}
	if !state.dirty || (!urgent && !state.lastLog.IsZero() && now.Sub(state.lastLog) < windowTraceInterval) {
		return
	}

	if !state.environmentLogged {
		windowTracef("environment WSL=%t WSLg=%t decorated=%t customDrag=%t monitors=%+v ebitenMonitors=%+v",
			isWSL(), isWSLg(), ebiten.IsWindowDecorated(), s.draggable != nil, layout, liveMonitorGeometries())
		if target, ok := nativeWindowTarget(); ok {
			windowTracef("native frame client=%+v frame=%+v", target.Client, target.Frame)
		}
		state.environmentLogged = true
	}

	windowTracef("geometry ebiten=(%d,%d %dx%d) native=(%d,%d %dx%d ok=%t) monitor=%q wmMaximized=%t decorated=%t customDrag=%t manual={%s}",
		x, y, w, h, state.latestNative.x, state.latestNative.y, state.latestNative.w, state.latestNative.h,
		state.nativeOK, state.monitor, maximized, ebiten.IsWindowDecorated(), s.draggable != nil, manualMaxTrace())
	if m := ebiten.Monitor(); m != nil {
		windowTracef("geometry selectedEbitenMonitor=%q", m.Name())
	}
	state.lastLog = now
	state.dirty = false
}
