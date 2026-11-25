//go:build linux && !android

package window

import (
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

const drmSysfsRoot = "/sys/class/drm"

// detectMonitorHz queries the refresh rate (Hz) of the Linux display without
// cgo or background X11 connection goroutines. It reads connected DRM sysfs
// mode lines when available and falls back to defaultMonitorHz (60).
func detectMonitorHz(_, _, _, _ int) int {
	if hz, ok := detectMonitorHzDRM(drmSysfsRoot); ok {
		return sanitizeMonitorHz(hz)
	}
	return defaultMonitorHz
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
