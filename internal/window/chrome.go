package window

import (
	"os"
	"runtime"
	"strings"
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

// RequestClose asks the run loop to exit on its next frame. It does nothing
// when no desktop loop is running, so Serve, web, and mobile callers can
// share the same button wiring.
func RequestClose() {
	closeRequested.Store(true)
}

// takeCloseRequest consumes a pending close request.
func takeCloseRequest() bool {
	return closeRequested.Swap(false)
}

// Minimize minimizes the desktop window. It does nothing when no desktop
// run loop is active or the platform is not a desktop.
func Minimize() {
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
	return manualMax.active || ebiten.IsWindowMaximized()
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

// pendingStartMaximize defers a remembered WM maximize past window creation:
// MaximizeWindow before the run loop never takes effect, so the resize path
// applies it on the first update once the window exists. WSL never uses it;
// the manual path installs client geometry up front instead.
var pendingStartMaximize bool

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
	active     bool // manually maximized: geometry is client-set workarea
	known      bool // workarea learned for learnedMon
	learning   bool // a WM maximize is in flight to reveal the workarea
	restoring  bool // learned; waiting for the restore configure before applying
	until      time.Time
	sx, sy     int // saved windowed position
	sw, sh     int // saved windowed size
	wx, wy     int // learned workarea position
	ww, wh     int // learned workarea size
	learnedMon string
	areas      map[string]workarea
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

	return monW-area.w <= workareaGapMax && monH-area.h <= workareaGapMax
}

// currentMonitorName reports the live monitor's name, or "" when headless.
func currentMonitorName() string {
	if monitor := ebiten.Monitor(); monitor != nil {
		return monitor.Name()
	}

	return ""
}

// toggleManualMaximize flips between the saved windowed rect and the
// learned workarea. The first maximize per monitor learns the workarea with
// one window-manager maximize (see completeManualMaximize) and converts it
// to client geometry once the restore configure lands.
func toggleManualMaximize() {
	if manualMax.learning || manualMax.restoring {
		// A workarea probe is in flight; swallowing the extra click keeps
		// one maximize in the air instead of stacking two.
		return
	}

	if manualMax.active {
		if w, h := ebiten.WindowSize(); w != manualMax.ww || h != manualMax.wh {
			// The user moved or resized out from under the manual
			// maximize; drop it and start fresh below.
			manualMax.active = false
		} else {
			ebiten.SetWindowSize(manualMax.sw, manualMax.sh)
			ebiten.SetWindowPosition(manualMax.sx, manualMax.sy)
			manualMax.active = false

			return
		}
	}

	manualMax.sx, manualMax.sy = ebiten.WindowPosition()
	manualMax.sw, manualMax.sh = ebiten.WindowSize()
	current := currentMonitorName()

	if area, ok := manualMax.areas[current]; ok {
		// A stale entry from the other monitor teleports the window
		// across screens. Drop the entry and re-probe instead.
		if !workareaMatchesMonitor(current, area) {
			delete(manualMax.areas, current)
		} else {
			applyWorkarea(current, area)

			return
		}
	}

	if manualMax.known && manualMax.learnedMon == current {
		learned := workarea{x: manualMax.wx, y: manualMax.wy, w: manualMax.ww, h: manualMax.wh}
		// Same stale-entry guard as the cache above: a rect learned
		// under a wrong monitor claim must re-probe, not teleport.
		if workareaMatchesMonitor(current, learned) {
			applyWorkarea(current, learned)

			return
		}
	}

	manualMax.learning = true
	manualMax.until = time.Now().Add(manualMaxLearnTimeout)
	ebiten.MaximizeWindow()
}

// applyWorkarea installs a learned rect as plain client geometry and marks
// the manual maximize active. Callers must have saved the windowed rect.
func applyWorkarea(mon string, area workarea) {
	manualMax.wx, manualMax.wy = area.x, area.y
	manualMax.ww, manualMax.wh = area.w, area.h
	manualMax.learnedMon = mon
	manualMax.known = true
	if manualMax.areas == nil {
		manualMax.areas = make(map[string]workarea)
	}
	manualMax.areas[mon] = area
	ebiten.SetWindowPosition(area.x, area.y)
	ebiten.SetWindowSize(area.w, area.h)
	manualMax.active = true
}

// completeManualMaximize watches a workarea probe from the resize path. When
// the maximized configure arrives it records the workarea, leaves the
// maximized window state, and re-applies the learned rect as plain client
// geometry once the restore configure lands. A same-frame Restore followed
// by SetWindowSize loses the resize on WSLg, which is the double-click the
// probe reproduced, so the apply waits for the windowed size to return.
// Call it with the latest Layout size every update.
func completeManualMaximize(pendingW, pendingH int) {
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

		if pendingW == manualMax.sw && pendingH == manualMax.sh {
			return // still the windowed size; the configure is in flight
		}

		area := workarea{w: pendingW, h: pendingH}
		area.x, area.y = ebiten.WindowPosition()
		current := currentMonitorName()
		manualMax.wx, manualMax.wy = area.x, area.y
		manualMax.ww, manualMax.wh = area.w, area.h
		manualMax.learnedMon = current
		manualMax.known = true
		// ebiten.Monitor is unreliable while a window straddles monitors,
		// so a probe learned on one panel can claim a name from the other.
		// Cache only a rect that plausibly fills its claimant; the dance
		// below still applies the learned geometry, so the current
		// maximize is fine and the next one simply re-probes.
		if workareaMatchesMonitor(current, area) {
			if manualMax.areas == nil {
				manualMax.areas = make(map[string]workarea)
			}
			manualMax.areas[current] = area
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
	ebiten.SetWindowPosition(manualMax.wx, manualMax.wy)
	ebiten.SetWindowSize(manualMax.ww, manualMax.wh)
	manualMax.active = true
}
