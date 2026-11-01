package window

import (
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
	if isWSL() {
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
	if dragging || time.Now().After(pendingWindowedUntil) {
		pendingWindowedRestore = false
		return
	}
	x, y := windowPosition()
	w, h := ebiten.WindowSize()
	if restoreRectMatches(workarea{x: x, y: y, w: w, h: h}, workarea{x: pendingWindowedX, y: pendingWindowedY, w: pendingWindowedW, h: pendingWindowedH}) {
		pendingWindowedRestore = false
		return
	}
	setWindowRect(pendingWindowedX, pendingWindowedY, pendingWindowedW, pendingWindowedH)
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
	defer func() { wasExternalMaximized = maxed }()

	if !isWSL() || manualMax.active || manualMax.learning || manualMax.restoring {
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

	if isWSL() && ebiten.IsWindowMaximized() && !manualMax.active {
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
