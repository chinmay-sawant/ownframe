package window

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// placementVersion is the on-disk schema for the remembered rect.
	placementVersion = 1
	// placementMaxIDLen bounds a PlacementID so it stays a file stem.
	placementMaxIDLen = 64
	// placementMaxDimension rejects absurd rects without clamping real ones.
	placementMaxDimension = 16384
	// placementSaveInterval debounces geometry writes off the update path.
	placementSaveInterval = time.Second
	// placementDirMode and placementFileMode keep a window rect owner-only.
	placementDirMode  = 0o700
	placementFileMode = 0o600
)

// windowPlacement is the rect an app restores on its next launch: where the
// windowed frame was, whether it closed maximized, and the maximized rect
// plus the monitor name that owned it. Coordinates are global virtual-screen
// pixels, so the origin already selects the monitor.
type windowPlacement struct {
	Version   int    `json:"version"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Maximized bool   `json:"maximized"`
	MaxX      int    `json:"maxX"`
	MaxY      int    `json:"maxY"`
	MaxW      int    `json:"maxW"`
	MaxH      int    `json:"maxH"`
	Monitor   string `json:"monitor,omitempty"`
}

// validatePlacementID keeps a PlacementID a safe file stem: letters,
// digits, hyphen, and underscore only, so no separator or traversal.
func validatePlacementID(id string) error {
	if strings.TrimSpace(id) == "" || len(id) > placementMaxIDLen {
		return errors.New("window: bad placement id")
	}

	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}

		return errors.New("window: bad placement id")
	}

	return nil
}

// placementPath resolves the remembered rect for id under the user config
// directory. It creates nothing; the save makes the directory owner-only.
func placementPath(id string) (string, error) {
	if err := validatePlacementID(id); err != nil {
		return "", err
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "ownframe", "placement", id+".json"), nil
}

// validPlacementRect reports whether w and h describe a real window.
func validPlacementRect(w, h int) bool {
	return w > 0 && h > 0 && w <= placementMaxDimension && h <= placementMaxDimension
}

// clampPlacement pulls value up to low and, when high is positive, down to
// high, mirroring the window size limits runWindow installs.
func clampPlacement(value, low, high int) int {
	if value < low {
		return low
	}

	if high > 0 && value > high {
		return high
	}

	return value
}

// loadPlacement reads the remembered rect for id. Any failure (missing,
// corrupt, invalid) reports ok false so the caller falls back to defaults.
func loadPlacement(id string) (windowPlacement, bool) {
	var placement windowPlacement

	path, err := placementPath(id)
	if err != nil {
		return placement, false
	}

	data, err := os.ReadFile(path) //nolint:gosec // placementPath stays under the user config dir.
	if err != nil {
		return placement, false
	}

	if err := json.Unmarshal(data, &placement); err != nil {
		return placement, false
	}

	if placement.Version != placementVersion {
		return windowPlacement{}, false
	}

	if !validPlacementRect(placement.W, placement.H) {
		return windowPlacement{}, false
	}

	if placement.Maximized && !validPlacementRect(placement.MaxW, placement.MaxH) {
		return windowPlacement{}, false
	}

	return placement, true
}

// savePlacement writes the rect for id atomically through a temp file in the
// same directory. It is best-effort: callers ignore the error so a
// read-only config never breaks the window.
func savePlacement(id string, placement windowPlacement) error {
	path, err := placementPath(id)
	if err != nil {
		return err
	}

	placement.Version = placementVersion

	data, err := json.Marshal(placement)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), placementDirMode); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".placement-*.tmp")
	if err != nil {
		return err
	}

	name := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)

		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)

		return err
	}

	if err := os.Chmod(name, placementFileMode); err != nil {
		_ = os.Remove(name)

		return err
	}

	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)

		return err
	}

	return nil
}

// currentPlacement snapshots the live window: the windowed rect when idle,
// or the saved windowed rect plus the workarea when manually maximized. A
// WM-maximized window has no saved windowed rect, so the live rect doubles
// as both. Zero sizes (no desktop loop) are left for the caller to reject.
// The monitor is the one holding the window center from true origins, not
// Ebiten's straddle-prone report, so the JSON names the panel the window is
// actually on.
func currentPlacement() windowPlacement {
	placement := windowPlacement{
		Maximized: Maximized(),
	}

	x, y := windowPosition()
	w, h := ebiten.WindowSize()

	if manualMax.active {
		placement.X, placement.Y = manualMax.sx, manualMax.sy
		placement.W, placement.H = manualMax.sw, manualMax.sh
		placement.MaxX, placement.MaxY = manualMax.wx, manualMax.wy
		placement.MaxW, placement.MaxH = manualMax.ww, manualMax.wh
	} else {
		placement.X, placement.Y = x, y
		placement.W, placement.H = w, h
		placement.MaxX, placement.MaxY = x, y
		placement.MaxW, placement.MaxH = w, h
	}

	if layout := monitorLayout(); len(layout) != 0 {
		if m, ok := monitorForWindow(placement.X, placement.Y, placement.W, placement.H, layout); ok {
			placement.Monitor = m.name
			return placement
		}
		// Center outside every panel: the frame is mostly off-screen (like a
		// 160px strip). Keep the Ebiten name as a fallback below instead of
		// naming a panel the window barely touches.
	}

	if monitor := ebiten.Monitor(); monitor != nil {
		placement.Monitor = monitor.Name()
	}

	return placement
}

// placementMonitorKnown reports whether a monitor with name is present now.
// An empty name skips the check so older files without one still apply.
func placementMonitorKnown(name string) bool {
	if name == "" {
		return true
	}

	for _, monitor := range ebiten.AppendMonitors(nil) {
		if monitor.Name() == name {
			return true
		}
	}

	return false
}

// placementMonitorSize returns the size of the named monitor in
// device-independent pixels. It reports false when the monitor is gone or
// headless sizes are not available yet, so callers keep the old behavior.
func placementMonitorSize(name string) (int, int, bool) {
	for _, monitor := range ebiten.AppendMonitors(nil) {
		if monitor.Name() == name {
			w, h := monitor.Size()
			if w > 0 && h > 0 {
				return w, h, true
			}

			return 0, 0, false
		}
	}

	return 0, 0, false
}

// liveMonitorResolution reports the size of the monitor Ebiten thinks owns
// the window right now. It reports false when headless or before the window
// exists, so callers keep the old behavior. Check this resolution before
// maximizing: a rect learned on the 1920 panel must never expand onto the
// 2560 panel, and a 2560 workarea must never land on the 1920 panel.
func liveMonitorResolution() (int, int, bool) {
	if isWSLg() {
		if x, y, ok := nativeWindowPosition(); ok {
			w, h := ebiten.WindowSize()
			if m, ok := monitorForWindow(x, y, w, h, monitorLayout()); ok {
				return m.w, m.h, true
			}
		}
	}

	monitor := ebiten.Monitor()
	if monitor == nil {
		return 0, 0, false
	}
	w, h := monitor.Size()
	if w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

// placementPositionVisible reports whether an x,y window origin keeps a
// grabbable strip inside the true union. Origins come from xrandr when
// available, so inverted order (2560 at +1920, 1920 at +0) validates
// against the panels the window actually overlaps. Unknown bounds keep the
// old behavior and report true.
func placementPositionVisible(x, y, w, h int) bool {
	if layout := monitorLayout(); len(layout) != 0 {
		if minX, minY, maxX, maxY, ok := virtualUnion(layout); ok {
			nx, ny := clampDragTargetInUnion(x, y, w, h, minX, minY, maxX, maxY)
			return nx == x && ny == y
		}
	}
	bw, bh := dragOuterBound()
	return positionVisibleInBound(x, y, w, h, bw, bh)
}

// placementSubstantiallyVisible reports whether at least half the window
// area overlaps the true union. Dragging allows a 160px grabbable strip,
// but closing on that strip must not overwrite a good rect: the next launch
// would restore a nearly invisible frame.
func placementSubstantiallyVisible(x, y, w, h int) bool {
	return substantiallyVisibleInLayout(x, y, w, h, monitorLayout())
}

// placementOriginVisible reports whether the window origin (the custom
// header with minimize, maximize, and close) sits inside the true union.
// A 1120x760 frame at y=149 overlaps the 1920 panel by area, yet its top
// 68px header sits in the dead zone above y=320 and no double-tap can
// reach it. Saves and restores must reject that rect.
func placementOriginVisible(x, y int) bool {
	return originInLayout(x, y, monitorLayout())
}

// originInLayout is the pure rule behind placementOriginVisible.
func originInLayout(x, y int, layout []monitorRect) bool {
	if len(layout) == 0 {
		return true
	}
	for _, m := range layout {
		if x >= m.x && x < m.x+m.w && y >= m.y && y < m.y+m.h {
			return true
		}
	}
	return false
}

// placementTargetMonitor picks the panel a fresh or rejected rect belongs
// on: the saved monitor when it fits, else the smallest panel that fits
// the window, else the largest panel. Top-left of that panel is the
// explicit position, so launches start at a monitor top-left instead of a
// cascade in the dead zone.
func placementTargetMonitor(saved string, w, h int) (monitorRect, bool) {
	return targetMonitorInLayout(monitorLayout(), saved, w, h)
}

func targetMonitorInLayout(layout []monitorRect, saved string, w, h int) (monitorRect, bool) {
	if len(layout) == 0 {
		return monitorRect{}, false
	}
	for _, m := range layout {
		if m.name == saved && w <= m.w && h <= m.h {
			return m, true
		}
	}
	best := monitorRect{}
	bestArea := 0
	found := false
	for _, m := range layout {
		if w <= m.w && h <= m.h {
			area := m.w * m.h
			if !found || area < bestArea {
				best, bestArea, found = m, area, true
			}
		}
	}
	if found {
		return best, true
	}
	for _, m := range layout {
		area := m.w * m.h
		if !found || area > bestArea {
			best, bestArea, found = m, area, true
		}
	}
	return best, found
}

// substantiallyVisibleInLayout is the pure rule behind the save guard.
func substantiallyVisibleInLayout(x, y, w, h int, layout []monitorRect) bool {
	if len(layout) == 0 {
		return true
	}
	visible := 0
	for _, m := range layout {
		visible += overlapArea(x, y, w, h, m)
	}
	if w <= 0 || h <= 0 {
		return false
	}
	return visible*2 >= w*h
}

// positionVisibleInBound is the pure rule behind placementPositionVisible:
// the origin survives clampDragTarget in the given bound. Unknown bounds
// keep the old behavior and report true so a missing probe never discards
// a good rect. Split out for deterministic tests without a display.
func positionVisibleInBound(x, y, w, h, bw, bh int) bool {
	if bw <= 0 || bh <= 0 {
		return true
	}
	nx, ny := clampDragTarget(x, y, w, h, bw, bh)
	return nx == x && ny == y
}

// placementFits reports whether a w by h rect fits in the named monitor.
// An unknown monitor or unknown size keeps the old behavior and reports
// true so a missing probe never discards a good rect.
func placementFits(name string, w, h int) bool {
	monW, monH, ok := placementMonitorSize(name)
	if !ok {
		return true
	}

	return w <= monW && h <= monH
}

// placeExplicit centers a w by h frame on its target panel, so a
// fresh or rejected rect starts inside that panel instead of cascading
// into the dead zone above the 1920 panel or off to the right. A size set
// before RunGame is ignored on WSLg, so it also arms a post-map retry that
// repeats the position until the presented rect lands.
func placeExplicit(saved string, w, h int) (int, int) {
	if isWSLg() && saved == "" {
		saved = "XWAYLAND0"
	}
	if target, ok := placementTargetMonitor(saved, w, h); ok {
		ebiten.SetWindowSize(w, h)
		x, y := target.x+max(0, (target.w-w)/monitorCenterDivisor), target.y+max(0, (target.h-h)/monitorCenterDivisor)
		ebiten.SetWindowPosition(x, y)
		rememberWindowedRestore(x, y, w, h)
		return x, y
	}
	ebiten.SetWindowSize(w, h)
	return 0, 0
}

// applyPlacement installs a remembered rect before the game loop starts.
// defW and defH are the app's default size; minW, minH, maxW, and maxH clamp
// the restored size the way runWindow clamps the first frame. It returns the
// placement it installed so the periodic sync does not rewrite it.
func applyPlacement(id string, defW, defH, minW, minH, maxW, maxH int) windowPlacement {
	fallback := windowPlacement{X: 0, Y: 0, W: defW, H: defH}

	placement, ok := loadPlacement(id)
	if !ok {
		w, h := clampPlacement(defW, minW, maxW), clampPlacement(defH, minH, maxH)
		x, y := placeExplicit("", w, h)
		fallback.X, fallback.Y, fallback.W, fallback.H = x, y, w, h
		return fallback
	}

	if placement.Maximized && isWSL() {
		if isWSLg() {
			for _, m := range monitorLayout() {
				if m.name == placement.Monitor {
					if area, ok := wslgHostWorkarea(m); ok {
						placement.MaxX, placement.MaxY, placement.MaxW, placement.MaxH = area.x, area.y, area.w, area.h
					}
				}
				area := workarea{x: placement.MaxX, y: placement.MaxY, w: placement.MaxW, h: placement.MaxH}
				if m.name == placement.Monitor && areaMatchesMonitorSize(m.w, m.h, area) {
					placement.MaxX = m.x
					if placement.MaxY < m.y || placement.MaxY > m.y+workareaGapMax {
						placement.MaxY = m.y
					}
					break
				}
			}
		}
		known := placementMonitorKnown(placement.Monitor)
		fits := placementFits(placement.Monitor, placement.MaxW, placement.MaxH)
		applyNow, retry := wslMaximizePlan(known, fits)

		if !applyNow && !retry {
			// A maximized rect learned on one monitor must fit the monitor
			// that owns the file. A 2560-wide workarea restored onto a
			// 1920-wide panel overflows it and strands the window across
			// monitors, so fall back to the windowed size at that panel
			// top-left and let the user re-maximize to re-learn.
			manualMax.active = false
			manualMax.known = false
			manualMax.learning = false
			manualMax.restoring = false
			w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
			if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
				w, h = min(w, monW), min(h, monH)
				w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
			}
			x, y := placeExplicit(placement.Monitor, w, h)
			fallback.X, fallback.Y, fallback.W, fallback.H = x, y, w, h

			return fallback
		}

		// Check the resolution before expanding: the saved maximized rect
		// must fill its claimed monitor. A small rect is not a workarea, it
		// is a stale windowed size under a maximized flag.
		if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
			if !areaMatchesMonitorSize(monW, monH, workarea{w: placement.MaxW, h: placement.MaxH}) {
				manualMax.active = false
				manualMax.known = false
				manualMax.learning = false
				manualMax.restoring = false
				w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
				w, h = min(w, monW), min(h, monH)
				w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
				x, y := placeExplicit(placement.Monitor, w, h)
				fallback.X, fallback.Y, fallback.W, fallback.H = x, y, w, h

				return fallback
			}
		}

		// When the live monitor is already known, the expanded resolution
		// must match it. A 1920 workarea applied on the 2560 panel (or the
		// reverse) lands the window on the wrong screen and strands it to
		// the right, so fall back and re-learn on the live monitor.
		if liveW, liveH, ok := liveMonitorResolution(); ok && !isWSLg() {
			if !areaMatchesMonitorSize(liveW, liveH, workarea{w: placement.MaxW, h: placement.MaxH}) {
				manualMax.active = false
				manualMax.known = false
				manualMax.learning = false
				manualMax.restoring = false
				w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
				w, h = min(w, liveW), min(h, liveH)
				w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
				x, y := placeExplicit(placement.Monitor, w, h)
				fallback.X, fallback.Y, fallback.W, fallback.H = x, y, w, h

				return fallback
			}
		}

		rememberManualMaximize(placement)
		if retry {
			pendingManualRestore = true
			pendingManualUntil = time.Time{}
		}
		if !applyNow {
			// The monitor list is still empty. Keep the windowed size for
			// this frame; retryManualRestore installs the workarea once
			// the saved monitor is visible.
			w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
			ebiten.SetWindowSize(w, h)

			return placement
		}

		if !placementFits(placement.Monitor, placement.W, placement.H) {
			if target, ok := placementTargetMonitor(placement.Monitor, min(placement.W, defW), min(placement.H, defH)); ok {
				placement.X, placement.Y = target.x, target.y
			} else {
				placement.X, placement.Y = 0, 0
			}
			placement.W = min(placement.W, defW)
			placement.H = min(placement.H, defH)
			if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
				placement.W = min(placement.W, monW)
				placement.H = min(placement.H, monH)
			}
			placement.W = clampPlacement(placement.W, minW, maxW)
			placement.H = clampPlacement(placement.H, minH, maxH)
			manualMax.sx, manualMax.sy = placement.X, placement.Y
			manualMax.sw, manualMax.sh = placement.W, placement.H
		}

		// A size set before RunGame is ignored. retryManualRestore repeats
		// it once the window exists, until the presented size matches.
		ebiten.SetWindowPosition(placement.MaxX, placement.MaxY)
		ebiten.SetWindowSize(placement.MaxW, placement.MaxH)

		return placement
	}

	if !placementMonitorKnown(placement.Monitor) {
		// The owning monitor is gone; keep the size and start at a live
		// panel top-left instead of cascading into the dead zone.
		w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
		x, y := placeExplicit(placement.Monitor, w, h)
		fallback.X, fallback.Y, fallback.W, fallback.H = x, y, w, h

		return fallback
	}

	if placement.Maximized {
		w, h := clampPlacement(placement.MaxW, minW, maxW), clampPlacement(placement.MaxH, minH, maxH)
		if !placementFits(placement.Monitor, w, h) {
			// The maximized rect belongs to a bigger monitor. Open at
			// the default size and still maximize, so the window manager
			// sizes to the live monitor instead of flashing an oversized
			// frame that spans monitors.
			w, h = defW, defH
		}
		ebiten.SetWindowSize(w, h)
		// MaximizeWindow before the run loop never takes effect, so the
		// resize path applies it on the first update once the window exists.
		pendingStartMaximize = true

		return placement
	}

	w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)

	if !placementFits(placement.Monitor, w, h) {
		// The saved windowed size belongs to a bigger monitor. Clamp it
		// to the owning monitor and start at that panel top-left instead
		// of replaying a rect that spans monitors.
		if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
			w, h = min(w, monW), min(h, monH)
			w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
		}
		x, y := placeExplicit(placement.Monitor, w, h)
		placement.X, placement.Y, placement.W, placement.H = x, y, w, h
		ebiten.SetWindowSize(w, h)
		ebiten.SetWindowPosition(x, y)

		return placement
	}
	if !placementPositionVisible(placement.X, placement.Y, w, h) || !placementSubstantiallyVisible(placement.X, placement.Y, w, h) || !placementOriginVisible(placement.X, placement.Y) {
		// The saved origin belongs to another layout, is a drag-clamp
		// strip, or sits in the dead zone above the 1920 panel where the
		// header is unreachable. Start at the target panel top-left.
		x, y := placeExplicit(placement.Monitor, w, h)
		placement.X, placement.Y, placement.W, placement.H = x, y, w, h
		ebiten.SetWindowSize(w, h)
		ebiten.SetWindowPosition(x, y)

		return placement
	}
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowPosition(placement.X, placement.Y)
	if isWSLg() {
		rememberWindowedRestore(placement.X, placement.Y, w, h)
	}
	placement.W, placement.H = w, h

	return placement
}

// wslMaximizePlan decides how a saved WSL maximize is installed.
// An unknown monitor waits: the list is empty before the window exists.
// A rect that fits is applied now and again after RunGame, because the
// early call does not stick. A rect that does not fit is refused.
func wslMaximizePlan(monitorKnown, fits bool) (applyNow, retry bool) {
	if !monitorKnown {
		return false, true
	}

	if !fits {
		return false, false
	}

	return true, true
}

// rememberManualMaximize copies a saved maximize into the manual state so
// later syncs keep reporting it instead of the default 1500x900 frame.
func rememberManualMaximize(placement windowPlacement) {
	manualMax.sx, manualMax.sy = placement.X, placement.Y
	manualMax.sw, manualMax.sh = placement.W, placement.H
	manualMax.wx, manualMax.wy = placement.MaxX, placement.MaxY
	manualMax.ww, manualMax.wh = placement.MaxW, placement.MaxH
	manualMax.learnedMon = placement.Monitor
	manualMax.known = true
	manualMax.active = true
	manualMax.learning = false
	manualMax.restoring = false
	if manualMax.areas == nil {
		manualMax.areas = make(map[string]workarea)
	}
	if placement.Monitor != "" {
		manualMax.areas[placement.Monitor] = workarea{
			x: placement.MaxX,
			y: placement.MaxY,
			w: placement.MaxW,
			h: placement.MaxH,
		}
	}
}

// retryManualRestore installs the saved workarea once the window and the
// monitor list exist. It stops when the presented rect matches, the
// monitor cannot hold it, the deadline passes, or a drag has started.
func retryManualRestore(dragging bool) {
	if !pendingManualRestore {
		return
	}

	if pendingManualUntil.IsZero() {
		pendingManualUntil = time.Now().Add(manualMaxLearnTimeout)
	}
	if dragging || time.Now().After(pendingManualUntil) {
		pendingManualRestore = false

		return
	}

	if !placementMonitorKnown(manualMax.learnedMon) {
		return
	}

	if !placementFits(manualMax.learnedMon, manualMax.ww, manualMax.wh) {
		pendingManualRestore = false
		manualMax.active = false
		manualMax.known = false

		return
	}

	// Check the live resolution before expanding: the window exists now,
	// so the current monitor size is known. A workarea saved on the 1920
	// panel must not expand onto the 2560 panel, and a 2560 workarea must
	// not land on the 1920 panel. Drop the stale rect and let the user
	// re-maximize to learn the right workarea for this monitor.
	if liveW, liveH, ok := liveMonitorResolution(); ok && !isWSLg() {
		if !areaMatchesMonitorSize(liveW, liveH, workarea{w: manualMax.ww, h: manualMax.wh}) {
			pendingManualRestore = false
			manualMax.active = false
			manualMax.known = false

			return
		}
	}

	x, y := windowPosition()
	w, h := ebiten.WindowSize()
	landed := restoreRectMatches(workarea{x: x, y: y, w: w, h: h}, workarea{x: manualMax.wx, y: manualMax.wy, w: manualMax.ww, h: manualMax.wh})
	if landed {
		pendingManualRestore = false
		manualMax.active = true

		return
	}

	setWindowRect(manualMax.wx, manualMax.wy, manualMax.ww, manualMax.wh)
	manualMax.active = true
}

// syncPlacement remembers the live rect at most once per interval while the
// geometry keeps changing. Empty ids and headless sizes never touch the disk.
// A fullscreen frame is not a windowed rect, so it never overwrites the good
// one behind it.
func (s *shell) syncPlacement() {
	if pendingManualRestore || pendingWindowedRestore {
		return
	}

	if s.placementID == "" {
		return
	}

	if s.fullscreenNow() {
		return
	}

	if manualMax.learning || manualMax.restoring {
		// The workarea dance is mid-flight; the live rect is transitional.
		return
	}

	now := time.Now()
	if !s.lastPlacementSave.IsZero() && now.Sub(s.lastPlacementSave) < placementSaveInterval {
		return
	}

	s.lastPlacementSave = now
	placement := currentPlacement()
	if !validPlacementRect(placement.W, placement.H) {
		return
	}

	if !placementSubstantiallyVisible(placement.X, placement.Y, placement.W, placement.H) || !placementOriginVisible(placement.X, placement.Y) {
		// The frame is a drag-clamp strip (like x=-960 keeping 160px on
		// the 1920 panel) or sits in the dead zone above it (like y=149
		// hiding the header). Do not overwrite the last good rect with it.
		return
	}

	if s.lastPlacementValid && placement == s.lastPlacement {
		return
	}

	if err := savePlacement(s.placementID, placement); err != nil {
		return
	}

	s.lastPlacement = placement
	s.lastPlacementValid = true
	s.lastPlacementSave = now
}

// savePlacementNow writes the live rect once, for window close. Invalid
// sizes (the loop never ran) skip the write. A workarea probe in flight
// writes a transitional rect, so it skips the same dance syncPlacement
// skips instead of persisting garbage over the good rect. A fullscreen close
// keeps the windowed rect behind it instead of persisting the fullscreen
// size as windowed.
func savePlacementNow(id string) {
	if strings.TrimSpace(id) == "" {
		return
	}

	if ebiten.IsFullscreen() {
		return
	}

	if manualMax.learning || manualMax.restoring {
		return
	}

	placement := currentPlacement()
	if !validPlacementRect(placement.W, placement.H) {
		return
	}

	if !placementSubstantiallyVisible(placement.X, placement.Y, placement.W, placement.H) || !placementOriginVisible(placement.X, placement.Y) {
		return
	}

	_ = savePlacement(id, placement)
}
