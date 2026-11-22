//go:build linux && !android

// This file is the wl-clipboard fallback. The pure-Go data-control client
// (wayland_datacontrol.go) runs first; compositors without ext/wlr data-control
// (GNOME Mutter, Weston/WSLg) land here. Core wl_data_device needs a focused
// surface and input serial owned by the windowing layer, so it is not used.
// Diagnose / ErrWaylandToolsMissing surface missing helpers explicitly.
package clipboard

import (
	"context"
	"os/exec"
)

func wlCopyArgs() []string {
	return []string{"--type", "text/plain", "--"}
}

func wlClearArgs() []string {
	return []string{"--clear"}
}

func wlPasteArgs() []string {
	return []string{"--no-newline"}
}

func wlWrite(ctx context.Context, text string) bool {
	return runWL(ctx, "wl-copy", wlCopyArgs(), text)
}

func wlClear() {
	runWL(context.Background(), "wl-copy", wlClearArgs(), "")
}

func wlRead(ctx context.Context) (string, bool) {
	return outWL(ctx, "wl-paste", wlPasteArgs())
}

func lookWL(name string) (string, bool) {
	path, err := exec.LookPath(name)

	return path, err == nil
}
