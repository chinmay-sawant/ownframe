//go:build linux && !android

package clipboard

import "os"

func writeOS(text string) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		wlWrite(text)

		return
	}

	if os.Getenv("DISPLAY") == "" {
		return
	}

	xWrite(text)
}

func readOS() (string, bool) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return wlRead()
	}

	if os.Getenv("DISPLAY") == "" {
		return "", false
	}

	return xRead()
}

func clearOS() {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		wlClear()

		return
	}

	if os.Getenv("DISPLAY") == "" {
		return
	}

	xWrite("")
}
