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
func currentPlacement() windowPlacement {
	placement := windowPlacement{
		Maximized: Maximized(),
	}

	x, y := ebiten.WindowPosition()
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

// applyPlacement installs a remembered rect before the game loop starts.
// defW and defH are the app's default size; minW, minH, maxW, and maxH clamp
// the restored size the way runWindow clamps the first frame. It returns the
// placement it installed so the periodic sync does not rewrite it.
func applyPlacement(id string, defW, defH, minW, minH, maxW, maxH int) windowPlacement {
	fallback := windowPlacement{X: 0, Y: 0, W: defW, H: defH}

	placement, ok := loadPlacement(id)
	if !ok {
		return fallback
	}

	if !placementMonitorKnown(placement.Monitor) {
		// The owning monitor is gone; keep the size, let the window
		// manager choose the position on a live monitor.
		w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
		ebiten.SetWindowSize(w, h)
		fallback.W, fallback.H = w, h

		return fallback
	}

	if placement.Maximized && isWSL() {
		// A maximized rect learned on one monitor must fit the monitor
		// that owns the file. A 2560-wide workarea restored onto a
		// 1920-wide panel overflows it and strands the window across
		// monitors, so fall back to the windowed size and let the user
		// re-maximize to re-learn the right workarea.
		if !placementFits(placement.Monitor, placement.MaxW, placement.MaxH) {
			manualMax.active = false
			manualMax.known = false
			manualMax.learning = false
			manualMax.restoring = false
			w, h := clampPlacement(placement.W, minW, maxW), clampPlacement(placement.H, minH, maxH)
			if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
				w, h = min(w, monW), min(h, monH)
				w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
			}
			ebiten.SetWindowSize(w, h)
			fallback.W, fallback.H = w, h

			return fallback
		}
		if !placementFits(placement.Monitor, placement.W, placement.H) {
			placement.X, placement.Y = 0, 0
			placement.W = min(placement.W, defW)
			placement.H = min(placement.H, defH)
			if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
				placement.W = min(placement.W, monW)
				placement.H = min(placement.H, monH)
			}
			placement.W = clampPlacement(placement.W, minW, maxW)
			placement.H = clampPlacement(placement.H, minH, maxH)
		}
		manualMax.sx, manualMax.sy = placement.X, placement.Y
		manualMax.sw, manualMax.sh = placement.W, placement.H
		manualMax.wx, manualMax.wy = placement.MaxX, placement.MaxY
		manualMax.ww, manualMax.wh = placement.MaxW, placement.MaxH
		manualMax.learnedMon = placement.Monitor
		manualMax.known = true
		if manualMax.areas == nil {
			manualMax.areas = make(map[string]workarea)
		}
		if placement.Monitor != "" {
			manualMax.areas[placement.Monitor] = workarea{x: placement.MaxX, y: placement.MaxY, w: placement.MaxW, h: placement.MaxH}
		}
		ebiten.SetWindowPosition(placement.MaxX, placement.MaxY)
		ebiten.SetWindowSize(placement.MaxW, placement.MaxH)
		manualMax.active = true

		return placement
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
		// to the owning monitor and let the window manager choose the
		// position instead of replaying a rect that spans monitors.
		if monW, monH, ok := placementMonitorSize(placement.Monitor); ok {
			w, h = min(w, monW), min(h, monH)
			w, h = clampPlacement(w, minW, maxW), clampPlacement(h, minH, maxH)
		}
		ebiten.SetWindowSize(w, h)
		placement.W, placement.H = w, h

		return placement
	}
	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowPosition(placement.X, placement.Y)
	placement.W, placement.H = w, h

	return placement
}

// syncPlacement remembers the live rect at most once per interval while the
// geometry keeps changing. Empty ids and headless sizes never touch the disk.
func (s *shell) syncPlacement() {
	if s.placementID == "" {
		return
	}

	if manualMax.learning || manualMax.restoring {
		// The workarea dance is mid-flight; the live rect is transitional.
		return
	}

	now := time.Now()
	if now.Sub(s.lastPlacementSave) < placementSaveInterval && s.lastPlacementValid {
		return
	}

	placement := currentPlacement()
	if !validPlacementRect(placement.W, placement.H) {
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
// skips instead of persisting garbage over the good rect.
func savePlacementNow(id string) {
	if strings.TrimSpace(id) == "" {
		return
	}

	if manualMax.learning || manualMax.restoring {
		return
	}

	placement := currentPlacement()
	if !validPlacementRect(placement.W, placement.H) {
		return
	}

	_ = savePlacement(id, placement)
}
