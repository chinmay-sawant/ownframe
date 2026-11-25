//go:build linux && !android

package window

import (
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
)

const drmSysfsRoot = "/sys/class/drm"

// detectMonitorHz queries the refresh rate (Hz) of the Linux display hosting
// the window center without cgo. It prefers X11 RandR (covering X11 and
// XWayland), falls back to reading connected DRM sysfs mode lines when RandR
// is unavailable, and finally returns defaultMonitorHz (60).
func detectMonitorHz(posX, posY, winW, winH int) int {
	if hz, ok := detectMonitorHzRandR(posX, posY, winW, winH); ok {
		return sanitizeMonitorHz(hz)
	}
	if hz, ok := detectMonitorHzDRM(drmSysfsRoot); ok {
		return sanitizeMonitorHz(hz)
	}
	return defaultMonitorHz
}

func detectMonitorHzRandR(posX, posY, winW, winH int) (int, bool) {
	if os.Getenv("DISPLAY") == "" {
		return 0, false
	}

	conn, err := xgb.NewConn()
	if err != nil {
		return 0, false
	}
	defer conn.Close()

	if err := randr.Init(conn); err != nil {
		return 0, false
	}

	setup := xproto.Setup(conn)
	if setup == nil || len(setup.Roots) == 0 {
		return 0, false
	}
	root := setup.DefaultScreen(conn).Root

	res, err := randr.GetScreenResourcesCurrent(conn, root).Reply()
	if err != nil || res == nil {
		return 0, false
	}

	const half = 2
	cx := posX + winW/half
	cy := posY + winH/half

	fallbackHz := 0
	for _, crtc := range res.Crtcs {
		info, err := randr.GetCrtcInfo(conn, crtc, res.ConfigTimestamp).Reply()
		if err != nil || info == nil || info.Mode == 0 || info.Width == 0 || info.Height == 0 {
			continue
		}

		modeInfo, ok := findRandRMode(res.Modes, uint32(info.Mode))
		if !ok {
			continue
		}

		hz := randrModeHz(modeInfo.DotClock, modeInfo.Htotal, modeInfo.Vtotal, modeInfo.ModeFlags)
		if hz < minValidMonitorHz || hz > maxValidMonitorHz {
			continue
		}

		if pointInRect(cx, cy, int(info.X), int(info.Y), int(info.Width), int(info.Height)) {
			return hz, true
		}
		if fallbackHz == 0 {
			fallbackHz = hz
		}
	}

	if fallbackHz > 0 {
		return fallbackHz, true
	}
	return 0, false
}

func findRandRMode(modes []randr.ModeInfo, id uint32) (randr.ModeInfo, bool) {
	for _, m := range modes {
		if m.Id == id {
			return m, true
		}
	}
	return randr.ModeInfo{}, false
}

func pointInRect(px, py, rx, ry, rw, rh int) bool {
	return px >= rx && px < rx+rw && py >= ry && py < ry+rh
}

// randrModeHz computes the integer refresh rate (Hz) from XRandR ModeInfo
// timing fields. Returns 0 when timing values are zero or invalid.
func randrModeHz(dotClock uint32, htotal, vtotal uint16, flags uint32) int {
	if dotClock == 0 || htotal == 0 || vtotal == 0 {
		return 0
	}

	const scanFactor = 2
	vt := uint64(vtotal)
	if flags&randr.ModeFlagDoubleScan != 0 {
		vt *= scanFactor
	}
	if flags&randr.ModeFlagInterlace != 0 {
		vt /= scanFactor
	}
	denom := uint64(htotal) * vt
	if denom == 0 {
		return 0
	}
	return roundMonitorHz(float64(dotClock) / float64(denom))
}

// detectMonitorHzDRM scans connected DRM connector directories under rootPath
// (normally /sys/class/drm) for a preferred mode line carrying an explicit
// "@Hz" refresh rate.
func detectMonitorHzDRM(rootPath string) (int, bool) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return 0, false
	}
	defer func() { _ = root.Close() }()

	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return 0, false
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || !strings.Contains(name, "-") {
			continue
		}
		status, err := root.ReadFile(path.Join(name, "status"))
		if err != nil || strings.TrimSpace(string(status)) != "connected" {
			continue
		}
		modes, err := root.ReadFile(path.Join(name, "modes"))
		if err != nil {
			continue
		}
		if hz, ok := parseDRMModesHz(string(modes)); ok {
			return hz, true
		}
	}
	return 0, false
}

// parseDRMModesHz returns the first valid refresh rate encoded in DRM modes
// text (one mode per line, e.g. "2560x1440@144" or "1920x1080@60.00").
func parseDRMModesHz(modesText string) (int, bool) {
	for line := range strings.SplitSeq(modesText, "\n") {
		if hz, ok := parseDRMModeLineHz(line); ok {
			return hz, true
		}
	}
	return 0, false
}

// parseDRMModeLineHz parses an explicit "@Hz" suffix from a DRM mode line.
// Plain "WxH" or "WxHp" lines without a refresh rate return (0, false).
func parseDRMModeLineHz(line string) (int, bool) {
	line = strings.TrimSpace(line)
	_, ratePart, ok := strings.Cut(line, "@")
	if !ok {
		return 0, false
	}
	ratePart = strings.TrimRight(strings.TrimSpace(ratePart), "ipIP")
	if ratePart == "" {
		return 0, false
	}
	rate, err := strconv.ParseFloat(ratePart, 64)
	if err != nil {
		return 0, false
	}
	hz := roundMonitorHz(rate)
	if hz < minValidMonitorHz || hz > maxValidMonitorHz {
		return 0, false
	}
	return hz, true
}
