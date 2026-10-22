//go:build linux && !android

package clipboard

import "os"

func writeOS(text string) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		// Wayland data-device is not implemented.
		return
	}

	if os.Getenv("DISPLAY") == "" {
		return
	}

	xWrite(text)
}

func readOS() (string, bool) {
	if os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("DISPLAY") == "" {
		return "", false
	}

	return xRead()
}

func clearOS() {
	if os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("DISPLAY") == "" {
		return
	}

	xWrite("")
}
