package window

import (
	"bytes"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/debugtag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomedium"
	"image/color"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// closeRequested ends the desktop run loop with ebiten.Termination on the
// next frame. It backs Window.Close for borderless windows that draw their
// own close control and have no OS close button. RunGame maps Termination
// back to a nil return, so closing this way is a clean exit with no crash
// report.
var closeRequested atomic.Bool

// RequestClose asks the run loop to exit on its next frame. On WSLg it
// waits for the left mouse button to be released before destroying the window.
// It does nothing when no desktop loop is running, so Serve, web, and mobile
// callers can share the same button wiring. It also wakes a parked idle loop, so a close
// pressed while the window is clean exits on the next frame instead of
// waiting for the keep-alive tick or a page deadline.
func RequestClose() {
	closeRequested.Store(true)
	//nolint:staticcheck // reason: concurrent-safe wake for the event-wait idle park; no-op on the fast path.
	ebiten.ScheduleFrame()
}

// takeCloseRequest retains the request while WSLg still owes the window a
// mouse release. Destroying it on press leaves X11 button state stuck down.
func takeCloseRequest(waitForRelease bool) bool {
	if waitForRelease {
		return false
	}
	return closeRequested.Swap(false)
}

// Minimize minimizes the desktop window. It does nothing when no desktop
// run loop is active or the platform is not a desktop.
func Minimize() {
	if isWSLg() {
		if !minimizeWSLg() {
			log.Print("ownframe: Windows host did not accept the WSLg minimize request")
		}
		return
	}
	ebiten.MinimizeWindow()
}

// ToggleMaximize maximizes the desktop window, or restores it when it is
// already maximized. Under WSL it uses the manual workarea maximize below;
// everywhere else it uses the window manager. It does nothing when no
// desktop run loop is active or the platform is not a desktop.
func ToggleMaximize() {
	traceMaximizeState("toggle request")
	if manualMaximizeWanted() {
		toggleManualMaximize()
		return
	}

	if ebiten.IsWindowMaximized() {
		ebiten.RestoreWindow()
	} else {
		ebiten.MaximizeWindow()
	}
}

// Maximized reports whether the desktop window is maximized, including a
// manual workarea maximize on WSL. It is false when no desktop run loop is
// active or the platform is not a desktop.
func Maximized() bool {
	chromeMu.Lock()
	active := manualMax.active
	chromeMu.Unlock()

	return active || ebiten.IsWindowMaximized()
}

// isWSLVersion reports whether a /proc/version string comes from Windows
// Subsystem for Linux. WSL2 kernels carry "microsoft-standard-WSL2" and WSL1
// carries "Microsoft".
func isWSLVersion(version string) bool {
	return strings.Contains(strings.ToLower(version), "microsoft")
}

// isWSL reports whether the process runs inside Windows Subsystem for
// Linux. Only Linux can be WSL; anything else answers false without
// touching the filesystem.
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if _, ok := os.LookupEnv("WSL_DISTRO_NAME"); ok {
		return true
	}

	if _, ok := os.LookupEnv("WSL_INTEROP"); ok {
		return true
	}

	version, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return isWSLVersion(string(version))
}

// manualMaximizeWanted keeps custom chrome on WSL's monitor workarea path.
func manualMaximizeWanted() bool {
	return isWSL() && !ebiten.IsWindowDecorated()
}

// manualMaxLearnTimeout bounds a workarea probe. If the maximized configure
// has not arrived by then, the window stays WM-maximized rather than strand
// the dance halfway. It also bounds the restore leg that follows learning:
// SetWindowSize on a still-maximized window is ignored on WSLg, so the
// learned rect is applied only after the restore configure lands.
const manualMaxLearnTimeout = 2 * time.Second

// workarea is one monitor's usable rect in device-independent pixels.
type workarea struct {
	x, y int
	w, h int
}

// chromeMu guards the process-wide window-manager state below: manualMax,
// wasExternalMaximized, pendingStartMaximize, pendingManualRestore, and the
// pending windowed restore. The desktop run loop is single-goroutine, but
// unit tests drive separate shells in parallel and every resize funnels
// through these globals, so every accessor holds the mutex. A helper that
// runs under an already-held chromeMu carries a Locked suffix and never
// locks on its own.
var chromeMu sync.Mutex

// pendingStartMaximize defers a remembered WM maximize past window creation:
// MaximizeWindow before the run loop never takes effect, so the resize path
// applies it on the first update once the window exists. WSL never uses it;
// the manual path installs client geometry up front instead.
var pendingStartMaximize bool

// pendingManualRestore repeats a remembered WSL maximize after the window
// exists. The monitor list is empty before RunGame, and a size set that
// early is ignored, so a saved 1920 workarea was dropped and the next
// launch opened at the windowed size.
var (
	pendingManualRestore bool
	pendingManualUntil   time.Time
)

// pendingWindowedRestore repeats a pre-loop windowed position after the
// window maps: WSLg ignores sizes and positions set before RunGame, so the
// compositor cascades the frame below the terminal instead of the requested
// panel top-left. The retry re-asserts until landed, the deadline passes,
// or a drag starts.
var (
	pendingWindowedRestore             bool
	pendingWindowedX, pendingWindowedY int
	pendingWindowedW, pendingWindowedH int
	pendingWindowedUntil               time.Time
)

func rememberWindowedRestoreLocked(x, y, w, h int) {
	pendingWindowedRestore = true
	pendingWindowedX, pendingWindowedY = x, y
	pendingWindowedW, pendingWindowedH = w, h
	pendingWindowedUntil = time.Time{}
}

// retryWindowedRestore installs the requested windowed rect once the window
// exists. It yields to the manual workarea restore and to an active drag.
func retryWindowedRestore(dragging bool) {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	if !pendingWindowedRestore || pendingManualRestore {
		return
	}
	if pendingWindowedUntil.IsZero() {
		pendingWindowedUntil = time.Now().Add(manualMaxLearnTimeout)
	}
	if dragging {
		pendingWindowedRestore = false
		return
	}
	if time.Now().After(pendingWindowedUntil) {
		pendingWindowedRestore = false
		// WSLg can reject a cross-monitor programmatic move. Keep the
		// window grabbable on the monitor the compositor actually selected.
		if isWSLg() {
			if rect, ok := nativeWindowGeometry(); ok {
				if mon, ok := monitorForWindow(rect.x, rect.y, rect.w, rect.h, monitorLayout()); ok {
					if area, ok := wslgHostWorkareaLocked(mon); ok {
						x := max(area.x, min(rect.x, area.x+area.w-rect.w))
						y := max(area.y, min(rect.y, area.y+area.h-rect.h))
						setWindowRect(x, y, rect.w, rect.h)
						windowTracef("restore fallback monitor=%q visibleTarget=(%d,%d %dx%d)", mon.name, x, y, rect.w, rect.h)
					}
				}
			}
		}
		return
	}
	x, y := windowPosition()
	w, h := ebiten.WindowSize()
	if restoreRectMatches(workarea{x: x, y: y, w: w, h: h}, workarea{x: pendingWindowedX, y: pendingWindowedY, w: pendingWindowedW, h: pendingWindowedH}) {
		pendingWindowedRestore = false
		return
	}
	setWindowRect(pendingWindowedX, pendingWindowedY, pendingWindowedW, pendingWindowedH)
	windowTracef("restore retry requested=(%d,%d %dx%d) actual=(%d,%d %dx%d)", pendingWindowedX, pendingWindowedY, pendingWindowedW, pendingWindowedH, x, y, w, h)
}

// manualMax carries a hand-rolled maximize for WSL. WSLg can present a
// WM-maximized borderless window with an offset the client cannot see,
// which misaligns pointer input while the layout itself is correct. Manual
// maximize never enters the maximized window state: it saves the windowed
// rect and applies the workarea as plain client geometry, which WSLg
// presents correctly, then restores the saved rect on toggle.
//
// The workarea is cached per monitor name: a 1920-wide panel and a 2560-wide
// panel report different maximized rects, so a cache hit must come from the
// monitor the window is on now, not from a previous probe elsewhere. Names
// are stable across calls; Ebiten monitor pointers are not.
var manualMax = struct {
	active           bool // manually maximized: geometry is client-set workarea
	known            bool // workarea learned for learnedMon
	learning         bool // a WM maximize is in flight to reveal the workarea
	restoring        bool // learned; waiting for the restore configure before applying
	until            time.Time
	sx, sy           int // saved windowed position
	sw, sh           int // saved windowed size
	wx, wy           int // learned workarea position
	ww, wh           int // learned workarea size
	learnedMon       string
	areas            map[string]workarea
	observedWindowed workarea
	observedKnown    bool
	probe            workareaProbe
	hostAreas        []hostWorkarea
	hostAreasAt      time.Time
}{}

// workareaGapMax bounds how far a learned workarea may fall short of its
// monitor's fullscreen size. Taskbars and docks cost tens of pixels; a
// rect hundreds short belongs to another monitor.
const workareaGapMax = 256

// workareaMatchesMonitor reports whether area plausibly is name's workarea:
// it fits and it nearly fills the monitor. A 1920-wide rect fits a
// 2560-wide monitor but leaves 640 pixels unexplained, so it must not
// apply there. Unknown monitors keep the old behavior and report true so
// a missing probe never discards a good rect.
func workareaMatchesMonitor(name string, area workarea) bool {
	if !placementFits(name, area.w, area.h) {
		return false
	}

	monW, monH, ok := placementMonitorSize(name)
	if !ok {
		return true
	}

	return areaMatchesMonitorSize(monW, monH, area)
}

// areaMatchesMonitorSize is the pure gap rule behind workareaMatchesMonitor:
// the rect must fit and nearly fill the given monitor size. A rect hundreds
// of pixels short in either axis belongs to another monitor.
func areaMatchesMonitorSize(monW, monH int, area workarea) bool {
	if area.w > monW || area.h > monH {
		return false
	}

	return monW-area.w <= workareaGapMax && monH-area.h <= workareaGapMax
}

// monitorGeometry is one display's name and size in device pixels.
// Ebiten reports sizes but not origins, so attribution matches on size.
type monitorGeometry struct {
	name string
	w, h int
}

// liveMonitorGeometries snapshots the attached monitors in list order.
func liveMonitorGeometries() []monitorGeometry {
	var out []monitorGeometry
	for _, m := range ebiten.AppendMonitors(nil) {
		w, h := m.Size()
		out = append(out, monitorGeometry{name: m.Name(), w: w, h: h})
	}

	return out
}

// attributeWorkarea names the monitor a learned workarea plausibly fills.
// The claimant wins when it fits; otherwise the first matching monitor
// wins, so a rect learned while ebiten.Monitor straddled (or misreported)
// still lands on the panel it actually fills. No match keeps the claimant,
// which preserves the old apply-anyway behavior for exotic compositors.
func attributeWorkarea(area workarea, claimant string, mons []monitorGeometry) (string, bool) {
	for _, m := range mons {
		if m.name == claimant && areaMatchesMonitorSize(m.w, m.h, area) {
			return claimant, true
		}
	}
	for _, m := range mons {
		if m.name != claimant && areaMatchesMonitorSize(m.w, m.h, area) {
			return m.name, true
		}
	}

	return claimant, false
}

// currentMonitorName reports the monitor holding the window center, using
// true xrandr origins when available. Ebiten alone reports sizes but not
// origins and misreports while the window straddles monitors, so a probe
// learned on the 2560 panel can claim the 1920 name. The center answers
// which monitor the header is on when the user double-taps it.
func currentMonitorName() string {
	wx, wy := windowPosition()
	ww, wh := ebiten.WindowSize()
	if layout := monitorLayout(); len(layout) != 0 {
		if m, ok := monitorForWindow(wx, wy, ww, wh, layout); ok {
			return m.name
		}
	}
	if monitor := ebiten.Monitor(); monitor != nil {
		return monitor.Name()
	}

	return ""
}

// wasExternalMaximized tracks the raw maximized flag across updates so a
// foreign maximize (snap, shortcut, taskbar) is adopted exactly once.
var wasExternalMaximized bool

// adoptExternalMaximize converts a foreign window-manager maximize into the
// manual dance on WSL, where a raw maximized borderless window mispresents
// with an invisible offset. Drag-to-top snaps, taskbar buttons, and keyboard
// shortcuts all arrive here; only the titlebar button saves its own rect.
// While windowed it keeps the restore target fresh, so the dance after a
// convert restores the rect the snap replaced. Call it every update, next
// to completeManualMaximize.
func adoptExternalMaximize() {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	maxed := ebiten.IsWindowMaximized()
	if maxed != wasExternalMaximized {
		windowTracef("maximize window-manager state changed old=%t new=%t decorated=%t manualWanted=%t", wasExternalMaximized, maxed, ebiten.IsWindowDecorated(), manualMaximizeWanted())
	}
	defer func() { wasExternalMaximized = maxed }()

	if !manualMaximizeWanted() || manualMax.active || manualMax.learning || manualMax.restoring {
		return
	}
	if !maxed {
		if isWSLg() {
			x, y := ebiten.WindowPosition()
			w, h := ebiten.WindowSize()
			observed := workarea{x: x, y: y, w: w, h: h}
			if manualMax.observedKnown && manualMax.observedWindowed == observed {
				return
			}
			manualMax.observedKnown = true
			manualMax.observedWindowed = observed
		}
		// While windowed, keep the restore target fresh. The toggle path
		// overwrites these with the same live values, so both agree.
		manualMax.sx, manualMax.sy = windowPosition()
		manualMax.sw, manualMax.sh = ebiten.WindowSize()

		return
	}
	if wasExternalMaximized {
		return // already adopted (or already raw); never stack dances
	}
	if !validPlacementRect(manualMax.sw, manualMax.sh) {
		return
	}
	// Already maximized: MaximizeWindow would be redundant, so start the
	// learn dance directly and let completeManualMaximize pick up the
	// configure the manager is presenting.
	manualMax.probe = workareaProbe{}
	manualMax.learning = true
	manualMax.until = time.Now().Add(manualMaxLearnTimeout)
}

// toggleManualMaximize flips between the saved windowed rect and the
// learned workarea. The first maximize per monitor learns the workarea with
// one window-manager maximize (see completeManualMaximize) and converts it
// to client geometry once the restore configure lands.
func toggleManualMaximize() {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	if manualMax.learning || manualMax.restoring {
		// A workarea probe is in flight; swallowing the extra click keeps
		// one maximize in the air instead of stacking two.
		return
	}

	if manualMaximizeWanted() && ebiten.IsWindowMaximized() && !manualMax.active {
		// A raw window-manager maximize (keyboard shortcut, taskbar, or a
		// snap the release handler missed) has no saved windowed rect, so
		// a probe would relearn the size the window already has. Restore
		// first, in the style of a native titlebar; the next toggle
		// maximizes from the windowed rect.
		ebiten.RestoreWindow()

		return
	}

	if manualMax.active {
		if w, h := ebiten.WindowSize(); w != manualMax.ww || h != manualMax.wh {
			// The user moved or resized out from under the manual
			// maximize; drop it and start fresh below.
			manualMax.active = false
		} else {
			setWindowRect(manualMax.sx, manualMax.sy, manualMax.sw, manualMax.sh)
			manualMax.active = false
			rememberWindowedRestoreLocked(manualMax.sx, manualMax.sy, manualMax.sw, manualMax.sh)

			return
		}
	}

	manualMax.sx, manualMax.sy = windowPosition()
	manualMax.sw, manualMax.sh = ebiten.WindowSize()
	current := currentMonitorName()
	if isWSLg() {
		for _, m := range monitorLayout() {
			if m.name == current {
				if area, ok := wslgHostWorkareaLocked(m); ok {
					applyWorkareaLocked(current, area)
					return
				}
				break
			}
		}
	}

	// Check the resolution first: the cached workarea must fill the live
	// monitor, not just the named one. ebiten.Monitor is unreliable while
	// the window straddles monitors, so a probe learned on one panel can
	// claim a name from the other. Requiring a match against the live size
	// keeps a stale 1920 entry from teleporting the window across screens
	// when maximizing on the 2560 panel, and vice versa.
	liveW, liveH, liveOK := liveMonitorResolution()
	matchesLive := func(area workarea) bool {
		if !liveOK {
			return true
		}
		return areaMatchesMonitorSize(liveW, liveH, area)
	}

	if area, ok := manualMax.areas[current]; ok {
		// A stale entry from the other monitor teleports the window
		// across screens. Drop the entry and re-probe instead.
		if !workareaMatchesMonitor(current, area) || !matchesLive(area) {
			delete(manualMax.areas, current)
		} else {
			applyWorkareaLocked(current, area)

			return
		}
	}

	if manualMax.known && manualMax.learnedMon == current {
		learned := workarea{x: manualMax.wx, y: manualMax.wy, w: manualMax.ww, h: manualMax.wh}
		// Same stale-entry guard as the cache above: a rect learned
		// under a wrong monitor claim must re-probe, not teleport.
		if workareaMatchesMonitor(current, learned) && matchesLive(learned) {
			applyWorkareaLocked(current, learned)

			return
		}
	}

	manualMax.probe = workareaProbe{}
	manualMax.learning = true
	manualMax.until = time.Now().Add(manualMaxLearnTimeout)
	ebiten.MaximizeWindow()
}

// applyWorkarea installs a learned rect as plain client geometry and marks
// the manual maximize active. Callers must have saved the windowed rect.
func applyWorkarea(mon string, area workarea) {
	chromeMu.Lock()
	defer chromeMu.Unlock()
	applyWorkareaLocked(mon, area)
}

// applyWorkareaLocked is applyWorkarea for callers holding chromeMu.
func applyWorkareaLocked(mon string, area workarea) {
	windowTracef("manual maximize applying monitor=%q target=%+v saved=(%d,%d %dx%d)", mon, area, manualMax.sx, manualMax.sy, manualMax.sw, manualMax.sh)
	manualMax.wx, manualMax.wy = area.x, area.y
	manualMax.ww, manualMax.wh = area.w, area.h
	manualMax.learnedMon = mon
	manualMax.known = true
	if manualMax.areas == nil {
		manualMax.areas = make(map[string]workarea)
	}
	manualMax.areas[mon] = area
	setWindowRect(area.x, area.y, area.w, area.h)
	manualMax.active = true
	pendingManualRestore = true
	pendingManualUntil = time.Time{}
}

// completeManualMaximize watches a workarea probe from the resize path. When
// the maximized configure arrives it records the workarea, leaves the
// maximized window state, and re-applies the learned rect as plain client
// geometry once the restore configure lands. A same-frame Restore followed
// by SetWindowSize loses the resize on WSLg, which is the double-click the
// probe reproduced, so the apply waits for the windowed size to return.
// Call it with the latest Layout size every update.
func completeManualMaximize(pendingW, pendingH int) {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	if manualMax.learning {
		if time.Now().After(manualMax.until) {
			// The configure never arrived; stay WM-maximized rather than
			// strand the window halfway through the dance.
			manualMax.learning = false

			return
		}

		if !ebiten.IsWindowMaximized() {
			return
		}

		if !isWSLg() && pendingW == manualMax.sw && pendingH == manualMax.sh {
			return // still the windowed size; the configure is in flight
		}

		area := workarea{w: pendingW, h: pendingH}
		area.x, area.y = windowPosition()
		if isWSLg() {
			if actual, ok := nativeWindowGeometry(); ok {
				area = actual
				if !manualMax.probe.observe(area, time.Now()) {
					return
				}
			} else {
				return
			}
		}
		// Attribute by position first: the double-tap header is on one
		// monitor, and xrandr gives its true origin. Size-only matching
		// misattributes when Ebiten straddles or misreports, teleporting a
		// 2560 rect learned under a 1920 claimant across screens. The rect
		// must also fill that monitor, or it belongs to the other panel.
		mons := liveMonitorGeometries()
		current := currentMonitorName()
		mon, matched := attributeWorkarea(area, current, mons)
		if layout := monitorLayout(); len(layout) != 0 {
			if posMon, ok := monitorForWindow(area.x, area.y, area.w, area.h, layout); ok {
				if areaMatchesMonitorSize(posMon.w, posMon.h, area) {
					mon, matched = posMon.name, true
				} else if !matched {
					mon, matched = posMon.name, false
				}
			}
			// Snap the learned origin into the attributed panel. WSLg
			// reports a maximized 2560 workarea at virtual (0,0) while the
			// panel lives at (1920,0); applying (0,0) strands the frame
			// across the dead zone instead of the panel top-left.
			for _, m := range layout {
				if m.name == mon && areaMatchesMonitorSize(m.w, m.h, area) {
					area.x = m.x
					if area.y < m.y || area.y > m.y+workareaGapMax {
						area.y = m.y
					}
					break
				}
			}
		}
		manualMax.wx, manualMax.wy = area.x, area.y
		manualMax.ww, manualMax.wh = area.w, area.h
		manualMax.learnedMon = mon
		manualMax.known = true
		if matched {
			if manualMax.areas == nil {
				manualMax.areas = make(map[string]workarea)
			}
			manualMax.areas[mon] = area
		}
		manualMax.learning = false
		manualMax.restoring = true
		manualMax.until = time.Now().Add(manualMaxLearnTimeout)
		// RestoreWindow is guarded by the maximized check above, matching the
		// ToggleMaximize pattern. The resize below waits for the restore
		// configure; applying it in this frame would be ignored while the
		// window still reports maximized.
		ebiten.RestoreWindow()

		return
	}

	if !manualMax.restoring {
		return
	}

	if time.Now().After(manualMax.until) {
		// The restore configure never landed; keep the windowed rect and
		// let the next toggle use the learned workarea directly.
		manualMax.restoring = false

		return
	}

	if ebiten.IsWindowMaximized() {
		return // the restore has not taken effect yet
	}

	if pendingW != manualMax.sw || pendingH != manualMax.sh {
		return // still the maximized size; the restore configure is in flight
	}

	manualMax.restoring = false
	setWindowRect(manualMax.wx, manualMax.wy, manualMax.ww, manualMax.wh)
	manualMax.active = true
	pendingManualRestore = true
	pendingManualUntil = time.Time{}
}

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

// fullscreenSupported reports that the platform owns a window that can go
// fullscreen. wasm and mobile keep their own fullscreen.
func fullscreenSupported() bool {
	switch runtime.GOOS {
	case "js", "android", "ios":
		return false
	}

	return true
}

// fullscreenNow reports the current fullscreen state.
func (s *shell) fullscreenNow() bool {
	if s.readFullscreen != nil {
		return s.readFullscreen()
	}

	if !fullscreenSupported() {
		return false
	}

	return ebiten.IsFullscreen()
}

// setFullscreen changes the fullscreen state.
func (s *shell) setFullscreen(on bool) {
	if s.applyFullscreen != nil {
		s.applyFullscreen(on)

		return
	}

	if fullscreenSupported() {
		ebiten.SetFullscreen(on)
	}
}

// f11 toggles fullscreen and reports whether the window consumed the key.
// The press and its release are consumed together.
func (s *shell) f11(down bool) bool {
	if !down {
		taken := s.f11Eaten
		s.f11Eaten = false

		return taken
	}

	if !fullscreenSupported() {
		return false
	}

	s.setFullscreen(!s.fullscreenNow())
	s.f11Eaten = true

	return true
}

const titlebarHeight = 36
const titlebarButtonWidth = 46

// Caption button numbers, left to right, as the switch statements see them.
// titlebarButtonCount is how many caption buttons the titlebar draws.
const (
	buttonMaximize      = 2
	buttonClose         = 3
	titlebarButtonCount = 3
	// halfDivisor halves a size to find its center.
	halfDivisor = 2
)

// Title layout: font size in points, the smallest header height a caller can
// ask for, the padding around the title, and the room kept free at its end.
// midpointDivisor halves the search range when fitting the title to its width.
const (
	headerFontSize  = 13
	minHeaderHeight = 24
	midpointDivisor = 2
	titleInsetX     = 14
	titleEndReserve = 28
)

var contentTop atomic.Int32

// storeContentTop records the titlebar inset that page content starts below.
// Insets are never negative and are far below MaxInt32, so the clamp only
// makes the int32 conversion provably safe.
func storeContentTop(inset int) {
	contentTop.Store(int32(min(max(inset, 0), math.MaxInt32)))
}

// ContentCursorPosition returns the mouse position in page coordinates.
func ContentCursorPosition() (int, int) {
	x, y := ebiten.CursorPosition()
	return x, y - int(contentTop.Load())
}

type titlebarState struct {
	style       headerStyle
	enabled     bool
	starting    bool
	fixed       bool
	pressed     int
	down        bool
	canvas      *ebiten.Image
	title       string
	fittedTitle string
	titleWidth  int
}

type headerStyle struct {
	title                                     string
	height                                    int
	font                                      text.Face
	background, foreground, hover, closeHover color.Color
}

func newHeaderFont() (text.Face, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(gomedium.TTF))
	if err != nil {
		return nil, err
	}
	return &text.GoTextFace{Source: source, Size: headerFontSize}, nil
}

func fitHeaderTitle(title string, face text.Face, width int) string {
	if width <= 0 {
		return ""
	}
	if w, _ := text.Measure(title, face, 0); w <= float64(width) {
		return title
	}
	const ellipsis = "…"
	if w, _ := text.Measure(ellipsis, face, 0); w > float64(width) {
		return ""
	}
	runes := []rune(title)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / midpointDivisor
		w, _ := text.Measure(string(runes[:mid])+ellipsis, face, 0)
		if w <= float64(width) {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:lo]) + ellipsis
}

func headerHeight(o Options) int {
	if o.HeaderHeight <= 0 {
		return titlebarHeight
	}
	return max(minHeaderHeight, o.HeaderHeight)
}

func (s *shell) placeTitlebarWindow() {
	if !s.titlebar.starting {
		return
	}
	rect, ok := nativeWindowGeometry()
	if !ok {
		return
	}
	mon, ok := monitorForWindow(rect.x, rect.y, rect.w, rect.h, monitorLayout())
	if !ok {
		return
	}
	chromeMu.Lock()
	area, ok := wslgHostWorkareaLocked(mon)
	chromeMu.Unlock()
	if !ok {
		return
	}
	x := max(area.x, min(rect.x, area.x+area.w-rect.w))
	y := max(area.y, min(rect.y, area.y+area.h-rect.h))
	if x != rect.x || y != rect.y {
		setWindowRect(x, y, rect.w, rect.h)
		windowTracef("startup fitting default titlebar window monitor=%q target=(%d,%d %dx%d)", mon.name, x, y, rect.w, rect.h)
	}
	s.titlebar.starting = false
}

func usesTitlebar(o Options) bool {
	return runtime.GOOS != "js" && runtime.GOOS != "android" && runtime.GOOS != "ios" &&
		!o.NativeHeader && !o.CustomHeader && !o.Borderless && !o.Transparent &&
		!o.MousePassthrough && o.Interactive == nil
}

func (s *shell) titlebarInset() int {
	if !s.titlebar.enabled || s.fullscreenNow() {
		return 0
	}
	if s.titlebar.style.height > 0 {
		return s.titlebar.style.height
	}
	return titlebarHeight
}

func titlebarButton(x, width int) int {
	if x < max(0, width-3*titlebarButtonWidth) || x >= width {
		return 0
	}
	return (x-(width-3*titlebarButtonWidth))/titlebarButtonWidth + 1
}

func (s *shell) titlebarPointer(x, y int, down bool) bool {
	if s.titlebarInset() == 0 {
		return false
	}
	width, _ := ebiten.WindowSize()
	hit := 0
	if y < 0 && y >= -s.titlebarInset() {
		hit = titlebarButton(x, width)
	}
	pressed := s.titlebar.pressed
	if down && !s.titlebar.down {
		s.titlebar.pressed = hit
	}
	s.titlebar.down = down
	if !down && pressed != 0 {
		s.titlebar.pressed = 0
		if hit == pressed {
			switch hit {
			case 1:
				Minimize()
			case buttonMaximize:
				if !s.titlebar.fixed {
					ToggleMaximize()
				}
			case buttonClose:
				RequestClose()
			}
		}
		return true
	}
	return hit != 0 || s.titlebar.pressed != 0
}

func (s *shell) titlebarCanvas(screen *ebiten.Image) (*ebiten.Image, func()) {
	inset := s.titlebarInset()
	if inset == 0 {
		return screen, func() {}
	}
	w, h := screen.Bounds().Dx(), max(1, screen.Bounds().Dy()-inset)
	canvas := s.titlebar.canvas
	if canvas == nil || canvas.Bounds().Dx() != w || canvas.Bounds().Dy() != h {
		if canvas != nil {
			canvas.Deallocate()
		}
		canvas = ebiten.NewImage(w, h)
		s.titlebar.canvas = canvas
	}
	return canvas, func() {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(0, float64(inset))
		screen.DrawImage(canvas, &op)
		s.drawTitlebar(screen)
	}
}

func (s *shell) drawTitlebar(screen *ebiten.Image) {
	width := screen.Bounds().Dx()
	// The caption colors are RGBA channel tables; one const per channel would hide the color.
	var bg color.Color = color.NRGBA{R: 235, G: 231, B: 218, A: 255} //nolint:mnd // RGBA palette table
	var ink color.Color = color.NRGBA{R: 31, G: 41, B: 55, A: 255}   //nolint:mnd // RGBA palette table
	page := s.pageBackground()
	r, g, b, _ := page.RGBA()
	if r+g+b < 3*0x8000 {
		bg = color.NRGBA{R: 35, G: 38, B: 44, A: 255}     //nolint:mnd // RGBA palette table
		ink = color.NRGBA{R: 235, G: 237, B: 240, A: 255} //nolint:mnd // RGBA palette table
	}
	style := s.titlebar.style
	if style.background != nil {
		bg = style.background
	}
	if style.foreground != nil {
		ink = style.foreground
	}
	height := float32(s.titlebarInset())
	vector.FillRect(screen, 0, 0, float32(width), height, bg, false)
	x, y := s.cursorPosition()
	for button := 1; button <= 3; button++ {
		left := width - (titlebarButtonCount+1-button)*titlebarButtonWidth
		if y < 0 && y >= -s.titlebarInset() && titlebarButton(x, width) == button {
			var hover color.Color = color.NRGBA{R: 128, G: 128, B: 128, A: 90} //nolint:mnd // RGBA palette table
			if style.hover != nil {
				hover = style.hover
			}
			if button == buttonClose {
				hover = color.NRGBA{R: 210, G: 40, B: 40, A: 255} //nolint:mnd // RGBA palette table
				if style.closeHover != nil {
					hover = style.closeHover
				}
			}
			vector.FillRect(screen, float32(left), 0, titlebarButtonWidth, height, hover, false)
		}
		cx := float32(left) + titlebarButtonWidth/halfDivisor
		cy := height / halfDivisor
		switch button {
		case 1:
			vector.StrokeLine(screen, cx-5, cy+3, cx+5, cy+3, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		case buttonMaximize:
			if Maximized() {
				vector.StrokeRect(screen, cx-3, cy-5, 8, 8, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
			}
			vector.StrokeRect(screen, cx-5, cy-3, 8, 8, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		case buttonClose:
			vector.StrokeLine(screen, cx-4, cy-4, cx+4, cy+4, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
			vector.StrokeLine(screen, cx-4, cy+4, cx+4, cy-4, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		}
	}
	title := s.app.Title()
	if style.title != "" {
		title = style.title
	}
	available := max(0, width-3*titlebarButtonWidth-titleEndReserve)
	if title != s.titlebar.title || available != s.titlebar.titleWidth {
		s.titlebar.title = title
		s.titlebar.titleWidth = available
		s.titlebar.fittedTitle = fitHeaderTitle(title, style.font, available)
	}
	metrics := style.font.Metrics()
	var op text.DrawOptions
	op.GeoM.Translate(titleInsetX, (float64(height)-metrics.HAscent-metrics.HDescent)/halfDivisor)
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(screen, s.titlebar.fittedTitle, style.font, &op)
}
