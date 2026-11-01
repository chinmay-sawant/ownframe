package window

import "os"

func isWSLg() bool {
	return isWSL() && os.Getenv("WAYLAND_DISPLAY") != ""
}

// WSLg shifts XRandR's desktop to nonnegative X11 coordinates, while window
// positions are relative to the Windows primary display, XWAYLAND0.
func wslgMonitorLayout(layout []monitorRect) []monitorRect {
	for _, primary := range layout {
		if primary.name != "XWAYLAND0" {
			continue
		}
		out := append([]monitorRect(nil), layout...)
		for i := range out {
			out[i].x -= primary.x
			out[i].y -= primary.y
		}
		return out
	}
	return layout
}
