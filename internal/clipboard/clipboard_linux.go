//go:build linux && !android

package clipboard

import (
	"context"
	"os"
)

type clipHost string

const (
	hostNone    clipHost = "none"
	hostWayland clipHost = "wayland"
	hostX11     clipHost = "x11"
)

const backendWaylandNative = "wayland-data-control"

func writeOS(ctx context.Context, text string) bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" && wlNativeWrite(ctx, text) {
		return true
	}

	switch linuxHost() {
	case hostWayland:
		return wlWrite(ctx, text)
	case hostX11:
		return xWrite(text)
	default:
		return false
	}
}

func readOS(ctx context.Context) (string, bool) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if text, ok, handled := wlNativeRead(ctx); handled {
			return text, ok
		}
	}

	switch linuxHost() {
	case hostWayland:
		return wlRead(ctx)
	case hostX11:
		return xRead()
	default:
		return "", false
	}
}

func clearOS() {
	if os.Getenv("WAYLAND_DISPLAY") != "" && wlNativeClear() {
		return
	}

	switch linuxHost() {
	case hostWayland:
		wlClear()
	case hostX11:
		xWrite("")
	}
}

func linuxHost() clipHost {
	return routeLinux(os.Getenv("WAYLAND_DISPLAY"), os.Getenv("DISPLAY"), lookWL)
}

func routeLinux(wayland, display string, look func(string) (string, bool)) clipHost {
	if wayland != "" && haveWL(look) {
		return hostWayland
	}

	if display != "" {
		return hostX11
	}

	return hostNone
}

func haveWL(look func(string) (string, bool)) bool {
	_, copyOK := look("wl-copy")
	_, pasteOK := look("wl-paste")

	return copyOK && pasteOK
}

func diagnoseOS() Status {
	if os.Getenv("WAYLAND_DISPLAY") != "" && wlNativeAvailable() {
		return Status{Backend: backendWaylandNative, Available: true}
	}

	return diagnoseLinux(os.Getenv("WAYLAND_DISPLAY"), os.Getenv("DISPLAY"), lookWL)
}

func diagnoseLinux(wayland, display string, look func(string) (string, bool)) Status {
	switch routeLinux(wayland, display, look) {
	case hostWayland:
		return Status{Backend: string(hostWayland), Available: true}
	case hostX11:
		if wayland != "" {
			return Status{
				Backend:   string(hostX11),
				Available: true,
				Reason:    "WAYLAND_DISPLAY set without wl-copy/wl-paste on PATH; using DISPLAY",
			}
		}

		return Status{Backend: string(hostX11), Available: true}
	default:
		if wayland != "" {
			return Status{
				Backend:   string(hostNone),
				Available: false,
				Reason:    ErrWaylandToolsMissing.Error(),
			}
		}

		return Status{
			Backend:   string(hostNone),
			Available: false,
			Reason:    "neither WAYLAND_DISPLAY nor DISPLAY is set",
		}
	}
}
